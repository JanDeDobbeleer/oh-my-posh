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

type Path struct {
	props           properties.Properties
	env             runtime.Environment
	mappedLocations map[string]string
	root            string
	relative        string
	pwd             string
	Location        string
	pathSeparator   string
	Path            string
	Folders         Folders
	StackCount      int
	windowsPath     bool
	Writable        bool
	RootDir         bool
	cygPath         bool
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
	pt.root, pt.relative = pt.replaceMappedLocations(pt.pwd)
	pt.pwd = pt.join(pt.root, pt.relative)
}

func (pt *Path) Parent() string {
	if len(pt.pwd) == 0 {
		return ""
	}

	folders := pt.Folders.List()
	if len(folders) == 0 {
		// No parent.
		return ""
	}

	sb := new(strings.Builder)
	folderSeparator := pt.getFolderSeparator()

	sb.WriteString(pt.root)
	if !pt.endWithSeparator(pt.root) {
		sb.WriteString(folderSeparator)
	}
	for _, folder := range folders[:len(folders)-1] {
		sb.WriteString(folder)
		sb.WriteString(folderSeparator)
	}
	return sb.String()
}

func (pt *Path) Init(props properties.Properties, env runtime.Environment) {
	pt.props = props
	pt.env = env
}

func (pt *Path) setStyle() {
	if len(pt.relative) == 0 {
		root := pt.root

		// Only append a separator to a non-filesystem PSDrive root or a Windows drive root.
		if (len(pt.env.Flags().PSWD) != 0 || pt.windowsPath) && strings.HasSuffix(root, ":") {
			root += pt.getFolderSeparator()
		}

		pt.Path = pt.colorizePath(root, nil)
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
		Env:      pt.env,
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
		Env:      pt.env,
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
	root := pt.root
	folders := pt.Folders
	threshold := int(pt.props.GetFloat64(MixedThreshold, 4))
	folderIcon := pt.props.GetString(FolderIcon, "..")

	if pt.isRootFS(root) {
		root = folders[0].Name
		folders = folders[1:]
	}

	var elements []string

	for i, n := 0, len(folders); i < n; i++ {
		folderName := folders[i].Name
		if len(folderName) > threshold && i != n-1 && !folders[i].Display {
			elements = append(elements, folderIcon)
			continue
		}

		elements = append(elements, folderName)
	}

	return pt.colorizePath(root, elements)
}

func (pt *Path) getAgnosterPath() string {
	root := pt.root
	folders := pt.Folders
	folderIcon := pt.props.GetString(FolderIcon, "..")

	if pt.isRootFS(root) {
		root = folders[0].Name
		folders = folders[1:]
	}

	var elements []string

	for i, n := 0, len(folders); i < n; i++ {
		if folders[i].Display || i == n-1 {
			elements = append(elements, folders[i].Name)
			continue
		}

		elements = append(elements, folderIcon)
	}

	return pt.colorizePath(root, elements)
}

func (pt *Path) getAgnosterLeftPath() string {
	root := pt.root
	folders := pt.Folders
	folderIcon := pt.props.GetString(FolderIcon, "..")

	if pt.isRootFS(root) {
		root = folders[0].Name
		folders = folders[1:]
	}

	var elements []string
	elements = append(elements, folders[0].Name)
	for i, n := 1, len(folders); i < n; i++ {
		if folders[i].Display {
			elements = append(elements, folders[i].Name)
			continue
		}

		elements = append(elements, folderIcon)
	}

	return pt.colorizePath(root, elements)
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
	root := pt.root
	folders := pt.Folders

	if pt.isRootFS(root) {
		root = folders[0].Name
		folders = folders[1:]
	}

	root = pt.getRelevantLetter(&Folder{Name: root})

	var elements []string
	for i, n := 0, len(folders); i < n; i++ {
		if folders[i].Display || i == n-1 {
			elements = append(elements, folders[i].Name)
			continue
		}

		letter := pt.getRelevantLetter(folders[i])
		elements = append(elements, letter)
	}

	return pt.colorizePath(root, elements)
}

func (pt *Path) getUniqueLettersPath(maxWidth int) string {
	root := pt.root
	folders := pt.Folders
	separator := pt.getFolderSeparator()

	if pt.isRootFS(root) {
		root = folders[0].Name
		folders = folders[1:]
	}

	folderNames := folders.List()

	usePowerlevelStyle := func(root, relative string) bool {
		length := len(root) + len(relative)
		if !pt.endWithSeparator(root) {
			length += len(separator)
		}
		return length <= maxWidth
	}

	if maxWidth > 0 {
		relative := strings.Join(folderNames, separator)
		if usePowerlevelStyle(root, relative) {
			return pt.colorizePath(root, folderNames)
		}
	}

	root = pt.getRelevantLetter(&Folder{Name: root})

	var elements []string
	letters := make(map[string]bool)
	letters[root] = true

	for i, n := 0, len(folders); i < n; i++ {
		folderName := folderNames[i]

		if i == n-1 {
			elements = append(elements, folderName)
			break
		}

		letter := pt.getRelevantLetter(folders[i])

		for letters[letter] {
			if letter == folderName {
				break
			}
			letter += folderName[len(letter) : len(letter)+1]
		}

		letters[letter] = true
		elements = append(elements, letter)

		// only return early on maxWidth > 0
		// this enables the powerlevel10k behavior
		if maxWidth > 0 {
			list := elements
			list = append(list, folderNames[i+1:]...)
			relative := strings.Join(list, separator)
			if usePowerlevelStyle(root, relative) {
				return pt.colorizePath(root, list)
			}
		}
	}

	return pt.colorizePath(root, elements)
}

func (pt *Path) getAgnosterFullPath() string {
	root := pt.root
	folders := pt.Folders

	if pt.isRootFS(root) {
		root = folders[0].Name
		folders = folders[1:]
	}

	return pt.colorizePath(root, folders.List())
}

func (pt *Path) getAgnosterShortPath() string {
	root := pt.root
	folders := pt.Folders

	if pt.isRootFS(root) {
		root = folders[0].Name
		folders = folders[1:]
	}

	maxDepth := pt.props.GetInt(MaxDepth, 1)
	if maxDepth < 1 {
		maxDepth = 1
	}

	pathDepth := len(folders)
	hideRootLocation := pt.props.GetBool(HideRootLocation, false)
	folderIcon := pt.props.GetString(FolderIcon, "..")

	// No need to shorten.
	if pathDepth < maxDepth || (pathDepth == maxDepth && !hideRootLocation) {
		return pt.getAgnosterFullPath()
	}

	elements := []string{folderIcon}

	for i := pathDepth - maxDepth; i < pathDepth; i++ {
		elements = append(elements, folders[i].Name)
	}

	if hideRootLocation {
		return pt.colorizePath(elements[0], elements[1:])
	}

	return pt.colorizePath(root, elements)
}

func (pt *Path) getFullPath() string {
	return pt.colorizePath(pt.root, pt.Folders.List())
}

func (pt *Path) getFolderPath() string {
	folderName := pt.Folders[len(pt.Folders)-1].Name
	return pt.colorizePath(folderName, nil)
}

func (pt *Path) join(root, relative string) string {
	// this is a full replacement of the parent
	if len(root) == 0 {
		return relative
	}

	if !pt.endWithSeparator(root) && len(relative) > 0 {
		return root + pt.pathSeparator + relative
	}

	return root + relative
}

func (pt *Path) setMappedLocations() {
	if pt.mappedLocations != nil {
		return
	}

	mappedLocations := make(map[string]string)

	// predefined mapped locations, can be disabled
	if pt.props.GetBool(MappedLocationsEnabled, true) {
		mappedLocations["hkcu:"] = pt.props.GetString(WindowsRegistryIcon, "\uF013")
		mappedLocations["hklm:"] = pt.props.GetString(WindowsRegistryIcon, "\uF013")
		mappedLocations[pt.normalize(pt.env.Home())] = pt.props.GetString(HomeIcon, "~")
	}

	// merge custom locations with mapped locations
	// mapped locations can override predefined locations
	keyValues := pt.props.GetKeyValueMap(MappedLocations, make(map[string]string))
	for key, value := range keyValues {
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

		// When two templates resolve to the same key, the values are compared in ascending order and the latter is taken.
		if v, exist := mappedLocations[pt.normalize(path)]; exist && value <= v {
			continue
		}

		mappedLocations[pt.normalize(path)] = value
	}

	pt.mappedLocations = mappedLocations
}

func (pt *Path) replaceMappedLocations(inputPath string) (string, string) {
	root, relative := pt.parsePath(inputPath)
	if len(relative) == 0 {
		pt.RootDir = true
	}

	pt.setMappedLocations()
	if len(pt.mappedLocations) == 0 {
		return root, relative
	}

	// sort map keys in reverse order
	// fixes case when a subfoder and its parent are mapped
	// ex /users/test and /users/test/dev
	keys := make([]string, 0, len(pt.mappedLocations))
	for k := range pt.mappedLocations {
		keys = append(keys, k)
	}
	sort.Sort(sort.Reverse(sort.StringSlice(keys)))

	rootN := pt.normalize(root)
	relativeN := pt.normalize(relative)

	escape := func(path string) string {
		// Escape chevron characters to avoid applying unexpected text styles.
		return strings.NewReplacer("<", "<<>", ">", "<>>").Replace(path)
	}

	for _, key := range keys {
		keyRoot, keyRelative := pt.parsePath(key)
		matchSubFolders := strings.HasSuffix(keyRelative, pt.pathSeparator+"*")

		if matchSubFolders {
			// Remove the trailing wildcard (*).
			keyRelative = keyRelative[:len(keyRelative)-1]
		}

		if keyRoot != rootN || !strings.HasPrefix(relativeN, keyRelative) {
			continue
		}

		value := pt.mappedLocations[key]
		overflow := relative[len(keyRelative):]

		// exactly match the full path
		if len(overflow) == 0 {
			return value, ""
		}

		// only match the root
		if len(keyRelative) == 0 {
			return value, strings.Trim(escape(relative), pt.pathSeparator)
		}

		// match several prefix elements
		if matchSubFolders || overflow[:1] == pt.pathSeparator {
			return value, strings.Trim(escape(overflow), pt.pathSeparator)
		}
	}

	return escape(root), strings.Trim(escape(relative), pt.pathSeparator)
}

// parsePath parses a clean input path into a root and a relative.
func (pt *Path) parsePath(inputPath string) (string, string) {
	var root, relative string

	if len(inputPath) == 0 {
		return root, relative
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

	if pt.env.GOOS() == runtime.WINDOWS {
		// Handle a UNC path, if any.
		pattern := fmt.Sprintf(`^\%[1]s{2}(?P<hostname>[^\%[1]s]+)\%[1]s(?P<sharename>[^\%[1]s]+)(\%[1]s(?P<path>[\s\S]*))?$`, pt.pathSeparator)
		matches := regex.FindNamedRegexMatch(pattern, inputPath)
		if len(matches) > 0 {
			root = fmt.Sprintf(`%[1]s%[1]s%[2]s%[1]s%[3]s`, pt.pathSeparator, matches["hostname"], matches["sharename"])
			relative = matches["path"]
			return root, relative
		}
	}

	s := strings.SplitAfterN(inputPath, pt.pathSeparator, 2)
	root = s[0]

	if len(s) == 2 {
		if len(root) > 1 {
			root = root[:len(root)-1]
		}

		relative = s[1]
	}

	return root, relative
}

func (pt *Path) isRootFS(path string) bool {
	return len(path) == 1 && runtime.IsPathSeparator(pt.env, path[0])
}

func (pt *Path) endWithSeparator(path string) bool {
	if len(path) == 0 {
		return false
	}
	return runtime.IsPathSeparator(pt.env, path[len(path)-1])
}

func (pt *Path) normalize(inputPath string) string {
	normalized := inputPath

	if strings.HasPrefix(normalized, "~") && (len(normalized) == 1 || runtime.IsPathSeparator(pt.env, normalized[1])) {
		normalized = pt.env.Home() + normalized[1:]
	}

	normalized = runtime.CleanPath(pt.env, normalized)

	if pt.env.GOOS() == runtime.WINDOWS || pt.env.GOOS() == runtime.DARWIN {
		normalized = strings.ToLower(normalized)
	}

	return normalized
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
		formattedRoot := fmt.Sprintf(leftFormat, root)
		return colorizeElement(formattedRoot)
	}

	colorizeSeparator := func() string {
		if skipColorize || !colorSeparator {
			return folderSeparator
		}
		return fmt.Sprintf("<%s>%s</>", cycle[0], folderSeparator)
	}

	sb := new(strings.Builder)

	formattedRoot := fmt.Sprintf(leftFormat, root)
	sb.WriteString(colorizeElement(formattedRoot))

	if !pt.endWithSeparator(root) {
		sb.WriteString(colorizeSeparator())
	}

	for i, element := range elements {
		if len(element) == 0 {
			continue
		}

		format := folderFormat
		if i == len(elements)-1 {
			format = rightFormat
		}

		formattedElement := fmt.Sprintf(format, element)
		sb.WriteString(colorizeElement(formattedElement))
		if i != len(elements)-1 {
			sb.WriteString(colorizeSeparator())
		}
	}

	return sb.String()
}

func (pt *Path) splitPath() Folders {
	folders := Folders{}

	if len(pt.relative) == 0 {
		return folders
	}

	elements := strings.Split(pt.relative, pt.pathSeparator)
	folderFormatMap := pt.makeFolderFormatMap()
	currentPath := pt.root

	if !pt.endWithSeparator(pt.root) {
		currentPath += pt.pathSeparator
	}

	var display bool

	for _, element := range elements {
		currentPath += element

		if format := folderFormatMap[currentPath]; len(format) != 0 {
			element = fmt.Sprintf(format, element)
			display = true
		}

		folders = append(folders, &Folder{Name: element, Path: currentPath, Display: display})

		currentPath += pt.pathSeparator

		display = false
	}

	return folders
}

func (pt *Path) makeFolderFormatMap() map[string]string {
	folderFormatMap := make(map[string]string)

	if gitDirFormat := pt.props.GetString(GitDirFormat, ""); len(gitDirFormat) != 0 {
		dir, err := pt.env.HasParentFilePath(".git", false)
		if err == nil && dir.IsDir {
			// Make it consistent with the modified path.
			path := pt.join(pt.replaceMappedLocations(dir.ParentFolder))
			folderFormatMap[path] = gitDirFormat
		}
	}

	return folderFormatMap
}
