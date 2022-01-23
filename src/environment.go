package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/distatus/battery"
	process "github.com/shirou/gopsutil/v3/process"
)

const (
	unknown         = "unknown"
	windowsPlatform = "windows"
	darwinPlatform  = "darwin"
	linuxPlatform   = "linux"
)

type commandError struct {
	err      string
	exitCode int
}

func (e *commandError) Error() string {
	return e.err
}

type noBatteryError struct{}

func (m *noBatteryError) Error() string {
	return "no battery"
}

type FileInfo struct {
	ParentFolder string
	Path         string
	IsDir        bool
}

type Cache interface {
	Init(home string)
	Close()
	Get(key string) (string, bool)
	// ttl in minutes
	Set(key, value string, ttl int)
}

type HTTPRequestModifier func(request *http.Request)

type windowsRegistryValueType int

const (
	regQword windowsRegistryValueType = iota
	regDword
	regString
)

type WindowsRegistryValue struct {
	valueType windowsRegistryValueType
	qword     uint64
	dword     uint32
	str       string
}

type WifiType string

type WifiInfo struct {
	SSID           string
	Interface      string
	RadioType      WifiType
	PhysType       WifiType
	Authentication WifiType
	Cipher         WifiType
	Channel        int
	ReceiveRate    int
	TransmitRate   int
	Signal         int
	Error          string
}

type Environment interface {
	Getenv(key string) string
	Pwd() string
	Home() string
	User() string
	Root() bool
	Host() (string, error)
	GOOS() string
	Shell() string
	Platform() string
	ErrorCode() int
	PathSeperator() string
	HasFiles(pattern string) bool
	HasFilesInDir(dir, pattern string) bool
	HasFolder(folder string) bool
	HasParentFilePath(path string) (fileInfo *FileInfo, err error)
	HasCommand(command string) bool
	FileContent(file string) string
	FolderList(path string) []string
	RunCommand(command string, args ...string) (string, error)
	RunShellCommand(shell, command string) string
	ExecutionTime() float64
	Args() *Args
	BatteryInfo() ([]*battery.Battery, error)
	WindowTitle(imageName, windowTitleRegex string) (string, error)
	WindowsRegistryKeyValue(path string) (*WindowsRegistryValue, error)
	HTTPRequest(url string, timeout int, requestModifiers ...HTTPRequestModifier) ([]byte, error)
	IsWsl() bool
	IsWsl2() bool
	StackCount() int
	TerminalWidth() (int, error)
	CachePath() string
	Cache() Cache
	Close()
	Logs() string
	InWSLSharedDrive() bool
	ConvertToLinuxPath(path string) string
	ConvertToWindowsPath(path string) string
	WifiNetwork() (*WifiInfo, error)
	TemplateCache() *TemplateCache
}

type commandCache struct {
	commands *concurrentMap
}

func (c *commandCache) set(command, path string) {
	c.commands.set(command, path)
}

func (c *commandCache) get(command string) (string, bool) {
	cmd, found := c.commands.get(command)
	if !found {
		return "", false
	}
	command, ok := cmd.(string)
	return command, ok
}

type logType string

const (
	Error logType = "error"
	Debug logType = "debug"
)

type environment struct {
	args       *Args
	cwd        string
	cmdCache   *commandCache
	fileCache  *fileCache
	tmplCache  *TemplateCache
	logBuilder strings.Builder
	debug      bool
}

func (env *environment) init(args *Args) {
	env.args = args
	env.fileCache = &fileCache{}
	env.fileCache.Init(env.CachePath())
	env.resolveConfigPath()
	env.cmdCache = &commandCache{
		commands: newConcurrentMap(),
	}
	if env.args != nil && *env.args.Debug {
		env.debug = true
		log.SetOutput(&env.logBuilder)
	}
}

func (env *environment) resolveConfigPath() {
	if env.args == nil || env.args.Config == nil || len(*env.args.Config) == 0 {
		return
	}
	// Cygwin path always needs the full path as we're on Windows but not really.
	// Doing filepath actions will convert it to a Windows path and break the init script.
	if env.Platform() == windowsPlatform && env.Shell() == bash {
		return
	}
	configFile := *env.args.Config
	if strings.HasPrefix(configFile, "~") {
		configFile = strings.TrimPrefix(configFile, "~")
		configFile = filepath.Join(env.Home(), configFile)
	}
	if !filepath.IsAbs(configFile) {
		if absConfigFile, err := filepath.Abs(configFile); err == nil {
			configFile = absConfigFile
		}
	}
	*env.args.Config = filepath.Clean(configFile)
}

func (env *environment) trace(start time.Time, function string, args ...string) {
	if !env.debug {
		return
	}
	elapsed := time.Since(start)
	trace := fmt.Sprintf("%s duration: %s, args: %s", function, elapsed, strings.Trim(fmt.Sprint(args), "[]"))
	log.Println(trace)
}

func (env *environment) log(lt logType, function, message string) {
	if !env.debug {
		return
	}
	trace := fmt.Sprintf("%s: %s\n%s", lt, function, message)
	log.Println(trace)
}

func (env *environment) Getenv(key string) string {
	defer env.trace(time.Now(), "Getenv", key)
	val := os.Getenv(key)
	env.log(Debug, "Getenv", val)
	return val
}

func (env *environment) Pwd() string {
	defer env.trace(time.Now(), "Pwd")
	defer func() {
		env.log(Debug, "Pwd", env.cwd)
	}()
	if env.cwd != "" {
		return env.cwd
	}
	correctPath := func(pwd string) string {
		// on Windows, and being case sensitive and not consistent and all, this gives silly issues
		driveLetter := getCompiledRegex(`^[a-z]:`)
		return driveLetter.ReplaceAllStringFunc(pwd, strings.ToUpper)
	}
	if env.args != nil && *env.args.PWD != "" {
		env.cwd = correctPath(*env.args.PWD)
		return env.cwd
	}
	dir, err := os.Getwd()
	if err != nil {
		env.log(Error, "Pwd", err.Error())
		return ""
	}
	env.cwd = correctPath(dir)
	return env.cwd
}

func (env *environment) HasFiles(pattern string) bool {
	defer env.trace(time.Now(), "HasFiles", pattern)
	cwd := env.Pwd()
	pattern = cwd + env.PathSeperator() + pattern
	matches, err := filepath.Glob(pattern)
	if err != nil {
		env.log(Error, "HasFiles", err.Error())
		return false
	}
	return len(matches) > 0
}

func (env *environment) HasFilesInDir(dir, pattern string) bool {
	defer env.trace(time.Now(), "HasFilesInDir", pattern)
	pattern = dir + env.PathSeperator() + pattern
	matches, err := filepath.Glob(pattern)
	if err != nil {
		env.log(Error, "HasFilesInDir", err.Error())
		return false
	}
	return len(matches) > 0
}

func (env *environment) HasFolder(folder string) bool {
	defer env.trace(time.Now(), "HasFolder", folder)
	_, err := os.Stat(folder)
	return !os.IsNotExist(err)
}

func (env *environment) FileContent(file string) string {
	defer env.trace(time.Now(), "FileContent", file)
	content, err := ioutil.ReadFile(file)
	if err != nil {
		env.log(Error, "FileContent", err.Error())
		return ""
	}
	return string(content)
}

func (env *environment) FolderList(path string) []string {
	defer env.trace(time.Now(), "FolderList", path)
	content, err := os.ReadDir(path)
	if err != nil {
		env.log(Error, "FolderList", err.Error())
		return nil
	}
	var folderNames []string
	for _, s := range content {
		if s.IsDir() {
			folderNames = append(folderNames, s.Name())
		}
	}
	return folderNames
}

func (env *environment) PathSeperator() string {
	defer env.trace(time.Now(), "PathSeperator")
	return string(os.PathSeparator)
}

func (env *environment) User() string {
	defer env.trace(time.Now(), "User")
	user := os.Getenv("USER")
	if user == "" {
		user = os.Getenv("USERNAME")
	}
	return user
}

func (env *environment) Host() (string, error) {
	defer env.trace(time.Now(), "Host")
	hostName, err := os.Hostname()
	if err != nil {
		env.log(Error, "Host", err.Error())
		return "", err
	}
	return cleanHostName(hostName), nil
}

func (env *environment) GOOS() string {
	defer env.trace(time.Now(), "GOOS")
	return runtime.GOOS
}

func (env *environment) RunCommand(command string, args ...string) (string, error) {
	defer env.trace(time.Now(), "RunCommand", append([]string{command}, args...)...)
	if cmd, ok := env.cmdCache.get(command); ok {
		command = cmd
	}
	cmd := exec.Command(command, args...)
	var out bytes.Buffer
	var err bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &err
	cmdErr := cmd.Run()
	if cmdErr != nil {
		output := err.String()
		errorStr := fmt.Sprintf("cmd.Start() failed with '%s'", output)
		env.log(Error, "RunCommand", errorStr)
		return output, cmdErr
	}
	output := strings.TrimSuffix(out.String(), "\n")
	env.log(Debug, "RunCommand", output)
	return output, nil
}

func (env *environment) RunShellCommand(shell, command string) string {
	defer env.trace(time.Now(), "RunShellCommand", shell, command)
	out, _ := env.RunCommand(shell, "-c", command)
	return out
}

func (env *environment) HasCommand(command string) bool {
	defer env.trace(time.Now(), "HasCommand", command)
	if _, ok := env.cmdCache.get(command); ok {
		return true
	}
	path, err := exec.LookPath(command)
	if err == nil {
		env.cmdCache.set(command, path)
		return true
	}
	env.log(Error, "HasCommand", err.Error())
	return false
}

func (env *environment) ErrorCode() int {
	defer env.trace(time.Now(), "ErrorCode")
	return *env.args.ErrorCode
}

func (env *environment) ExecutionTime() float64 {
	defer env.trace(time.Now(), "ExecutionTime")
	if *env.args.ExecutionTime < 0 {
		return 0
	}
	return *env.args.ExecutionTime
}

func (env *environment) Args() *Args {
	defer env.trace(time.Now(), "Args")
	return env.args
}

func (env *environment) BatteryInfo() ([]*battery.Battery, error) {
	defer env.trace(time.Now(), "BatteryInfo")
	batteries, err := battery.GetAll()
	// actual error, return it
	if err != nil && len(batteries) == 0 {
		env.log(Error, "BatteryInfo", err.Error())
		return nil, err
	}
	// there are no batteries found
	if len(batteries) == 0 {
		return nil, &noBatteryError{}
	}
	// some batteries fail to get retrieved, filter them out if present
	validBatteries := []*battery.Battery{}
	for _, batt := range batteries {
		if batt != nil {
			validBatteries = append(validBatteries, batt)
		}
	}
	// clean minor errors
	unableToRetrieveBatteryInfo := "A device which does not exist was specified."
	unknownChargeRate := "Unknown value received"
	var fatalErr battery.Errors
	ignoreErr := func(err error) bool {
		if e, ok := err.(battery.ErrPartial); ok {
			// ignore unknown charge rate value error
			if e.Current == nil &&
				e.Design == nil &&
				e.DesignVoltage == nil &&
				e.Full == nil &&
				e.State == nil &&
				e.Voltage == nil &&
				e.ChargeRate != nil &&
				e.ChargeRate.Error() == unknownChargeRate {
				return true
			}
		}
		return false
	}
	if batErr, ok := err.(battery.Errors); ok {
		for _, err := range batErr {
			if !ignoreErr(err) {
				fatalErr = append(fatalErr, err)
			}
		}
	}

	// when battery info fails to get retrieved but there is at least one valid battery, return it without error
	if len(validBatteries) > 0 && fatalErr != nil && strings.Contains(fatalErr.Error(), unableToRetrieveBatteryInfo) {
		return validBatteries, nil
	}
	// another error occurred (possibly unmapped use-case), return it
	if fatalErr != nil {
		env.log(Error, "BatteryInfo", fatalErr.Error())
		return nil, fatalErr
	}
	// everything is fine
	return validBatteries, nil
}

func (env *environment) Shell() string {
	defer env.trace(time.Now(), "Shell")
	if *env.args.Shell != "" {
		return *env.args.Shell
	}
	pid := os.Getppid()
	p, _ := process.NewProcess(int32(pid))
	name, err := p.Name()
	if err != nil {
		env.log(Error, "Shell", err.Error())
		return unknown
	}
	if name == "cmd.exe" {
		p, _ = p.Parent()
		name, err = p.Name()
	}
	if err != nil {
		env.log(Error, "Shell", err.Error())
		return unknown
	}
	// Cache the shell value to speed things up.
	*env.args.Shell = strings.Trim(strings.Replace(name, ".exe", "", 1), " ")
	return *env.args.Shell
}

func (env *environment) HTTPRequest(url string, timeout int, requestModifiers ...HTTPRequestModifier) ([]byte, error) {
	defer env.trace(time.Now(), "HTTPRequest", url)
	ctx, cncl := context.WithTimeout(context.Background(), time.Millisecond*time.Duration(timeout))
	defer cncl()
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	for _, modifier := range requestModifiers {
		modifier(request)
	}
	response, err := client.Do(request)
	if err != nil {
		env.log(Error, "HTTPRequest", err.Error())
		return nil, err
	}
	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		env.log(Error, "HTTPRequest", err.Error())
		return nil, err
	}
	return body, nil
}

func (env *environment) HasParentFilePath(path string) (*FileInfo, error) {
	defer env.trace(time.Now(), "HasParentFilePath", path)
	currentFolder := env.Pwd()
	for {
		searchPath := filepath.Join(currentFolder, path)
		info, err := os.Stat(searchPath)
		if err == nil {
			return &FileInfo{
				ParentFolder: currentFolder,
				Path:         searchPath,
				IsDir:        info.IsDir(),
			}, nil
		}
		if !os.IsNotExist(err) {
			return nil, err
		}
		if dir := filepath.Dir(currentFolder); dir != currentFolder {
			currentFolder = dir
			continue
		}
		env.log(Error, "HasParentFilePath", err.Error())
		return nil, errors.New("no match at root level")
	}
}

func (env *environment) StackCount() int {
	defer env.trace(time.Now(), "StackCount")
	if *env.args.StackCount < 0 {
		return 0
	}
	return *env.args.StackCount
}

func (env *environment) Cache() Cache {
	return env.fileCache
}

func (env *environment) Close() {
	env.fileCache.Close()
}

func (env *environment) Logs() string {
	return env.logBuilder.String()
}

func (env *environment) TemplateCache() *TemplateCache {
	defer env.trace(time.Now(), "TemplateCache")
	if env.tmplCache != nil {
		return env.tmplCache
	}
	tmplCache := &TemplateCache{
		Root:  env.Root(),
		Shell: env.Shell(),
		Code:  env.ErrorCode(),
		WSL:   env.IsWsl(),
	}
	tmplCache.Env = make(map[string]string)
	const separator = "="
	values := os.Environ()
	for value := range values {
		splitted := strings.Split(values[value], separator)
		if len(splitted) != 2 {
			continue
		}
		key := splitted[0]
		val := splitted[1:]
		tmplCache.Env[key] = strings.Join(val, separator)
	}
	pwd := env.Pwd()
	pwd = strings.Replace(pwd, env.Home(), "~", 1)
	tmplCache.PWD = pwd
	tmplCache.Folder = base(pwd, env)
	tmplCache.UserName = env.User()
	if host, err := env.Host(); err == nil {
		tmplCache.HostName = host
	}
	goos := env.GOOS()
	tmplCache.OS = goos
	env.tmplCache = tmplCache
	return env.tmplCache
}

func cleanHostName(hostName string) string {
	garbage := []string{
		".lan",
		".local",
		".localdomain",
	}
	for _, g := range garbage {
		if strings.HasSuffix(hostName, g) {
			hostName = strings.Replace(hostName, g, "", 1)
		}
	}
	return hostName
}

func returnOrBuildCachePath(path string) string {
	// validate root path
	if _, err := os.Stat(path); err != nil {
		return ""
	}
	// validate oh-my-posh folder, if non existent, create it
	cachePath := path + "/oh-my-posh"
	if _, err := os.Stat(cachePath); err == nil {
		return cachePath
	}
	if err := os.Mkdir(cachePath, 0755); err != nil {
		return ""
	}
	return cachePath
}
