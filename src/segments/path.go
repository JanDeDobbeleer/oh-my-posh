package segments

import (
	"fmt"
	"sort"
	"strings"

	"github.com/jandedobbeleer/oh-my-posh/src/platform"
	"github.com/jandedobbeleer/oh-my-posh/src/properties"
	"github.com/jandedobbeleer/oh-my-posh/src/regex"
	"github.com/jandedobbeleer/oh-my-posh/src/shell"
	"github.com/jandedobbeleer/oh-my-posh/src/template"
)

type Path struct {
	props properties.Properties
	env   platform.Environment

	root     string
	relative string
	pwd      string

	Path       string
	StackCount int
	Location   string
	Writable   bool
	RootDir    bool
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
	// Powerlevel tries to mimic the powerlevel10k path,
	// used in combination with max_width.
	Powerlevel string = "powerlevel"
	// MixedThreshold the threshold of the length of the path Mixed will display
	MixedThreshold properties.Property = "mixed_threshold"
	// MappedLocations allows overriding certain location with an icon
	MappedLocations properties.Property = "mapped_locations"
	// MappedLocationsEnabled enables overriding certain locations with an icon
	MappedLocationsEnabled properties.Property = "mapped_locations_enabled"
	// MaxDepth Maximum path depth to display whithout shortening
	MaxDepth properties.Property = "max_depth"
	// MaxWidth Maximum path width to display for powerlevel style
	MaxWidth properties.Property = "max_width"
	// Hides the root location if it doesn't fit in max_depth. Used in Agnoster Short
	HideRootLocation properties.Property = "hide_root_location"
	// A color override cycle
	Cycle properties.Property = "cycle"
	// Color the path separators within the cycle
	CycleFolderSeparator properties.Property = "cycle_folder_separator"
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
	pwd := pt.env.Pwd()
	if pt.env.IsWsl() {
		pt.Location, _ = pt.env.RunCommand("wslpath", "-m", pwd)
	} else {
		pt.Location = pwd
	}
	pt.StackCount = pt.env.StackCount()
	pt.Writable = pt.env.DirIsWritable(pwd)
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
	// this is a full replacement of the parent
	if len(pt.root) == 0 {
		pt.pwd = pt.relative
		return
	}
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
	base := platform.Base(pt.env, pt.pwd)
	path := pt.replaceFolderSeparators(pt.pwd[:len(pt.pwd)-len(base)])
	return path
}

func (pt *Path) Init(props properties.Properties, env platform.Environment) {
	pt.props = props
	pt.env = env
}

func (pt *Path) setStyle() {
	if len(pt.relative) == 0 {
		pt.Path = pt.root
		if strings.HasSuffix(pt.Path, ":") {
			pt.Path += pt.env.PathSeparator()
		}
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
		pt.Path = pt.getUniqueLettersPath(0)
	case AgnosterLeft:
		pt.Path = pt.getAgnosterLeftPath()
	case Short:
		// "short" is a duplicate of "full", just here for backwards compatibility
		fallthrough
	case Full:
		pt.Path = pt.getFullPath()
	case Folder:
		pt.Path = pt.getFolderPath()
	case Powerlevel:
		maxWidth := int(pt.props.GetFloat64(MaxWidth, 0))
		pt.Path = pt.getUniqueLettersPath(maxWidth)
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
		pt.env.Error(err)
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
	folderIcon := pt.props.GetString(FolderIcon, "..")
	splitted := strings.Split(pt.relative, pt.env.PathSeparator())
	if pt.root == pt.env.PathSeparator() {
		pt.root = splitted[0]
		splitted = splitted[1:]
	}

	var elements []string
	n := len(splitted)
	for i := 1; i < n; i++ {
		elements = append(elements, folderIcon)
	}
	elements = append(elements, splitted[n-1])

	return pt.colorizePath(pt.root, elements)
}

func (pt *Path) getAgnosterLeftPath() string {
	folderIcon := pt.props.GetString(FolderIcon, "..")
	splitted := strings.Split(pt.relative, pt.env.PathSeparator())
	if pt.root == pt.env.PathSeparator() {
		pt.root = splitted[0]
		splitted = splitted[1:]
	}

	var elements []string
	n := len(splitted)
	elements = append(elements, splitted[0])
	for i := 1; i < n; i++ {
		elements = append(elements, folderIcon)
	}

	return pt.colorizePath(pt.root, elements)
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
	splitted := strings.Split(pt.relative, pt.env.PathSeparator())
	if pt.root == pt.env.PathSeparator() {
		pt.root = splitted[0]
		splitted = splitted[1:]
	}
	pt.root = pt.getRelevantLetter(pt.root)

	var elements []string
	n := len(splitted)
	for i := 0; i < n-1; i++ {
		letter := pt.getRelevantLetter(splitted[i])
		elements = append(elements, letter)
	}
	elements = append(elements, splitted[n-1])

	return pt.colorizePath(pt.root, elements)
}

func (pt *Path) getUniqueLettersPath(maxWidth int) string {
	separator := pt.getFolderSeparator()
	splitted := strings.Split(pt.relative, pt.env.PathSeparator())
	if pt.root == pt.env.PathSeparator() {
		pt.root = splitted[0]
		splitted = splitted[1:]
	}

	if maxWidth > 0 {
		path := strings.Join(splitted, separator)
		if len(path) <= maxWidth {
			return pt.colorizePath(pt.root, splitted)
		}
	}

	pt.root = pt.getRelevantLetter(pt.root)

	var elements []string
	n := len(splitted)
	letters := make(map[string]bool)
	letters[pt.root] = true
	for i := 0; i < n-1; i++ {
		folder := splitted[i]
		letter := pt.getRelevantLetter(folder)
		for letters[letter] {
			if letter == folder {
				break
			}
			letter += folder[len(letter) : len(letter)+1]
		}
		letters[letter] = true
		elements = append(elements, letter)
		// only return early on maxWidth > 0
		// this enables the powerlevel10k behavior
		if maxWidth > 0 {
			list := splitted[i+1:]
			list = append(list, elements...)
			current := strings.Join(list, separator)
			leftover := maxWidth - len(current) - len(pt.root) - len(separator)
			if leftover >= 0 {
				elements = append(elements, strings.Join(splitted[i+1:], separator))
				return pt.colorizePath(pt.root, elements)
			}
		}
	}
	elements = append(elements, splitted[n-1])

	return pt.colorizePath(pt.root, elements)
}

func (pt *Path) getAgnosterFullPath() string {
	splitted := strings.Split(pt.relative, pt.env.PathSeparator())
	if pt.root == pt.env.PathSeparator() {
		pt.root = splitted[0]
		splitted = splitted[1:]
	}

	return pt.colorizePath(pt.root, splitted)
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
	rel := strings.TrimPrefix(pt.relative, pathSeparator)
	splitted := strings.Split(rel, pathSeparator)
	splitPos := pathDepth - maxDepth
	// var buffer strings.Builder
	var elements []string
	// unix root, needs to be replaced with the folder we're in at root level
	root := pt.root
	room := pathDepth - maxDepth
	if root == pathSeparator {
		root = splitted[0]
		room--
	}

	if hideRootLocation || room > 0 {
		elements = append(elements, folderIcon)
	}

	if hideRootLocation {
		root = ""
	}

	for i := splitPos; i < pathDepth; i++ {
		elements = append(elements, splitted[i])
	}
	return pt.colorizePath(root, elements)
}

func (pt *Path) getFullPath() string {
	elements := strings.Split(pt.relative, pt.env.PathSeparator())
	return pt.colorizePath(pt.root, elements)
}

func (pt *Path) getFolderPath() string {
	pwd := platform.Base(pt.env, pt.pwd)
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
		if len(key) == 0 {
			continue
		}
		tmpl := &template.Text{
			Template: key,
			Context:  pt,
			Env:      pt.env,
		}
		path, err := tmpl.Render()
		if err != nil {
			pt.env.Error(err)
		}
		if len(path) == 0 {
			continue
		}
		mappedLocations[pt.normalize(path)] = val
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
	if len(relative) == 0 {
		pt.RootDir = true
	}
	rootN := pt.normalize(root)
	relativeN := pt.normalize(relative)
	pathSeparator := pt.env.PathSeparator()

	for _, key := range keys {
		keyRoot, keyRelative := pt.parsePath(key)
		matchSubFolders := strings.HasSuffix(keyRelative, "*")
		if matchSubFolders && len(keyRelative) > 1 {
			keyRelative = keyRelative[0 : len(keyRelative)-1] // remove trailing /* or \*
		}
		if keyRoot != rootN || !strings.HasPrefix(relativeN, keyRelative) {
			continue
		}
		value := mappedLocations[key]
		overflow := relative[len(keyRelative):]
		if len(overflow) == 0 {
			// exactly match the full path
			return value, ""
		}
		if len(keyRelative) == 0 {
			// only match the root
			return value, strings.Trim(relative, pathSeparator)
		}
		// match several prefix elements
		if matchSubFolders || overflow[0:1] == pt.env.PathSeparator() {
			return value, strings.Trim(overflow, pathSeparator)
		}
	}
	return root, strings.Trim(relative, pathSeparator)
}

func (pt *Path) normalizePath(path string) string {
	if pt.env.GOOS() != platform.WINDOWS {
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

	if pt.env.GOOS() == platform.WINDOWS {
		inputPath = pt.normalizePath(inputPath)
		// for a UNC path, extract \\hostname\sharename as the root
		matches := regex.FindNamedRegexMatch(`^\\\\(?P<hostname>[^\\]+)\\+(?P<sharename>[^\\]+)\\*(?P<path>[\s\S]*)$`, inputPath)
		if len(matches) > 0 {
			root = `\\` + matches["hostname"] + `\` + matches["sharename"]
			path = clean(matches["path"])
			return
		}
	}
	s := strings.SplitAfterN(inputPath, separator, 2)
	root = s[0]
	if pt.env.GOOS() == platform.WINDOWS {
		root = strings.TrimSuffix(root, separator)
	}
	if len(s) == 2 {
		path = clean(s[1])
	}
	return root, path
}

func (pt *Path) normalize(inputPath string) string {
	normalized := inputPath
	if strings.HasPrefix(normalized, "~") && (len(normalized) == 1 || platform.IsPathSeparator(pt.env, normalized[1])) {
		normalized = pt.env.Home() + normalized[1:]
	}
	switch pt.env.GOOS() {
	case platform.WINDOWS:
		normalized = pt.normalizePath(normalized)
		fallthrough
	case platform.DARWIN:
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

func (pt *Path) colorizePath(root string, elements []string) string {
	cycle := pt.props.GetStringArray(Cycle, []string{})
	skipColorize := len(cycle) == 0
	folderSeparator := pt.getFolderSeparator()
	colorSeparator := pt.props.GetBool(CycleFolderSeparator, false)

	colorizeElement := func(element string) string {
		if skipColorize || len(element) == 0 {
			return element
		}
		defer func() {
			cycle = append(cycle[1:], cycle[0])
		}()
		return fmt.Sprintf("<%s>%s</>", cycle[0], element)
	}

	colorizeSeparator := func() string {
		if skipColorize || !colorSeparator {
			return folderSeparator
		}
		return fmt.Sprintf("<%s>%s</>", cycle[0], folderSeparator)
	}

	var builder strings.Builder

	builder.WriteString(colorizeElement(root))

	if root != pt.env.PathSeparator() && len(root) != 0 {
		builder.WriteString(colorizeSeparator())
	}

	for i, element := range elements {
		if len(element) == 0 {
			continue
		}
		builder.WriteString(colorizeElement(element))
		if i != len(elements)-1 {
			builder.WriteString(colorizeSeparator())
		}
	}

	return builder.String()
}
