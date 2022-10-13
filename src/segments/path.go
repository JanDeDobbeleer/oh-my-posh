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

	root     string
	relative string
	pwd      string

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
	pt.setPaths()
	if len(pt.pwd) == 0 {
		return false
	}
	pt.setStyle()
	if pt.env.IsWsl() {
		pt.Location, _ = pt.env.RunCommand("wslpath", "-m", pt.pwd)
	} else {
		pt.Location = pt.pwd
	}
	pt.StackCount = pt.env.StackCount()
	pt.Writable = pt.env.DirIsWritable(pt.env.Pwd())
	return true
}

func (pt *Path) setPaths() {
	pt.pwd = pt.env.Pwd()
	if (pt.env.Shell() == shell.PWSH || pt.env.Shell() == shell.PWSH5) && len(pt.env.Flags().PSWD) != 0 {
		pt.pwd = pt.env.Flags().PSWD
	}
	if len(pt.pwd) == 0 {
		return
	}
	// ensure a clean path
	pt.root, pt.relative = pt.replaceMappedLocations()
	pathSeparator := pt.env.PathSeparator()
	if !strings.HasSuffix(pt.root, pathSeparator) && len(pt.relative) > 0 {
		pt.pwd = pt.root + pathSeparator + pt.relative
		return
	}
	pt.pwd = pt.root + pt.relative
}

func (pt *Path) Parent() string {
	if len(pt.pwd) == 0 {
		return ""
	}
	if len(pt.relative) == 0 {
		// a root path has no parent
		return ""
	}
	base := environment.Base(pt.env, pt.pwd)
	path := pt.replaceFolderSeparators(pt.pwd[:len(pt.pwd)-len(base)])
	return path
}

func (pt *Path) Init(props properties.Properties, env environment.Environment) {
	pt.props = props
	pt.env = env
}

func (pt *Path) setStyle() {
	if len(pt.relative) == 0 {
		pt.Path = pt.root
		return
	}
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

func (pt *Path) getMixedPath() string {
	var buffer strings.Builder
	threshold := int(pt.props.GetFloat64(MixedThreshold, 4))
	folderIcon := pt.props.GetString(FolderIcon, "..")
	separator := pt.getFolderSeparator()
	elements := strings.Split(pt.relative, pt.env.PathSeparator())
	if pt.root != pt.env.PathSeparator() {
		elements = append([]string{pt.root}, elements...)
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

func (pt *Path) pathDepth(pwd string) int {
	splitted := strings.Split(pwd, pt.env.PathSeparator())
	depth := 0
	for _, part := range splitted {
		if part != "" {
			depth++
		}
	}
	return depth
}

func (pt *Path) getAgnosterPath() string {
	var buffer strings.Builder
	folderIcon := pt.props.GetString(FolderIcon, "..")
	separator := pt.getFolderSeparator()
	elements := strings.Split(pt.relative, pt.env.PathSeparator())
	if pt.root != pt.env.PathSeparator() {
		elements = append([]string{pt.root}, elements...)
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

func (pt *Path) getAgnosterLeftPath() string {
	var buffer strings.Builder
	folderIcon := pt.props.GetString(FolderIcon, "..")
	separator := pt.getFolderSeparator()
	elements := strings.Split(pt.relative, pt.env.PathSeparator())
	if pt.root != pt.env.PathSeparator() {
		elements = append([]string{pt.root}, elements...)
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

func (pt *Path) getLetterPath() string {
	var buffer strings.Builder
	separator := pt.getFolderSeparator()
	elements := strings.Split(pt.relative, pt.env.PathSeparator())
	if pt.root != pt.env.PathSeparator() {
		elements = append([]string{pt.root}, elements...)
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

func (pt *Path) getUniqueLettersPath() string {
	var buffer strings.Builder
	separator := pt.getFolderSeparator()
	elements := strings.Split(pt.relative, pt.env.PathSeparator())
	if pt.root != pt.env.PathSeparator() {
		elements = append([]string{pt.root}, elements...)
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

func (pt *Path) getAgnosterFullPath() string {
	path := strings.Trim(pt.relative, pt.env.PathSeparator())
	path = pt.replaceFolderSeparators(path)
	if pt.root == pt.env.PathSeparator() {
		return path
	}
	return pt.root + pt.getFolderSeparator() + path
}

func (pt *Path) getAgnosterShortPath() string {
	pathDepth := pt.pathDepth(pt.relative)
	maxDepth := pt.props.GetInt(MaxDepth, 1)
	if maxDepth < 1 {
		maxDepth = 1
	}
	folderIcon := pt.props.GetString(FolderIcon, "..")
	hideRootLocation := pt.props.GetBool(HideRootLocation, false)
	if pathDepth <= maxDepth {
		if hideRootLocation {
			pt.root = folderIcon
		}
		return pt.getAgnosterFullPath()
	}
	pathSeparator := pt.env.PathSeparator()
	folderSeparator := pt.getFolderSeparator()
	rel := strings.TrimPrefix(pt.relative, pathSeparator)
	splitted := strings.Split(rel, pathSeparator)
	splitPos := pathDepth - maxDepth
	var buffer strings.Builder
	// unix root, needs to be replaced with the folder we're in at root level
	root := pt.root
	room := pathDepth - maxDepth
	if root == pathSeparator {
		root = splitted[0]
		room--
	}
	if hideRootLocation {
		buffer.WriteString(folderIcon)
	} else {
		buffer.WriteString(root)
		if room > 0 {
			buffer.WriteString(folderSeparator)
			buffer.WriteString(folderIcon)
		}
	}
	for i := splitPos; i < pathDepth; i++ {
		buffer.WriteString(fmt.Sprintf("%s%s", folderSeparator, splitted[i]))
	}
	return buffer.String()
}

func (pt *Path) getFullPath() string {
	rel := pt.relative
	if pt.root != pt.env.PathSeparator() {
		rel = pt.env.PathSeparator() + rel
	}
	path := pt.replaceFolderSeparators(rel)
	return pt.root + path
}

func (pt *Path) getFolderPath() string {
	pwd := environment.Base(pt.env, pt.pwd)
	return pt.replaceFolderSeparators(pwd)
}

func (pt *Path) replaceMappedLocations() (string, string) {
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

	root, relative := pt.parsePath(pt.pwd)
	rootN := pt.normalize(root)
	relativeN := pt.normalize(relative)
	pathSeparator := pt.env.PathSeparator()

	formatRoot := func(root string) string {
		// trim the trailing separator first
		root = strings.TrimSuffix(root, pathSeparator)
		// only preserve the trailing separator for a Unix/Windows/PSDrive root
		if len(root) == 0 || strings.HasSuffix(root, ":") {
			return root + pathSeparator
		}
		return root
	}

	for _, key := range keys {
		keyRoot, keyRelative := pt.parsePath(key)
		if keyRoot != rootN || !strings.HasPrefix(relativeN, keyRelative) {
			continue
		}
		value := mappedLocations[key]
		overflow := relative[len(keyRelative):]
		if len(overflow) == 0 {
			// exactly match the full path
			return formatRoot(value), ""
		}
		if len(keyRelative) == 0 {
			// only match the root
			return formatRoot(value), strings.Trim(relative, pathSeparator)
		}
		// match several prefix elements
		if overflow[0:1] == pt.env.PathSeparator() {
			return formatRoot(value), strings.Trim(overflow, pathSeparator)
		}
	}
	return formatRoot(root), strings.Trim(relative, pathSeparator)
}

func (pt *Path) normalizePath(path string) string {
	if pt.env.GOOS() != environment.WINDOWS {
		return path
	}
	var clean []rune
	for _, char := range path {
		var lastChar rune
		if len(clean) > 0 {
			lastChar = clean[len(clean)-1:][0]
		}
		if char == '/' && lastChar != 60 { // 60 == <, this is done to ovoid replacing color codes
			clean = append(clean, 92) // 92 == \
			continue
		}
		clean = append(clean, char)
	}
	return string(clean)
}

// ParsePath parses an input path and returns a clean root and a clean path.
func (pt *Path) parsePath(inputPath string) (root, path string) {
	if len(inputPath) == 0 {
		return
	}
	separator := pt.env.PathSeparator()
	clean := func(path string) string {
		matches := regex.FindAllNamedRegexMatch(fmt.Sprintf(`(?P<element>[^\%s]+)`, separator), path)
		n := len(matches) - 1
		s := new(strings.Builder)
		for i, m := range matches {
			s.WriteString(m["element"])
			if i != n {
				s.WriteString(separator)
			}
		}
		return s.String()
	}

	if pt.env.GOOS() == environment.WINDOWS {
		inputPath = pt.normalizePath(inputPath)
		// for a UNC path, extract \\hostname\sharename as the root
		matches := regex.FindNamedRegexMatch(`^\\\\(?P<hostname>[^\\]+)\\+(?P<sharename>[^\\]+)\\*(?P<path>[\s\S]*)$`, inputPath)
		if len(matches) > 0 {
			root = `\\` + matches["hostname"] + `\` + matches["sharename"] + `\`
			path = clean(matches["path"])
			return
		}
	}
	s := strings.SplitAfterN(inputPath, separator, 2)
	root = s[0]
	if !strings.HasSuffix(root, separator) {
		// a root should end with a separator
		root += separator
	}
	if len(s) == 2 {
		path = clean(s[1])
	}
	return root, path
}

func (pt *Path) normalize(inputPath string) string {
	normalized := inputPath
	if strings.HasPrefix(normalized, "~") && (len(normalized) == 1 || environment.IsPathSeparator(pt.env, normalized[1])) {
		normalized = pt.env.Home() + normalized[1:]
	}
	switch pt.env.GOOS() {
	case environment.WINDOWS:
		normalized = pt.normalizePath(normalized)
		fallthrough
	case environment.DARWIN:
		normalized = strings.ToLower(normalized)
	}
	return normalized
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
