package main

import (
	"bytes"
	"fmt"
	"path/filepath"
	"strings"
)

type path struct {
	props *properties
	env   environmentInfo
}

const (
	//FolderSeparatorIcon the path which is split will be separated by this icon
	FolderSeparatorIcon Property = "folder_separator_icon"
	//HomeIcon indicates the $HOME location
	HomeIcon Property = "home_icon"
	//FolderIcon identifies one folder
	FolderIcon Property = "folder_icon"
	//WindowsRegistryIcon indicates the registry location on Windows
	WindowsRegistryIcon Property = "windows_registry_icon"
	//Agnoster displays a short path with separator icon, this the default style
	Agnoster string = "agnoster"
	//AgnosterFull displays all the folder names with the folder_separator_icon
	AgnosterFull string = "agnoster_full"
	//Short displays a shorter path
	Short string = "short"
	//Full displays the full path
	Full string = "full"
	//Folder displays the current folder
	Folder string = "folder"
)

func (pt *path) enabled() bool {
	return true
}

func (pt *path) string() string {
	switch style := pt.props.getString(Style, Agnoster); style {
	case Agnoster:
		return pt.getAgnosterPath()
	case AgnosterFull:
		return pt.getAgnosterFullPath()
	case Short:
		return pt.getShortPath()
	case Full:
		return pt.env.getcwd()
	case Folder:
		return base(pt.env.getcwd(), pt.env)
	default:
		return fmt.Sprintf("Path style: %s is not available", style)
	}
}

func (pt *path) init(props *properties, env environmentInfo) {
	pt.props = props
	pt.env = env
}

func (pt *path) getShortPath() string {
	pwd := pt.env.getcwd()
	mappedLocations := map[string]string{
		"HKCU:": pt.props.getString(WindowsRegistryIcon, "\uE0B1"),
		"HKLM:": pt.props.getString(WindowsRegistryIcon, "\uE0B1"),
		"Microsoft.PowerShell.Core\\FileSystem::": "",
		pt.env.homeDir(): pt.props.getString(HomeIcon, "~"),
	}
	for location, value := range mappedLocations {
		if strings.HasPrefix(pwd, location) {
			return strings.Replace(pwd, location, value, 1)
		}
	}
	return pwd
}

func (pt *path) getAgnosterPath() string {
	pwd := pt.env.getcwd()
	buffer := new(bytes.Buffer)
	buffer.WriteString(pt.rootLocation(pwd))
	pathDepth := pt.pathDepth(pwd)
	for i := 1; i < pathDepth; i++ {
		buffer.WriteString(fmt.Sprintf("%s%s", pt.props.getString(FolderSeparatorIcon, pt.env.getPathSeperator()), pt.props.getString(FolderIcon, "..")))
	}
	if pathDepth > 0 {
		buffer.WriteString(fmt.Sprintf("%s%s", pt.props.getString(FolderSeparatorIcon, pt.env.getPathSeperator()), base(pwd, pt.env)))
	}
	return buffer.String()
}

func (pt *path) getAgnosterFullPath() string {
	pwd := pt.env.getcwd()
	pathSeparator := pt.env.getPathSeperator()
	folderSeparator := pt.props.getString(FolderSeparatorIcon, pathSeparator)
	if string(pwd[0]) == pathSeparator {
		pwd = pwd[1:]
	}
	return strings.Replace(pwd, pathSeparator, folderSeparator, -1)
}

func (pt *path) inHomeDir(pwd string) bool {
	return strings.HasPrefix(pwd, pt.env.homeDir())
}

func (pt *path) rootLocation(pwd string) string {
	//See https://community.idera.com/database-tools/powershell/powertips/b/tips/posts/correcting-powershell-paths
	if strings.HasPrefix(pwd, "Microsoft.PowerShell.Core\\FileSystem::") {
		pwd = strings.Replace(pwd, "Microsoft.PowerShell.Core\\FileSystem::", "", 1)
	}
	if pt.inHomeDir(pwd) {
		return pt.props.getString(HomeIcon, "~")
	}
	pwd = strings.TrimPrefix(pwd, pt.env.getPathSeperator())
	splitted := strings.Split(pwd, pt.env.getPathSeperator())
	rootLocation := splitted[0]
	mappedLocations := map[string]string{
		"HKCU:": pt.props.getString(WindowsRegistryIcon, "\uE0B1"),
	}
	if val, ok := mappedLocations[rootLocation]; ok {
		return val
	}
	return rootLocation
}

func (pt *path) pathDepth(pwd string) int {
	if pt.inHomeDir(pwd) {
		pwd = strings.Replace(pwd, pt.env.homeDir(), "root", 1)
	}
	splitted := strings.Split(pwd, pt.env.getPathSeperator())
	var validParts []string
	for _, part := range splitted {
		if part != "" {
			validParts = append(validParts, part)
		}
	}
	depth := len(validParts)
	return depth - 1
}

// Base returns the last element of path.
// Trailing path separators are removed before extracting the last element.
// If the path is empty, Base returns ".".
// If the path consists entirely of separators, Base returns a single separator.
func base(path string, env environmentInfo) string {
	if path == "" {
		return "."
	}
	// Strip trailing slashes.
	for len(path) > 0 && string(path[len(path)-1]) == env.getPathSeperator() {
		path = path[0 : len(path)-1]
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
		return string(env.getPathSeperator())
	}
	return path
}
