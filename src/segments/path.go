package segments

import (
	"fmt"
	"oh-my-posh/environment"
	"oh-my-posh/properties"
	"oh-my-posh/regex"
	"oh-my-posh/template"
	"path/filepath"
	"sort"
	"strings"
)

type Path struct {
	props properties.Properties
	env   environment.Environment

	pwd        string
	Path       string
	StackCount int
	Location   string
	Writable   bool
}

const (
	// FolderSeparatorIcon the path which is split will be separated by this icon
	FolderSeparatorIcon properties.Property = "folder_separator_icon"
	// FolderSeparatorTemplate the path which is split will be separated by this template
	FolderSeparatorTemplate properties.Property = "folder_separator_template"
	// HomeIcon indicates the $HOME location
	HomeIcon properties.Property = "home_icon"
	// FolderIcon identifies one folder
	FolderIcon properties.Property = "folder_icon"
	// WindowsRegistryIcon indicates the registry location on Windows
	WindowsRegistryIcon properties.Property = "windows_registry_icon"
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
	// Unique like agnoster, but with the first unique letters of each folder name
	Unique string = "unique"
	// AgnosterLeft like agnoster, but keeps the left side of the path
	AgnosterLeft string = "agnoster_left"
	// MixedThreshold the threshold of the length of the path Mixed will display
	MixedThreshold properties.Property = "mixed_threshold"
	// MappedLocations allows overriding certain location with an icon
	MappedLocations properties.Property = "mapped_locations"
	// MappedLocationsEnabled enables overriding certain locations with an icon
	MappedLocationsEnabled properties.Property = "mapped_locations_enabled"
	// MaxDepth Maximum path depth to display whithout shortening
	MaxDepth properties.Property = "max_depth"
	// Hides the root location if it doesn't fit in max_depth. Used in Agnoster Short
	HideRootLocation properties.Property = "hide_root_location"
)

func (pt *Path) Template() string {
	return " {{ .Path }} "
}

func (pt *Path) Enabled() bool {
	pt.pwd = pt.env.Pwd()
	switch style := pt.props.GetString(properties.Style, Agnoster); style {
	case Agnoster:
		pt.Path = pt.getAgnosterPath()
	case AgnosterFull:
		pt.Path = pt.getAgnosterFullPath()
	case AgnosterShort:
		pt.Path = pt.getAgnosterShortPath()
	case Mixed:
		pt.Path = pt.getMixedPath()
	case Letter:
		pt.Path = pt.getLetterPath()
	case Unique:
		pt.Path = pt.getUniqueLettersPath()
	case AgnosterLeft:
		pt.Path = pt.getAgnosterLeftPath()
	case Short:
		// "short" is a duplicate of "full", just here for backwards compatibility
		fallthrough
	case Full:
		pt.Path = pt.getFullPath()
	case Folder:
		pt.Path = pt.getFolderPath()
	default:
		pt.Path = fmt.Sprintf("Path style: %s is not available", style)
	}
	pt.Path = pt.formatWindowsDrive(pt.Path)
	if pt.env.IsWsl() {
		pt.Location, _ = pt.env.RunCommand("wslpath", "-m", pt.pwd)
	} else {
		pt.Location = pt.pwd
	}

	pt.StackCount = pt.env.StackCount()
	pt.Writable = pt.env.DirIsWritable(pt.pwd)
	return true
}

func (pt *Path) Parent() string {
	if pt.pwd == pt.env.Home() {
		return ""
	}
	parent := filepath.Dir(pt.pwd)
	if pt.pwd == parent {
		return ""
	}
	separator := pt.env.PathSeparator()
	if parent == pt.rootLocation() || parent == separator {
		separator = ""
	}
	return pt.replaceMappedLocations(parent) + separator
}

func (pt *Path) formatWindowsDrive(pwd string) string {
	if pt.env.GOOS() != environment.WINDOWS || !strings.HasSuffix(pwd, ":") {
		return pwd
	}
	return pwd + "\\"
}

func (pt *Path) Init(props properties.Properties, env environment.Environment) {
	pt.props = props
	pt.env = env
}

func (pt *Path) getFolderSeparator() string {
	separatorTemplate := pt.props.GetString(FolderSeparatorTemplate, "")
	if len(separatorTemplate) == 0 {
		return pt.props.GetString(FolderSeparatorIcon, pt.env.PathSeparator())
	}
	tmpl := &template.Text{
		Template: separatorTemplate,
		Context:  pt,
		Env:      pt.env,
	}
	text, err := tmpl.Render()
	if err != nil {
		pt.env.Log(environment.Error, "getFolderSeparator", err.Error())
	}
	return text
}

func (pt *Path) getMixedPath() string {
	var buffer strings.Builder
	pwd := pt.getPwd()
	splitted := strings.Split(pwd, pt.env.PathSeparator())
	threshold := int(pt.props.GetFloat64(MixedThreshold, 4))
	for i, part := range splitted {
		if part == "" {
			continue
		}

		folder := part
		if len(part) > threshold && i != 0 && i != len(splitted)-1 {
			folder = pt.props.GetString(FolderIcon, "..")
		}
		separator := pt.getFolderSeparator()
		if i == 0 {
			separator = ""
		}
		buffer.WriteString(fmt.Sprintf("%s%s", separator, folder))
	}

	return buffer.String()
}

func (pt *Path) getAgnosterPath() string {
	var buffer strings.Builder
	pwd := pt.getPwd()
	buffer.WriteString(pt.rootLocation())
	pathDepth := pt.pathDepth(pwd)
	folderIcon := pt.props.GetString(FolderIcon, "..")
	separator := pt.getFolderSeparator()
	for i := 1; i < pathDepth; i++ {
		buffer.WriteString(fmt.Sprintf("%s%s", separator, folderIcon))
	}
	if pathDepth > 0 {
		buffer.WriteString(fmt.Sprintf("%s%s", separator, environment.Base(pt.env, pwd)))
	}
	return buffer.String()
}

func (pt *Path) getAgnosterLeftPath() string {
	pwd := pt.getPwd()
	separator := pt.env.PathSeparator()
	pwd = strings.Trim(pwd, separator)
	splitted := strings.Split(pwd, separator)
	folderIcon := pt.props.GetString(FolderIcon, "..")
	separator = pt.getFolderSeparator()
	switch len(splitted) {
	case 0:
		return ""
	case 1:
		return splitted[0]
	case 2:
		return fmt.Sprintf("%s%s%s", splitted[0], separator, splitted[1])
	}
	var buffer strings.Builder
	buffer.WriteString(fmt.Sprintf("%s%s%s", splitted[0], separator, splitted[1]))
	for i := 2; i < len(splitted); i++ {
		buffer.WriteString(fmt.Sprintf("%s%s", separator, folderIcon))
	}
	return buffer.String()
}

func (pt *Path) getRelevantLetter(folder string) string {
	// check if there is at least a letter we can use
	matches := regex.FindNamedRegexMatch(`(?P<letter>[\p{L}0-9]).*`, folder)
	if matches == nil || matches["letter"] == "" {
		// no letter found, keep the folder unchanged
		return folder
	}
	letter := matches["letter"]
	// handle non-letter characters before the first found letter
	letter = folder[0:strings.Index(folder, letter)] + letter
	return letter
}

func (pt *Path) getLetterPath() string {
	var buffer strings.Builder
	pwd := pt.getPwd()
	splitted := strings.Split(pwd, pt.env.PathSeparator())
	separator := pt.getFolderSeparator()
	for i := 0; i < len(splitted)-1; i++ {
		folder := splitted[i]
		if len(folder) == 0 {
			continue
		}
		letter := pt.getRelevantLetter(folder)
		buffer.WriteString(fmt.Sprintf("%s%s", letter, separator))
	}
	if len(splitted) > 0 {
		buffer.WriteString(splitted[len(splitted)-1])
	}
	return buffer.String()
}

func (pt *Path) getUniqueLettersPath() string {
	var buffer strings.Builder
	pwd := pt.getPwd()
	splitted := strings.Split(pwd, pt.env.PathSeparator())
	separator := pt.getFolderSeparator()
	letters := make(map[string]bool, len(splitted))
	for i := 0; i < len(splitted)-1; i++ {
		folder := splitted[i]
		if len(folder) == 0 {
			continue
		}
		letter := pt.getRelevantLetter(folder)
		for letters[letter] {
			if letter == folder {
				break
			}
			letter += folder[len(letter) : len(letter)+1]
		}
		letters[letter] = true
		buffer.WriteString(fmt.Sprintf("%s%s", letter, separator))
	}
	if len(splitted) > 0 {
		buffer.WriteString(splitted[len(splitted)-1])
	}
	return buffer.String()
}

func (pt *Path) getAgnosterFullPath() string {
	pwd := pt.getPwd()
	for len(pwd) > 1 && string(pwd[0]) == pt.env.PathSeparator() {
		pwd = pwd[1:]
	}
	return pt.replaceFolderSeparators(pwd)
}

func (pt *Path) getAgnosterShortPath() string {
	pwd := pt.getPwd()
	pathDepth := pt.pathDepth(pwd)
	maxDepth := pt.props.GetInt(MaxDepth, 1)
	if maxDepth < 1 {
		maxDepth = 1
	}
	hideRootLocation := pt.props.GetBool(HideRootLocation, false)
	if hideRootLocation {
		// 1-indexing to avoid showing the root location when exceeding the max depth
		pathDepth++
	}
	if pathDepth <= maxDepth {
		return pt.getAgnosterFullPath()
	}
	folderSeparator := pt.getFolderSeparator()
	pathSeparator := pt.env.PathSeparator()
	splitted := strings.Split(pwd, pathSeparator)
	fullPathDepth := len(splitted)
	splitPos := fullPathDepth - maxDepth
	var buffer strings.Builder
	if hideRootLocation {
		buffer.WriteString(splitted[splitPos])
		splitPos++
	} else {
		folderIcon := pt.props.GetString(FolderIcon, "..")
		root := pt.rootLocation()
		buffer.WriteString(fmt.Sprintf("%s%s%s", root, folderSeparator, folderIcon))
	}
	for i := splitPos; i < fullPathDepth; i++ {
		buffer.WriteString(fmt.Sprintf("%s%s", folderSeparator, splitted[i]))
	}
	return buffer.String()
}

func (pt *Path) getFullPath() string {
	pwd := pt.getPwd()
	return pt.replaceFolderSeparators(pwd)
}

func (pt *Path) getFolderPath() string {
	pwd := pt.getPwd()
	pwd = environment.Base(pt.env, pwd)
	return pt.replaceFolderSeparators(pwd)
}

func (pt *Path) getPwd() string {
	pwd := pt.env.Flags().PSWD
	if pwd == "" {
		pwd = pt.env.Pwd()
	}
	pwd = pt.replaceMappedLocations(pwd)
	return pwd
}

func (pt *Path) normalize(inputPath string) string {
	normalized := inputPath
	if strings.HasPrefix(inputPath, "~") {
		normalized = pt.env.Home() + normalized[1:]
	}
	normalized = strings.ReplaceAll(normalized, "\\", "/")
	goos := pt.env.GOOS()
	if goos == environment.WINDOWS || goos == environment.DARWIN {
		normalized = strings.ToLower(normalized)
	}
	return normalized
}

func (pt *Path) replaceMappedLocations(pwd string) string {
	if strings.HasPrefix(pwd, "Microsoft.PowerShell.Core\\FileSystem::") {
		pwd = strings.Replace(pwd, "Microsoft.PowerShell.Core\\FileSystem::", "", 1)
	}

	enableMappedLocations := pt.props.GetBool(MappedLocationsEnabled, true)

	// built-in mapped locations, can be disabled
	mappedLocations := map[string]string{}
	if enableMappedLocations {
		mappedLocations["hkcu:"] = pt.props.GetString(WindowsRegistryIcon, "\uF013")
		mappedLocations["hklm:"] = pt.props.GetString(WindowsRegistryIcon, "\uF013")
	}

	// merge custom locations with mapped locations
	// mapped locations can override predefined locations
	keyValues := pt.props.GetKeyValueMap(MappedLocations, make(map[string]string))
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

	// treat this as a built-in mapped location
	if enableMappedLocations && pt.inHomeDir(pwd) {
		return pt.props.GetString(HomeIcon, "~") + pwd[len(pt.env.Home()):]
	}

	return pwd
}

func (pt *Path) replaceFolderSeparators(pwd string) string {
	defaultSeparator := pt.env.PathSeparator()
	if pwd == defaultSeparator {
		return pwd
	}
	folderSeparator := pt.getFolderSeparator()
	if folderSeparator == defaultSeparator {
		return pwd
	}

	pwd = strings.ReplaceAll(pwd, defaultSeparator, folderSeparator)
	return pwd
}

func (pt *Path) inHomeDir(pwd string) bool {
	home := pt.env.Home()
	if home == pwd {
		return true
	}
	if !strings.HasSuffix(home, pt.env.PathSeparator()) {
		home += pt.env.PathSeparator()
	}
	return strings.HasPrefix(pwd, home)
}

func (pt *Path) rootLocation() string {
	pwd := pt.getPwd()
	pwd = strings.TrimPrefix(pwd, pt.env.PathSeparator())
	splitted := strings.Split(pwd, pt.env.PathSeparator())
	rootLocation := splitted[0]
	return rootLocation
}

func (pt *Path) pathDepth(pwd string) int {
	splitted := strings.Split(pwd, pt.env.PathSeparator())
	depth := 0
	for _, part := range splitted {
		if part != "" {
			depth++
		}
	}
	return depth - 1
}
