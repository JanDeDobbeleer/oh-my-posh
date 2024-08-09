package segments

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/jandedobbeleer/oh-my-posh/src/properties"
	"github.com/jandedobbeleer/oh-my-posh/src/regex"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime"
	"github.com/jandedobbeleer/oh-my-posh/src/shell"
	"github.com/jandedobbeleer/oh-my-posh/src/template"
)

type Path struct {
	props         properties.Properties
	env           runtime.Environment
	root          string
	relative      string
	pwd           string
	Location      string
	pathSeparator string
	Path          string
	Folders       Folders
	StackCount    int
	windowsPath   bool
	Writable      bool
	RootDir       bool
	cygPath       bool
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
	// FolderType displays the current folder
	FolderType string = "folder"
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
	// format to use on the folder names
	FolderFormat properties.Property = "folder_format"
	// format to use on the first and last folder of the path
	EdgeFormat properties.Property = "edge_format"
	// format to use on first folder of the path
	LeftFormat properties.Property = "left_format"
	// format to use on the last folder of the path
	RightFormat properties.Property = "right_format"
	// GitDirFormat format to use on the git directory
	GitDirFormat properties.Property = "gitdir_format"
	// DisplayCygpath transforms the path to a cygpath format
	DisplayCygpath properties.Property = "display_cygpath"
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

	pt.Location = pt.env.TemplateCache().AbsolutePWD
	if pt.env.GOOS() == runtime.WINDOWS {
		pt.Location = strings.ReplaceAll(pt.Location, `\`, `/`)
	}

	pt.StackCount = pt.env.StackCount()
	pt.Writable = pt.env.DirIsWritable(pwd)
	return true
}

func (pt *Path) setPaths() {
	defer func() {
		pt.Folders = pt.splitPath()
	}()

	displayCygpath := func() bool {
		enableCygpath := pt.props.GetBool(DisplayCygpath, false)
		if !enableCygpath {
			return false
		}

		return pt.env.IsCygwin()
	}

	pt.cygPath = displayCygpath()
	pt.windowsPath = pt.env.GOOS() == runtime.WINDOWS && !pt.cygPath
	pt.pathSeparator = pt.env.PathSeparator()

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

	if !strings.HasSuffix(pt.root, pt.pathSeparator) && len(pt.relative) > 0 {
		pt.pwd = pt.root + pt.pathSeparator + pt.relative
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
	base := runtime.Base(pt.env, pt.pwd)
	path := pt.replaceFolderSeparators(pt.pwd[:len(pt.pwd)-len(base)])
	return path
}

func (pt *Path) Init(props properties.Properties, env runtime.Environment) {
	pt.props = props
	pt.env = env
}

func (pt *Path) setStyle() {
	if len(pt.relative) == 0 {
		// Only append a separator to a non-filesystem PSDrive root or a Windows drive root.
		if (len(pt.env.Flags().PSWD) != 0 || pt.windowsPath) && strings.HasSuffix(pt.root, ":") {
			pt.root += pt.getFolderSeparator()
		}

		pt.Path = pt.colorizePath(pt.root, nil)
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
	case Full, Short: // "short" is a duplicate of "full", just here for backwards compatibility
		pt.Path = pt.getFullPath()
	case FolderType:
		pt.Path = pt.getFolderPath()
	case Powerlevel:
		maxWidth := pt.getMaxWidth()
		pt.Path = pt.getUniqueLettersPath(maxWidth)
	default:
		pt.Path = fmt.Sprintf("Path style: %s is not available", style)
	}
}

func (pt *Path) getMaxWidth() int {
	width := pt.props.GetString(MaxWidth, "")
	if len(width) == 0 {
		return 0
	}

	tmpl := &template.Text{
		Template: width,
		Context:  pt,
	}

	text, err := tmpl.Render()
	if err != nil {
		pt.env.Error(err)
		return 0
	}

	value, err := strconv.Atoi(text)
	if err != nil {
		pt.env.Error(err)
		return 0
	}

	return value
}

func (pt *Path) getFolderSeparator() string {
	separatorTemplate := pt.props.GetString(FolderSeparatorTemplate, "")
	if len(separatorTemplate) == 0 {
		separator := pt.props.GetString(FolderSeparatorIcon, pt.pathSeparator)
		// if empty, use the default separator
		if len(separator) == 0 {
			return pt.pathSeparator
		}
		return separator
	}

	tmpl := &template.Text{
		Template: separatorTemplate,
		Context:  pt,
	}

	text, err := tmpl.Render()
	if err != nil {
		pt.env.Error(err)
	}

	if len(text) == 0 {
		return pt.pathSeparator
	}

	return text
}

func (pt *Path) getMixedPath() string {
	threshold := int(pt.props.GetFloat64(MixedThreshold, 4))
	folderIcon := pt.props.GetString(FolderIcon, "..")

	if pt.root == pt.pathSeparator {
		pt.root = pt.Folders[0].Name
		pt.Folders = pt.Folders[1:]
	}

	var folders []string

	for i, n := 0, len(pt.Folders); i < n; i++ {
		folder := pt.Folders[i].Name
		if len(folder) > threshold && i != n-1 && !pt.Folders[i].Display {
			folder = folderIcon
		}
		folders = append(folders, folder)
	}

	return pt.colorizePath(pt.root, folders)
}

func (pt *Path) getAgnosterPath() string {
	folderIcon := pt.props.GetString(FolderIcon, "..")

	if pt.root == pt.pathSeparator {
		pt.root = pt.Folders[0].Name
		pt.Folders = pt.Folders[1:]
	}

	var elements []string
	n := len(pt.Folders)
	for i := 0; i < n-1; i++ {
		name := folderIcon

		if pt.Folders[i].Display {
			name = pt.Folders[i].Name
		}

		elements = append(elements, name)
	}

	if len(pt.Folders) > 0 {
		elements = append(elements, pt.Folders[n-1].Name)
	}

	return pt.colorizePath(pt.root, elements)
}

func (pt *Path) getAgnosterLeftPath() string {
	folderIcon := pt.props.GetString(FolderIcon, "..")

	if pt.root == pt.pathSeparator {
		pt.root = pt.Folders[0].Name
		pt.Folders = pt.Folders[1:]
	}

	var elements []string
	n := len(pt.Folders)
	elements = append(elements, pt.Folders[0].Name)
	for i := 1; i < n; i++ {
		if pt.Folders[i].Display {
			elements = append(elements, pt.Folders[i].Name)
			continue
		}

		elements = append(elements, folderIcon)
	}

	return pt.colorizePath(pt.root, elements)
}

func (pt *Path) getRelevantLetter(folder *Folder) string {
	if folder.Display {
		return folder.Name
	}

	// check if there is at least a letter we can use
	matches := regex.FindNamedRegexMatch(`(?P<letter>[\p{L}0-9]).*`, folder.Name)
	if matches == nil || len(matches["letter"]) == 0 {
		// no letter found, keep the folder unchanged
		return folder.Name
	}
	letter := matches["letter"]
	// handle non-letter characters before the first found letter
	letter = folder.Name[0:strings.Index(folder.Name, letter)] + letter
	return letter
}

func (pt *Path) getLetterPath() string {
	if pt.root == pt.pathSeparator {
		pt.root = pt.Folders[0].Name
		pt.Folders = pt.Folders[1:]
	}

	pt.root = pt.getRelevantLetter(&Folder{Name: pt.root})

	var elements []string
	n := len(pt.Folders)
	for i := 0; i < n-1; i++ {
		if pt.Folders[i].Display {
			elements = append(elements, pt.Folders[i].Name)
			continue
		}

		letter := pt.getRelevantLetter(pt.Folders[i])
		elements = append(elements, letter)
	}

	if len(pt.Folders) > 0 {
		elements = append(elements, pt.Folders[n-1].Name)
	}

	return pt.colorizePath(pt.root, elements)
}

func (pt *Path) getUniqueLettersPath(maxWidth int) string {
	separator := pt.getFolderSeparator()

	if pt.root == pt.pathSeparator {
		pt.root = pt.Folders[0].Name
		pt.Folders = pt.Folders[1:]
	}

	if maxWidth > 0 {
		path := strings.Join(pt.Folders.List(), separator)
		if len(path) <= maxWidth {
			return pt.colorizePath(pt.root, pt.Folders.List())
		}
	}

	pt.root = pt.getRelevantLetter(&Folder{Name: pt.root})

	var elements []string
	n := len(pt.Folders)
	letters := make(map[string]bool)
	letters[pt.root] = true
	for i := 0; i < n-1; i++ {
		folder := pt.Folders[i].Name
		letter := pt.getRelevantLetter(pt.Folders[i])

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
			list := pt.Folders[i+1:].List()
			list = append(list, elements...)
			current := strings.Join(list, separator)
			leftover := maxWidth - len(current) - len(pt.root) - len(separator)
			if leftover >= 0 {
				elements = append(elements, strings.Join(pt.Folders[i+1:].List(), separator))
				return pt.colorizePath(pt.root, elements)
			}
		}
	}

	if len(pt.Folders) > 0 {
		elements = append(elements, pt.Folders[n-1].Name)
	}

	return pt.colorizePath(pt.root, elements)
}

func (pt *Path) getAgnosterFullPath() string {
	if pt.root == pt.pathSeparator {
		pt.root = pt.Folders[0].Name
		pt.Folders = pt.Folders[1:]
	}

	return pt.colorizePath(pt.root, pt.Folders.List())
}

func (pt *Path) getAgnosterShortPath() string {
	pathDepth := len(pt.Folders)

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

	splitPos := pathDepth - maxDepth

	var folders []string
	// unix root, needs to be replaced with the folder we're in at root level
	root := pt.root
	room := pathDepth - maxDepth
	if root == pt.pathSeparator {
		root = pt.Folders[0].Name
		room--
	}

	if hideRootLocation || room > 0 {
		folders = append(folders, folderIcon)
	}

	if hideRootLocation {
		root = ""
	}

	for i := splitPos; i < pathDepth; i++ {
		folders = append(folders, pt.Folders[i].Name)
	}

	return pt.colorizePath(root, folders)
}

func (pt *Path) getFullPath() string {
	return pt.colorizePath(pt.root, pt.Folders.List())
}

func (pt *Path) getFolderPath() string {
	return pt.colorizePath(runtime.Base(pt.env, pt.pwd), nil)
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
			return value, strings.Trim(relative, pt.pathSeparator)
		}

		// match several prefix elements
		if matchSubFolders || overflow[0:1] == pt.pathSeparator {
			return value, strings.Trim(overflow, pt.pathSeparator)
		}
	}

	return root, strings.Trim(relative, pt.pathSeparator)
}

func (pt *Path) normalizePath(path string) string {
	if pt.env.GOOS() != runtime.WINDOWS || pt.cygPath {
		return path
	}

	var clean []rune
	for _, char := range path {
		var lastChar rune
		if len(clean) > 0 {
			lastChar = clean[len(clean)-1:][0]
		}

		if char == '/' && lastChar != 60 { // 60 == <, this is done to avoid replacing color codes
			clean = append(clean, 92) // 92 == \
			continue
		}

		clean = append(clean, char)
	}

	return string(clean)
}

// ParsePath parses an input path and returns a clean root and a clean path.
func (pt *Path) parsePath(inputPath string) (string, string) {
	var root, path string
	if len(inputPath) == 0 {
		return root, path
	}

	if pt.cygPath {
		path, err := pt.env.RunCommand("cygpath", "-u", inputPath)
		if len(path) != 0 {
			inputPath = path
			pt.pathSeparator = "/"
		}

		if err != nil {
			pt.cygPath = false
			pt.windowsPath = true
		}
	}

	clean := func(path string) string {
		matches := regex.FindAllNamedRegexMatch(fmt.Sprintf(`(?P<element>[^\%s]+)`, pt.pathSeparator), path)
		n := len(matches) - 1
		s := new(strings.Builder)
		for i, m := range matches {
			s.WriteString(m["element"])
			if i != n {
				s.WriteString(pt.pathSeparator)
			}
		}
		return s.String()
	}

	if pt.windowsPath {
		inputPath = pt.normalizePath(inputPath)
		// for a UNC path, extract \\hostname\sharename as the root
		matches := regex.FindNamedRegexMatch(`^\\\\(?P<hostname>[^\\]+)\\+(?P<sharename>[^\\]+)\\*(?P<path>[\s\S]*)$`, inputPath)
		if len(matches) > 0 {
			root = `\\` + matches["hostname"] + `\` + matches["sharename"]
			path = clean(matches["path"])
			return root, path
		}
	}

	s := strings.SplitAfterN(inputPath, pt.pathSeparator, 2)
	root = s[0]

	if pt.windowsPath {
		root = strings.TrimSuffix(root, pt.pathSeparator)
	}

	if len(s) == 2 {
		path = clean(s[1])
	}

	return root, path
}

func (pt *Path) normalize(inputPath string) string {
	normalized := inputPath
	if strings.HasPrefix(normalized, "~") && (len(normalized) == 1 || runtime.IsPathSeparator(pt.env, normalized[1])) {
		normalized = pt.env.Home() + normalized[1:]
	}

	if pt.cygPath {
		return normalized
	}

	switch pt.env.GOOS() {
	case runtime.WINDOWS:
		normalized = pt.normalizePath(normalized)
		fallthrough
	case runtime.DARWIN:
		normalized = strings.ToLower(normalized)
	}

	return normalized
}

func (pt *Path) replaceFolderSeparators(pwd string) string {
	defaultSeparator := pt.pathSeparator
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
	folderFormat := pt.props.GetString(FolderFormat, "%s")

	edgeFormat := pt.props.GetString(EdgeFormat, folderFormat)
	leftFormat := pt.props.GetString(LeftFormat, edgeFormat)
	rightFormat := pt.props.GetString(RightFormat, edgeFormat)

	colorizeElement := func(element string) string {
		if skipColorize || len(element) == 0 {
			return element
		}
		defer func() {
			cycle = append(cycle[1:], cycle[0])
		}()
		return fmt.Sprintf("<%s>%s</>", cycle[0], element)
	}

	if len(elements) == 0 {
		root = fmt.Sprintf(leftFormat, root)
		return colorizeElement(root)
	}

	colorizeSeparator := func() string {
		if skipColorize || !colorSeparator {
			return folderSeparator
		}
		return fmt.Sprintf("<%s>%s</>", cycle[0], folderSeparator)
	}

	var builder strings.Builder

	formattedRoot := fmt.Sprintf(leftFormat, root)
	builder.WriteString(colorizeElement(formattedRoot))

	if root != pt.pathSeparator && len(root) != 0 {
		builder.WriteString(colorizeSeparator())
	}

	for i, element := range elements {
		if len(element) == 0 {
			continue
		}

		format := folderFormat
		if i == len(elements)-1 {
			format = rightFormat
		}

		element = fmt.Sprintf(format, element)
		builder.WriteString(colorizeElement(element))
		if i != len(elements)-1 {
			builder.WriteString(colorizeSeparator())
		}
	}

	return builder.String()
}

type Folder struct {
	Name    string
	Path    string
	Display bool
}

type Folders []*Folder

func (f Folders) List() []string {
	var list []string

	for _, folder := range f {
		list = append(list, folder.Name)
	}

	return list
}

func (pt *Path) splitPath() Folders {
	result := Folders{}
	folders := []string{}

	if len(pt.relative) != 0 {
		folders = strings.Split(pt.relative, pt.pathSeparator)
	}

	folderFormatMap := pt.makeFolderFormatMap()

	getCurrentPath := func() string {
		if pt.root == "~" {
			return pt.env.Home() + pt.pathSeparator
		}

		if pt.windowsPath {
			return pt.root + pt.pathSeparator
		}

		return pt.root
	}

	currentPath := getCurrentPath()

	var display bool

	for _, folder := range folders {
		currentPath += folder

		if format := folderFormatMap[currentPath]; len(format) != 0 {
			folder = fmt.Sprintf(format, folder)
			display = true
		}

		result = append(result, &Folder{Name: folder, Path: currentPath, Display: display})

		currentPath += pt.pathSeparator

		display = false
	}

	return result
}

func (pt *Path) makeFolderFormatMap() map[string]string {
	folderFormatMap := make(map[string]string)

	if gitDirFormat := pt.props.GetString(GitDirFormat, ""); len(gitDirFormat) != 0 {
		dir, err := pt.env.HasParentFilePath(".git", false)
		if err == nil && dir.IsDir {
			folderFormatMap[dir.ParentFolder] = gitDirFormat
		}
	}

	return folderFormatMap
}
