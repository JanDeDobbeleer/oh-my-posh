package main

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"
)

type path struct {
	props *properties
	env   environmentInfo
}

const (
	// FolderSeparatorIcon the path which is split will be separated by this icon
	FolderSeparatorIcon Property = "folder_separator_icon"
	// HomeIcon indicates the $HOME location
	HomeIcon Property = "home_icon"
	// FolderIcon identifies one folder
	FolderIcon Property = "folder_icon"
	// WindowsRegistryIcon indicates the registry location on Windows
	WindowsRegistryIcon Property = "windows_registry_icon"
	// Agnoster displays a short path with separator icon, this the default style
	Agnoster string = "agnoster"
	// AgnosterFull displays all the folder names with the folder_separator_icon
	AgnosterFull string = "agnoster_full"
	// AgnosterShort displays the folder names with one folder_separator_icon, regardless of depth
	AgnosterShort string = "agnoster_short"
	// Short displays a shorter path
	Short string = "short"
	// Full displays the full path
	Full string = "full"
	// Folder displays the current folder
	Folder string = "folder"
	// Mixed like agnoster, but if the path is short it displays it
	Mixed string = "mixed"
	// Letter like agnoster, but with the first letter of each folder name
	Letter string = "letter"
	// MixedThreshold the threshold of the length of the path Mixed will display
	MixedThreshold Property = "mixed_threshold"
	// MappedLocations allows overriding certain location with an icon
	MappedLocations Property = "mapped_locations"
	// MappedLocationsEnabled enables overriding certain locations with an icon
	MappedLocationsEnabled Property = "mapped_locations_enabled"
	// StackCountEnabled enables the stack count display
	StackCountEnabled Property = "stack_count_enabled"
	// MaxDepth Maximum path depth to display whithout shortening
	MaxDepth Property = "max_depth"
)

func (pt *path) enabled() bool {
	return true
}

func (pt *path) string() string {
	cwd := pt.env.getcwd()
	var formattedPath string
	switch style := pt.props.getString(Style, Agnoster); style {
	case Agnoster:
		formattedPath = pt.getAgnosterPath()
	case AgnosterFull:
		formattedPath = pt.getAgnosterFullPath()
	case AgnosterShort:
		formattedPath = pt.getAgnosterShortPath()
	case Mixed:
		formattedPath = pt.getMixedPath()
	case Letter:
		formattedPath = pt.getLetterPath()
	case Short:
		// "short" is a duplicate of "full", just here for backwards compatibility
		fallthrough
	case Full:
		formattedPath = pt.getFullPath()
	case Folder:
		formattedPath = pt.getFolderPath()
	default:
		return fmt.Sprintf("Path style: %s is not available", style)
	}
	formattedPath = pt.formatWindowsDrive(formattedPath)
	if pt.props.getBool(EnableHyperlink, false) {
		// wsl check
		if pt.env.isWsl() {
			cwd, _ = pt.env.runCommand("wslpath", "-m", cwd)
		}
		return fmt.Sprintf("[%s](file://%s)", formattedPath, cwd)
	}

	if pt.props.getBool(StackCountEnabled, false) && pt.env.stackCount() > 0 {
		return fmt.Sprintf("%d %s", pt.env.stackCount(), formattedPath)
	}

	return formattedPath
}

func (pt *path) formatWindowsDrive(pwd string) string {
	if pt.env.getRuntimeGOOS() != windowsPlatform || !strings.HasSuffix(pwd, ":") {
		return pwd
	}
	return pwd + "\\"
}

func (pt *path) init(props *properties, env environmentInfo) {
	pt.props = props
	pt.env = env
}

func (pt *path) getMixedPath() string {
	var buffer strings.Builder
	pwd := pt.getPwd()
	splitted := strings.Split(pwd, pt.env.getPathSeperator())
	threshold := int(pt.props.getFloat64(MixedThreshold, 4))
	for i, part := range splitted {
		if part == "" {
			continue
		}

		folder := part
		if len(part) > threshold && i != 0 && i != len(splitted)-1 {
			folder = pt.props.getString(FolderIcon, "..")
		}
		separator := pt.props.getString(FolderSeparatorIcon, pt.env.getPathSeperator())
		if i == 0 {
			separator = ""
		}
		buffer.WriteString(fmt.Sprintf("%s%s", separator, folder))
	}

	return buffer.String()
}

func (pt *path) getAgnosterPath() string {
	var buffer strings.Builder
	pwd := pt.getPwd()
	buffer.WriteString(pt.rootLocation())
	pathDepth := pt.pathDepth(pwd)
	folderIcon := pt.props.getString(FolderIcon, "..")
	separator := pt.props.getString(FolderSeparatorIcon, pt.env.getPathSeperator())
	for i := 1; i < pathDepth; i++ {
		buffer.WriteString(fmt.Sprintf("%s%s", separator, folderIcon))
	}
	if pathDepth > 0 {
		buffer.WriteString(fmt.Sprintf("%s%s", separator, base(pwd, pt.env)))
	}
	return buffer.String()
}

func (pt *path) getLetterPath() string {
	var buffer strings.Builder
	pwd := pt.getPwd()
	splitted := strings.Split(pwd, pt.env.getPathSeperator())
	separator := pt.props.getString(FolderSeparatorIcon, pt.env.getPathSeperator())
	for i := 0; i < len(splitted)-1; i++ {
		folder := splitted[i]
		if len(folder) == 0 {
			continue
		}

		// check if there is at least a letter we can use
		matches := findNamedRegexMatch(`(?P<letter>[\p{L}0-9]).*`, folder)

		if matches == nil || matches["letter"] == "" {
			// no letter found, keep the folder unchanged
			buffer.WriteString(fmt.Sprintf("%s%s", folder, separator))
			continue
		}

		letter := matches["letter"]
		// handle non-letter characters before the first found letter
		letter = folder[0:strings.Index(folder, letter)] + letter

		buffer.WriteString(fmt.Sprintf("%s%s", letter, separator))
	}
	buffer.WriteString(splitted[len(splitted)-1])
	return buffer.String()
}

func (pt *path) getAgnosterFullPath() string {
	pwd := pt.getPwd()
	if len(pwd) > 1 && string(pwd[0]) == pt.env.getPathSeperator() {
		pwd = pwd[1:]
	}
	return pt.replaceFolderSeparators(pwd)
}

func (pt *path) getAgnosterShortPath() string {
	pwd := pt.getPwd()
	pathDepth := pt.pathDepth(pwd)
	maxDepth := pt.props.getInt(MaxDepth, 1)
	if maxDepth < 1 {
		maxDepth = 1
	}
	if pathDepth <= maxDepth {
		return pt.getAgnosterFullPath()
	}
	pathSeparator := pt.env.getPathSeperator()
	folderSeparator := pt.props.getString(FolderSeparatorIcon, pathSeparator)
	folderIcon := pt.props.getString(FolderIcon, "..")
	root := pt.rootLocation()
	splitted := strings.Split(pwd, pathSeparator)
	fullPathDepth := len(splitted)
	splitPos := fullPathDepth - maxDepth
	var buffer strings.Builder
	buffer.WriteString(fmt.Sprintf("%s%s%s", root, folderSeparator, folderIcon))
	for i := splitPos; i < fullPathDepth; i++ {
		buffer.WriteString(fmt.Sprintf("%s%s", folderSeparator, splitted[i]))
	}
	return buffer.String()
}

func (pt *path) getFullPath() string {
	pwd := pt.getPwd()
	return pt.replaceFolderSeparators(pwd)
}

func (pt *path) getFolderPath() string {
	pwd := pt.getPwd()
	pwd = base(pwd, pt.env)
	return pt.replaceFolderSeparators(pwd)
}

func (pt *path) getPwd() string {
	pwd := *pt.env.getArgs().PSWD
	if pwd == "" {
		pwd = pt.env.getcwd()
	}
	pwd = pt.replaceMappedLocations(pwd)
	return pwd
}

func (pt *path) normalize(inputPath string) string {
	normalized := inputPath
	if strings.HasPrefix(inputPath, "~") {
		normalized = pt.env.homeDir() + normalized[1:]
	}
	normalized = strings.ReplaceAll(normalized, "\\", "/")
	goos := pt.env.getRuntimeGOOS()
	if goos == windowsPlatform || goos == darwinPlatform {
		normalized = strings.ToLower(normalized)
	}
	return normalized
}

func (pt *path) replaceMappedLocations(pwd string) string {
	if strings.HasPrefix(pwd, "Microsoft.PowerShell.Core\\FileSystem::") {
		pwd = strings.Replace(pwd, "Microsoft.PowerShell.Core\\FileSystem::", "", 1)
	}

	mappedLocations := map[string]string{}
	if pt.props.getBool(MappedLocationsEnabled, true) {
		mappedLocations["HKCU:"] = pt.props.getString(WindowsRegistryIcon, "\uF013")
		mappedLocations["HKLM:"] = pt.props.getString(WindowsRegistryIcon, "\uF013")
		mappedLocations[pt.normalize(pt.env.homeDir())] = pt.props.getString(HomeIcon, "~")
	}

	// merge custom locations with mapped locations
	// mapped locations can override predefined locations
	keyValues := pt.props.getKeyValueMap(MappedLocations, make(map[string]string))
	for key, val := range keyValues {
		mappedLocations[pt.normalize(key)] = val
	}

	// sort map keys in reverse order
	// fixes case when a subfoder and its parent are mapped
	// ex /users/test and /users/test/dev
	keys := make([]string, len(mappedLocations))
	i := 0
	for k := range mappedLocations {
		keys[i] = k
		i++
	}
	sort.Sort(sort.Reverse(sort.StringSlice(keys)))

	normalizedPwd := pt.normalize(pwd)
	for _, key := range keys {
		if strings.HasPrefix(normalizedPwd, key) {
			value := mappedLocations[key]
			return value + pwd[len(key):]
		}
	}
	return pwd
}

func (pt *path) replaceFolderSeparators(pwd string) string {
	defaultSeparator := pt.env.getPathSeperator()
	if pwd == defaultSeparator {
		return pwd
	}
	folderSeparator := pt.props.getString(FolderSeparatorIcon, defaultSeparator)
	if folderSeparator == defaultSeparator {
		return pwd
	}

	pwd = strings.ReplaceAll(pwd, defaultSeparator, folderSeparator)
	return pwd
}

func (pt *path) inHomeDir(pwd string) bool {
	return strings.HasPrefix(pwd, pt.env.homeDir())
}

func (pt *path) rootLocation() string {
	pwd := pt.getPwd()
	pwd = strings.TrimPrefix(pwd, pt.env.getPathSeperator())
	splitted := strings.Split(pwd, pt.env.getPathSeperator())
	rootLocation := splitted[0]
	return rootLocation
}

func (pt *path) pathDepth(pwd string) int {
	splitted := strings.Split(pwd, pt.env.getPathSeperator())
	depth := 0
	for _, part := range splitted {
		if part != "" {
			depth++
		}
	}
	return depth - 1
}

// Base returns the last element of path.
// Trailing path separators are removed before extracting the last element.
// If the path consists entirely of separators, Base returns a single separator.
func base(path string, env environmentInfo) string {
	if path == "/" {
		return path
	}
	volumeName := filepath.VolumeName(path)
	// Strip trailing slashes.
	for len(path) > 0 && string(path[len(path)-1]) == env.getPathSeperator() {
		path = path[0 : len(path)-1]
	}
	if volumeName == path {
		return path
	}
	// Throw away volume name
	path = path[len(filepath.VolumeName(path)):]
	// Find the last element
	i := len(path) - 1
	for i >= 0 && string(path[i]) != env.getPathSeperator() {
		i--
	}
	if i >= 0 {
		path = path[i+1:]
	}
	// If empty now, it had only slashes.
	if path == "" {
		return env.getPathSeperator()
	}
	return path
}
