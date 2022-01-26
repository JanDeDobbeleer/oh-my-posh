package main

import (
	"fmt"
	"oh-my-posh/environment"
	"oh-my-posh/regex"
	"sort"
	"strings"
)

type path struct {
	props Properties
	env   environment.Environment

	pwd        string
	Path       string
	StackCount int
	Location   string
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
	// AgnosterLeft like agnoster, but keeps the left side of the path
	AgnosterLeft string = "agnoster_left"
	// MixedThreshold the threshold of the length of the path Mixed will display
	MixedThreshold Property = "mixed_threshold"
	// MappedLocations allows overriding certain location with an icon
	MappedLocations Property = "mapped_locations"
	// MappedLocationsEnabled enables overriding certain locations with an icon
	MappedLocationsEnabled Property = "mapped_locations_enabled"
	// MaxDepth Maximum path depth to display whithout shortening
	MaxDepth Property = "max_depth"
)

func (pt *path) template() string {
	return "{{ .Path }}"
}

func (pt *path) enabled() bool {
	pt.pwd = pt.env.Pwd()
	switch style := pt.props.GetString(Style, Agnoster); style {
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
	return true
}

func (pt *path) formatWindowsDrive(pwd string) string {
	if pt.env.GOOS() != environment.WindowsPlatform || !strings.HasSuffix(pwd, ":") {
		return pwd
	}
	return pwd + "\\"
}

func (pt *path) init(props Properties, env environment.Environment) {
	pt.props = props
	pt.env = env
}

func (pt *path) getMixedPath() string {
	var buffer strings.Builder
	pwd := pt.getPwd()
	splitted := strings.Split(pwd, pt.env.PathSeperator())
	threshold := int(pt.props.GetFloat64(MixedThreshold, 4))
	for i, part := range splitted {
		if part == "" {
			continue
		}

		folder := part
		if len(part) > threshold && i != 0 && i != len(splitted)-1 {
			folder = pt.props.GetString(FolderIcon, "..")
		}
		separator := pt.props.GetString(FolderSeparatorIcon, pt.env.PathSeperator())
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
	folderIcon := pt.props.GetString(FolderIcon, "..")
	separator := pt.props.GetString(FolderSeparatorIcon, pt.env.PathSeperator())
	for i := 1; i < pathDepth; i++ {
		buffer.WriteString(fmt.Sprintf("%s%s", separator, folderIcon))
	}
	if pathDepth > 0 {
		buffer.WriteString(fmt.Sprintf("%s%s", separator, environment.Base(pt.env, pwd)))
	}
	return buffer.String()
}

func (pt *path) getAgnosterLeftPath() string {
	pwd := pt.getPwd()
	separator := pt.env.PathSeperator()
	pwd = strings.Trim(pwd, separator)
	splitted := strings.Split(pwd, separator)
	folderIcon := pt.props.GetString(FolderIcon, "..")
	separator = pt.props.GetString(FolderSeparatorIcon, separator)
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

func (pt *path) getLetterPath() string {
	var buffer strings.Builder
	pwd := pt.getPwd()
	splitted := strings.Split(pwd, pt.env.PathSeperator())
	separator := pt.props.GetString(FolderSeparatorIcon, pt.env.PathSeperator())
	for i := 0; i < len(splitted)-1; i++ {
		folder := splitted[i]
		if len(folder) == 0 {
			continue
		}

		// check if there is at least a letter we can use
		matches := regex.FindNamedRegexMatch(`(?P<letter>[\p{L}0-9]).*`, folder)

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
	if len(splitted) > 0 {
		buffer.WriteString(splitted[len(splitted)-1])
	}
	return buffer.String()
}

func (pt *path) getAgnosterFullPath() string {
	pwd := pt.getPwd()
	if len(pwd) > 1 && string(pwd[0]) == pt.env.PathSeperator() {
		pwd = pwd[1:]
	}
	return pt.replaceFolderSeparators(pwd)
}

func (pt *path) getAgnosterShortPath() string {
	pwd := pt.getPwd()
	pathDepth := pt.pathDepth(pwd)
	maxDepth := pt.props.GetInt(MaxDepth, 1)
	if maxDepth < 1 {
		maxDepth = 1
	}
	if pathDepth <= maxDepth {
		return pt.getAgnosterFullPath()
	}
	pathSeparator := pt.env.PathSeperator()
	folderSeparator := pt.props.GetString(FolderSeparatorIcon, pathSeparator)
	folderIcon := pt.props.GetString(FolderIcon, "..")
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
	pwd = environment.Base(pt.env, pwd)
	return pt.replaceFolderSeparators(pwd)
}

func (pt *path) getPwd() string {
	pwd := *pt.env.Args().PSWD
	if pwd == "" {
		pwd = pt.env.Pwd()
	}
	pwd = pt.replaceMappedLocations(pwd)
	return pwd
}

func (pt *path) normalize(inputPath string) string {
	normalized := inputPath
	if strings.HasPrefix(inputPath, "~") {
		normalized = pt.env.Home() + normalized[1:]
	}
	normalized = strings.ReplaceAll(normalized, "\\", "/")
	goos := pt.env.GOOS()
	if goos == environment.WindowsPlatform || goos == environment.DarwinPlatform {
		normalized = strings.ToLower(normalized)
	}
	return normalized
}

func (pt *path) replaceMappedLocations(pwd string) string {
	if strings.HasPrefix(pwd, "Microsoft.PowerShell.Core\\FileSystem::") {
		pwd = strings.Replace(pwd, "Microsoft.PowerShell.Core\\FileSystem::", "", 1)
	}

	mappedLocations := map[string]string{}
	if pt.props.GetBool(MappedLocationsEnabled, true) {
		mappedLocations["HKCU:"] = pt.props.GetString(WindowsRegistryIcon, "\uF013")
		mappedLocations["HKLM:"] = pt.props.GetString(WindowsRegistryIcon, "\uF013")
		mappedLocations[pt.normalize(pt.env.Home())] = pt.props.GetString(HomeIcon, "~")
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
	return pwd
}

func (pt *path) replaceFolderSeparators(pwd string) string {
	defaultSeparator := pt.env.PathSeperator()
	if pwd == defaultSeparator {
		return pwd
	}
	folderSeparator := pt.props.GetString(FolderSeparatorIcon, defaultSeparator)
	if folderSeparator == defaultSeparator {
		return pwd
	}

	pwd = strings.ReplaceAll(pwd, defaultSeparator, folderSeparator)
	return pwd
}

func (pt *path) inHomeDir(pwd string) bool {
	return strings.HasPrefix(pwd, pt.env.Home())
}

func (pt *path) rootLocation() string {
	pwd := pt.getPwd()
	pwd = strings.TrimPrefix(pwd, pt.env.PathSeperator())
	splitted := strings.Split(pwd, pt.env.PathSeperator())
	rootLocation := splitted[0]
	return rootLocation
}

func (pt *path) pathDepth(pwd string) int {
	splitted := strings.Split(pwd, pt.env.PathSeperator())
	depth := 0
	for _, part := range splitted {
		if part != "" {
			depth++
		}
	}
	return depth - 1
}
