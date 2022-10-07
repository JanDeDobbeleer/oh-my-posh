package segments

import (
	"fmt"
	"oh-my-posh/environment"
	"oh-my-posh/properties"
	"oh-my-posh/regex"
	"oh-my-posh/shell"
	"oh-my-posh/template"
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
	// Mixed like agnoster, but if a folder name is short enough, it is displayed as-is
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
	pt.setPath()
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
	pwd := pt.getPwd()
	if len(pwd) == 0 {
		return ""
	}
	root, path := environment.ParsePath(pt.env, pwd)
	if len(path) == 0 {
		// a root path has no parent
		return ""
	}
	base := environment.Base(pt.env, path)
	path = pt.replaceFolderSeparators(path[:len(path)-len(base)])
	if root != pt.env.PathSeparator() {
		root = root[:len(root)-1] + pt.getFolderSeparator()
	}
	return root + path
}

func (pt *Path) Init(props properties.Properties, env environment.Environment) {
	pt.props = props
	pt.env = env
}

func (pt *Path) setPath() {
	pwd := pt.getPwd()
	if len(pwd) == 0 {
		return
	}
	root, path := environment.ParsePath(pt.env, pwd)
	if len(path) == 0 {
		pt.Path = pt.formatRoot(root)
		return
	}
	switch style := pt.props.GetString(properties.Style, Agnoster); style {
	case Agnoster:
		pt.Path = pt.getAgnosterPath(root, path)
	case AgnosterFull:
		pt.Path = pt.getAgnosterFullPath(root, path)
	case AgnosterShort:
		pt.Path = pt.getAgnosterShortPath(root, path)
	case Mixed:
		pt.Path = pt.getMixedPath(root, path)
	case Letter:
		pt.Path = pt.getLetterPath(root, path)
	case Unique:
		pt.Path = pt.getUniqueLettersPath(root, path)
	case AgnosterLeft:
		pt.Path = pt.getAgnosterLeftPath(root, path)
	case Short:
		// "short" is a duplicate of "full", just here for backwards compatibility
		fallthrough
	case Full:
		pt.Path = pt.getFullPath(root, path)
	case Folder:
		pt.Path = pt.getFolderPath(path)
	default:
		pt.Path = fmt.Sprintf("Path style: %s is not available", style)
	}
}

func (pt *Path) getFolderSeparator() string {
	separatorTemplate := pt.props.GetString(FolderSeparatorTemplate, "")
	if len(separatorTemplate) == 0 {
		separator := pt.props.GetString(FolderSeparatorIcon, pt.env.PathSeparator())
		// if empty, use the default separator
		if len(separator) == 0 {
			return pt.env.PathSeparator()
		}
		return separator
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
	if len(text) == 0 {
		return pt.env.PathSeparator()
	}
	return text
}

func (pt *Path) getMixedPath(root, path string) string {
	var buffer strings.Builder
	threshold := int(pt.props.GetFloat64(MixedThreshold, 4))
	folderIcon := pt.props.GetString(FolderIcon, "..")
	separator := pt.getFolderSeparator()
	elements := strings.Split(path, pt.env.PathSeparator())
	if root != pt.env.PathSeparator() {
		elements = append([]string{root[:len(root)-1]}, elements...)
	}
	n := len(elements)
	buffer.WriteString(elements[0])
	for i := 1; i < n; i++ {
		folder := elements[i]
		if len(folder) > threshold && i != n-1 {
			folder = folderIcon
		}
		buffer.WriteString(fmt.Sprintf("%s%s", separator, folder))
	}
	return buffer.String()
}

func (pt *Path) getAgnosterPath(root, path string) string {
	var buffer strings.Builder
	folderIcon := pt.props.GetString(FolderIcon, "..")
	separator := pt.getFolderSeparator()
	elements := strings.Split(path, pt.env.PathSeparator())
	if root != pt.env.PathSeparator() {
		elements = append([]string{root[:len(root)-1]}, elements...)
	}
	n := len(elements)
	buffer.WriteString(elements[0])
	for i := 2; i < n; i++ {
		buffer.WriteString(fmt.Sprintf("%s%s", separator, folderIcon))
	}
	if n > 1 {
		buffer.WriteString(fmt.Sprintf("%s%s", separator, elements[n-1]))
	}
	return buffer.String()
}

func (pt *Path) getAgnosterLeftPath(root, path string) string {
	var buffer strings.Builder
	folderIcon := pt.props.GetString(FolderIcon, "..")
	separator := pt.getFolderSeparator()
	elements := strings.Split(path, pt.env.PathSeparator())
	if root != pt.env.PathSeparator() {
		elements = append([]string{root[:len(root)-1]}, elements...)
	}
	n := len(elements)
	buffer.WriteString(elements[0])
	if n > 1 {
		buffer.WriteString(fmt.Sprintf("%s%s", separator, elements[1]))
	}
	for i := 2; i < n; i++ {
		buffer.WriteString(fmt.Sprintf("%s%s", separator, folderIcon))
	}
	return buffer.String()
}

func (pt *Path) getRelevantLetter(folder string) string {
	// check if there is at least a letter we can use
	matches := regex.FindNamedRegexMatch(`(?P<letter>[\p{L}0-9]).*`, folder)
	if matches == nil || len(matches["letter"]) == 0 {
		// no letter found, keep the folder unchanged
		return folder
	}
	letter := matches["letter"]
	// handle non-letter characters before the first found letter
	letter = folder[0:strings.Index(folder, letter)] + letter
	return letter
}

func (pt *Path) getLetterPath(root, path string) string {
	var buffer strings.Builder
	separator := pt.getFolderSeparator()
	elements := strings.Split(path, pt.env.PathSeparator())
	if root != pt.env.PathSeparator() {
		elements = append([]string{root[:len(root)-1]}, elements...)
	}
	n := len(elements)
	for i := 0; i < n-1; i++ {
		letter := pt.getRelevantLetter(elements[i])
		if i != 0 {
			buffer.WriteString(separator)
		}
		buffer.WriteString(letter)
	}
	buffer.WriteString(fmt.Sprintf("%s%s", separator, elements[n-1]))
	return buffer.String()
}

func (pt *Path) getUniqueLettersPath(root, path string) string {
	var buffer strings.Builder
	separator := pt.getFolderSeparator()
	elements := strings.Split(path, pt.env.PathSeparator())
	if root != pt.env.PathSeparator() {
		elements = append([]string{root[:len(root)-1]}, elements...)
	}
	n := len(elements)
	letters := make(map[string]bool)
	for i := 0; i < n-1; i++ {
		folder := elements[i]
		letter := pt.getRelevantLetter(folder)
		for letters[letter] {
			if letter == folder {
				break
			}
			letter += folder[len(letter) : len(letter)+1]
		}
		letters[letter] = true
		if i != 0 {
			buffer.WriteString(separator)
		}
		buffer.WriteString(letter)
	}
	buffer.WriteString(fmt.Sprintf("%s%s", separator, elements[n-1]))
	return buffer.String()
}

func (pt *Path) getAgnosterFullPath(root, path string) string {
	path = pt.replaceFolderSeparators(path)
	if root == pt.env.PathSeparator() {
		return path
	}
	root = root[:len(root)-1] + pt.getFolderSeparator()
	return root + path
}

func (pt *Path) getAgnosterShortPath(root, path string) string {
	elements := strings.Split(path, pt.env.PathSeparator())
	if root != pt.env.PathSeparator() {
		elements = append([]string{root[:len(root)-1]}, elements...)
	}
	depth := len(elements)
	maxDepth := pt.props.GetInt(MaxDepth, 1)
	if maxDepth < 1 {
		maxDepth = 1
	}
	hideRootLocation := pt.props.GetBool(HideRootLocation, false)
	if !hideRootLocation {
		maxDepth++
	}
	if depth <= maxDepth {
		return pt.getAgnosterFullPath(root, path)
	}
	separator := pt.getFolderSeparator()
	folderIcon := pt.props.GetString(FolderIcon, "..")
	var buffer strings.Builder
	if !hideRootLocation {
		buffer.WriteString(fmt.Sprintf("%s%s", elements[0], separator))
		maxDepth--
	}
	splitPos := depth - maxDepth
	if splitPos != 1 {
		buffer.WriteString(fmt.Sprintf("%s%s", folderIcon, separator))
	}
	for i := splitPos; i < depth; i++ {
		buffer.WriteString(elements[i])
		if i != depth-1 {
			buffer.WriteString(separator)
		}
	}
	return buffer.String()
}

func (pt *Path) getFullPath(root, path string) string {
	if root != pt.env.PathSeparator() {
		root = root[:len(root)-1] + pt.getFolderSeparator()
	}
	path = pt.replaceFolderSeparators(path)
	return root + path
}

func (pt *Path) getFolderPath(path string) string {
	return environment.Base(pt.env, path)
}

func (pt *Path) setPwd() {
	if len(pt.pwd) > 0 {
		return
	}
	if pt.env.Shell() == shell.PWSH || pt.env.Shell() == shell.PWSH5 {
		pt.pwd = pt.env.Flags().PSWD
	}
	if len(pt.pwd) == 0 {
		pt.pwd = pt.env.Pwd()
		return
	}
	// ensure a clean path
	root, path := environment.ParsePath(pt.env, pt.pwd)
	pt.pwd = root + path
}

func (pt *Path) getPwd() string {
	pt.setPwd()
	return pt.replaceMappedLocations(pt.pwd)
}

func (pt *Path) formatRoot(root string) string {
	n := len(root)
	// trim the trailing separator first
	root = root[:n-1]
	// only preserve the trailing separator for a Unix/Windows/PSDrive root
	if len(root) == 0 || (strings.HasPrefix(pt.pwd, root) && strings.HasSuffix(root, ":")) {
		return root + pt.env.PathSeparator()
	}
	return root
}

func (pt *Path) normalize(inputPath string) string {
	normalized := inputPath
	if strings.HasPrefix(normalized, "~") && (len(normalized) == 1 || environment.IsPathSeparator(pt.env, normalized[1])) {
		normalized = pt.env.Home() + normalized[1:]
	}
	switch pt.env.GOOS() {
	case environment.WINDOWS:
		normalized = strings.ReplaceAll(normalized, "/", `\`)
		fallthrough
	case environment.DARWIN:
		normalized = strings.ToLower(normalized)
	}
	return normalized
}

func (pt *Path) replaceMappedLocations(pwd string) string {
	mappedLocations := map[string]string{}
	// predefined mapped locations, can be disabled
	if pt.props.GetBool(MappedLocationsEnabled, true) {
		mappedLocations["hkcu:"] = pt.props.GetString(WindowsRegistryIcon, "\uF013")
		mappedLocations["hklm:"] = pt.props.GetString(WindowsRegistryIcon, "\uF013")
		mappedLocations[pt.normalize(pt.env.Home())] = pt.props.GetString(HomeIcon, "~")
	}

	// merge custom locations with mapped locations
	// mapped locations can override predefined locations
	keyValues := pt.props.GetKeyValueMap(MappedLocations, make(map[string]string))
	for key, val := range keyValues {
		if key != "" {
			mappedLocations[pt.normalize(key)] = val
		}
	}

	// sort map keys in reverse order
	// fixes case when a subfoder and its parent are mapped
	// ex /users/test and /users/test/dev
	keys := make([]string, 0, len(mappedLocations))
	for k := range mappedLocations {
		keys = append(keys, k)
	}
	sort.Sort(sort.Reverse(sort.StringSlice(keys)))

	cleanPwdRoot, cleanPwdPath := environment.ParsePath(pt.env, pwd)
	pwdRoot := pt.normalize(cleanPwdRoot)
	pwdPath := pt.normalize(cleanPwdPath)
	for _, key := range keys {
		keyRoot, keyPath := environment.ParsePath(pt.env, key)
		if keyRoot != pwdRoot || !strings.HasPrefix(pwdPath, keyPath) {
			continue
		}
		value := mappedLocations[key]
		rem := cleanPwdPath[len(keyPath):]
		if len(rem) == 0 {
			// exactly match the full path
			return value
		}
		if len(keyPath) == 0 {
			// only match the root
			return value + pt.env.PathSeparator() + cleanPwdPath
		}
		// match several prefix elements
		if rem[0:1] == pt.env.PathSeparator() {
			return value + rem
		}
	}
	return cleanPwdRoot + cleanPwdPath
}

func (pt *Path) replaceFolderSeparators(pwd string) string {
	defaultSeparator := pt.env.PathSeparator()
	folderSeparator := pt.getFolderSeparator()
	if folderSeparator == defaultSeparator {
		return pwd
	}

	pwd = strings.ReplaceAll(pwd, defaultSeparator, folderSeparator)
	return pwd
}
