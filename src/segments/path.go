package segments

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/jandedobbeleer/oh-my-posh/src/log"
	"github.com/jandedobbeleer/oh-my-posh/src/regex"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime/path"
	"github.com/jandedobbeleer/oh-my-posh/src/segments/options"
	"github.com/jandedobbeleer/oh-my-posh/src/shell"
	"github.com/jandedobbeleer/oh-my-posh/src/template"
	"github.com/jandedobbeleer/oh-my-posh/src/text"
)

const (
	regexPrefix = "re:"
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

func (f Folders) Last() *Folder {
	return f[len(f)-1]
}

type Path struct {
	Base

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
	FolderSeparatorIcon options.Option = "folder_separator_icon"
	// FolderSeparatorTemplate the path which is split will be separated by this template
	FolderSeparatorTemplate options.Option = "folder_separator_template"
	// HomeIcon indicates the $HOME location
	HomeIcon options.Option = "home_icon"
	// FolderIcon identifies one folder
	FolderIcon options.Option = "folder_icon"
	// WindowsRegistryIcon indicates the registry location on Windows
	WindowsRegistryIcon options.Option = "windows_registry_icon"
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
	MixedThreshold options.Option = "mixed_threshold"
	// MappedLocations allows overriding certain location with an icon
	MappedLocations options.Option = "mapped_locations"
	// MappedLocationsEnabled enables overriding certain locations with an icon
	MappedLocationsEnabled options.Option = "mapped_locations_enabled"
	// MaxDepth Maximum path depth to display without shortening
	MaxDepth options.Option = "max_depth"
	// MaxWidth Maximum path width to display for powerlevel style
	MaxWidth options.Option = "max_width"
	// Hides the root location if it doesn't fit in max_depth. Used in Agnoster Short
	HideRootLocation options.Option = "hide_root_location"
	// A color override cycle
	Cycle options.Option = "cycle"
	// Color the path separators within the cycle
	CycleFolderSeparator options.Option = "cycle_folder_separator"
	// format to use on the folder names
	FolderFormat options.Option = "folder_format"
	// format to use on the first and last folder of the path
	EdgeFormat options.Option = "edge_format"
	// format to use on first folder of the path
	LeftFormat options.Option = "left_format"
	// format to use on the last folder of the path
	RightFormat options.Option = "right_format"
	// GitDirFormat format to use on the git directory
	GitDirFormat options.Option = "gitdir_format"
	// DisplayCygpath transforms the path to a cygpath format
	DisplayCygpath options.Option = "display_cygpath"
	// DisplayRoot indicates if the linux root slash should be displayed
	DisplayRoot options.Option = "display_root"
	// Fish displays the path in a fish-like style
	Fish string = "fish"
	// DirLength the length of the directory name to display in fish style
	DirLength options.Option = "dir_length"
	// FullLengthDirs indicates how many full length directory names should be displayed in fish style
	FullLengthDirs options.Option = "full_length_dirs"
)

func (pt *Path) Template() string {
	return " {{ .Path }} "
}

func (pt *Path) Enabled() bool {
	pt.setPaths()
	if pt.pwd == "" {
		return false
	}

	pt.setStyle()
	pwd := pt.env.Pwd()

	pt.Location = pt.env.Flags().AbsolutePWD
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
		enableCygpath := pt.options.Bool(DisplayCygpath, false)
		if !enableCygpath {
			return false
		}

		return pt.env.IsCygwin()
	}

	pt.cygPath = displayCygpath()
	pt.windowsPath = pt.env.GOOS() == runtime.WINDOWS && !pt.cygPath

	if pt.pathSeparator == "" {
		pt.pathSeparator = path.Separator()
	}

	pt.pwd = pt.env.Pwd()
	if pt.env.Shell() == shell.PWSH && len(pt.env.Flags().PSWD) != 0 {
		pt.pwd = pt.env.Flags().PSWD
	}

	if pt.pwd == "" {
		return
	}

	// ensure a clean path
	pt.root, pt.relative = pt.replaceMappedLocations(pt.pwd)
	pt.pwd = pt.join(pt.root, pt.relative)
}

func (pt *Path) Parent() string {
	if pt.pwd == "" {
		return ""
	}

	folders := pt.Folders.List()
	if len(folders) == 0 {
		// No parent.
		return ""
	}

	sb := text.NewBuilder()

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

func (pt *Path) Format(inputPath string) string {
	separator := path.Separator()

	elements := strings.Split(inputPath, separator)
	if len(elements) == 0 {
		return inputPath
	}

	if len(elements) == 1 {
		return pt.colorizePath(elements[0], nil)
	}

	return pt.colorizePath(elements[0], elements[1:])
}

func (pt *Path) setStyle() {
	if pt.relative == "" {
		root := pt.root

		// Only append a separator to a non-filesystem PSDrive root or a Windows drive root.
		if (len(pt.env.Flags().PSWD) != 0 || pt.windowsPath) && strings.HasSuffix(root, ":") {
			root += pt.getFolderSeparator()
		}

		pt.Path = pt.colorizePath(root, nil)
		return
	}

	switch style := pt.options.String(options.Style, Agnoster); style {
	case Agnoster:
		maxWidth := pt.getMaxWidth()
		pt.Path = pt.getAgnosterPath(maxWidth)
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
	case Fish:
		pt.Path = pt.getFishPath()
	default:
		pt.Path = fmt.Sprintf("Path style: %s is not available", style)
	}

	// make sure we resolve all templates
	if txt, err := template.Render(pt.Path, pt); err == nil {
		pt.Path = txt
	}
}

func (pt *Path) getMaxWidth() int {
	width := pt.options.String(MaxWidth, "")
	if width == "" {
		return 0
	}

	txt, err := template.Render(width, pt)
	if err != nil {
		log.Error(err)
		return 0
	}

	value, err := strconv.Atoi(txt)
	if err != nil {
		log.Error(err)
		return 0
	}

	return value
}

func (pt *Path) getFolderSeparator() string {
	separatorTemplate := pt.options.String(FolderSeparatorTemplate, "")
	if separatorTemplate == "" {
		separator := pt.options.String(FolderSeparatorIcon, pt.pathSeparator)
		// if empty, use the default separator
		if separator == "" {
			return pt.pathSeparator
		}

		return separator
	}

	txt, err := template.Render(separatorTemplate, pt)
	if err != nil {
		log.Error(err)
	}

	if txt == "" {
		return pt.pathSeparator
	}

	return txt
}

func (pt *Path) getMixedPath() string {
	threshold := int(pt.options.Float64(MixedThreshold, 4))
	folderIcon := pt.options.String(FolderIcon, "..")

	root, folders := pt.getPaths()

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

func (pt *Path) getAgnosterPath(maxWidth int) string {
	if maxWidth > 0 {
		return pt.getAgnosterMaxWidth(maxWidth)
	}

	folderIcon := pt.options.String(FolderIcon, "..")

	root, folders := pt.getPaths()

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
	folderIcon := pt.options.String(FolderIcon, "..")

	root, folders := pt.getPaths()

	var elements []string
	if len(folders) == 0 {
		return pt.colorizePath(root, elements)
	}

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

func (pt *Path) findFirstLetterOrNumber(txt string) (letter string, index int) {
	for i, char := range txt {
		if unicode.IsLetter(char) || unicode.IsNumber(char) {
			return string(char), i
		}
	}

	return txt, 0
}

func (pt *Path) getRelevantLetter(folder *Folder) string {
	if folder.Display {
		return folder.Name
	}

	letter, index := pt.findFirstLetterOrNumber(folder.Name)
	if index == 0 {
		return letter
	}

	// handle non-letter characters before the first found letter
	return folder.Name[0:index] + letter
}

func (pt *Path) getLetterPath() string {
	root, folders := pt.getPaths()

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

func (pt *Path) getFishPath() string {
	root, folders := pt.getPaths()
	folders = append(Folders{&Folder{Name: root, Display: false}}, folders...)

	dirLength := pt.options.Int(DirLength, 1)
	fullLengthDirs := max(pt.options.Int(FullLengthDirs, 1), 1)

	folderCount := len(folders)
	stopAt := folderCount - fullLengthDirs

	var elements []string
	for i := range folderCount {
		name := folders[i].Name
		runeCount := utf8.RuneCountInString(name)
		if folders[i].Display || dirLength <= 0 || runeCount < dirLength || i >= stopAt {
			elements = append(elements, name)
			continue
		}

		// Convert string to rune slice to properly handle multi-byte characters
		runes := []rune(name)
		elements = append(elements, string(runes[:dirLength]))
	}

	if len(elements) == 1 {
		return pt.colorizePath(elements[0], nil)
	}

	return pt.colorizePath(elements[0], elements[1:])
}

func (pt *Path) getUniqueLettersPath(maxWidth int) string {
	dr := pt.options.Bool(DisplayRoot, false)
	log.Debugf("%t", dr)
	separator := pt.getFolderSeparator()

	root, folders := pt.getPaths()

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

func (pt *Path) getAgnosterMaxWidth(maxWidth int) string {
	separator := pt.getFolderSeparator()
	folderIcon := pt.options.String(FolderIcon, "..")

	root, folders := pt.getPaths()
	folderNames := append([]string{root}, folders.List()...)

	// this assumes that the root is never a single character
	// except when it really is / on unix systems
	if len(root) == 1 {
		maxWidth++ // add one for the separator
	}

	if len(folderNames) == 0 {
		return pt.colorizePath(root, nil)
	}

	fullPath := strings.Join(folderNames, separator)

	for i := 0; i < len(folderNames)-1 && utf8.RuneCountInString(fullPath) > maxWidth; i++ {
		folderNames[i] = folderIcon
		fullPath = strings.Join(folderNames, separator)
	}

	for len(folderNames) > 1 && utf8.RuneCountInString(fullPath) > maxWidth {
		// remove every folder until the path is short enough
		folderNames = folderNames[1:]
		fullPath = strings.Join(folderNames, separator)
	}

	if len(folderNames) == 1 {
		return pt.colorizePath(template.TruncE(maxWidth, folderNames[0]), nil)
	}

	return pt.colorizePath(folderNames[0], folderNames[1:])
}

func (pt *Path) getAgnosterFullPath() string {
	root, folders := pt.getPaths()

	return pt.colorizePath(root, folders.List())
}

func (pt *Path) getAgnosterShortPath() string {
	root, folders := pt.getPaths()

	maxDepth := max(pt.options.Int(MaxDepth, 1), 1)

	pathDepth := len(folders)
	hideRootLocation := pt.options.Bool(HideRootLocation, false)
	folderIcon := pt.options.String(FolderIcon, "..")

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
	if root == "" {
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
	if pt.options.Bool(MappedLocationsEnabled, true) {
		mappedLocations["hkcu:"] = pt.options.String(WindowsRegistryIcon, "\uF013")
		mappedLocations["hklm:"] = pt.options.String(WindowsRegistryIcon, "\uF013")
		mappedLocations[pt.normalize(pt.env.Home())] = pt.options.String(HomeIcon, "~")
	}

	// merge custom locations with mapped locations
	// mapped locations can override predefined locations
	keyValues := pt.options.KeyValueMap(MappedLocations, make(map[string]string))
	for key, value := range keyValues {
		if key == "" {
			continue
		}

		location, err := template.Render(key, pt)
		if err != nil {
			log.Error(err)
		}

		if location == "" {
			continue
		}

		if !strings.HasPrefix(location, regexPrefix) {
			location = pt.normalize(location)
		}

		// When two templates resolve to the same key, the values are compared in ascending order and the latter is taken.
		if v, exist := mappedLocations[location]; exist && value <= v {
			continue
		}

		mappedLocations[location] = value
	}

	pt.mappedLocations = mappedLocations
}

func (pt *Path) replaceMappedLocations(inputPath string) (string, string) {
	root, relative := pt.parsePath(inputPath)
	if relative == "" {
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

	handleRegex := func(key string) (string, bool) {
		if !strings.HasPrefix(key, regexPrefix) {
			return "", false
		}

		input := strings.ReplaceAll(inputPath, `\`, `/`)
		pattern := key[len(regexPrefix):]

		// Add (?i) at the start of the pattern for case-insensitive matching on Windows
		if pt.windowsPath || (pt.env.IsWsl() && strings.HasPrefix(input, "/mnt/")) {
			pattern = "(?i)" + pattern
		}

		match, OK := regex.FindStringMatch(pattern, input, 1)
		if !OK {
			return "", false
		}

		// Replace the first match with the mapped location.
		input = strings.Replace(input, match, pt.mappedLocations[key], 1)
		input = path.Clean(input)

		return input, true
	}

	for _, key := range keys {
		if input, OK := handleRegex(key); OK {
			return pt.parsePath(input)
		}

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
		if overflow == "" {
			return value, ""
		}

		// only match the root
		if keyRelative == "" {
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

	if inputPath == "" {
		return root, relative
	}

	if pt.cygPath {
		cygPath, err := pt.env.RunCommand("cygpath", "-u", inputPath)
		if len(cygPath) != 0 {
			inputPath = cygPath
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

func (pt *Path) getPaths() (string, Folders) {
	root := pt.root
	folders := pt.Folders

	isRootFS := func(inputPath string) bool {
		displayRoot := pt.options.Bool(DisplayRoot, false)
		if displayRoot {
			return false
		}

		return len(inputPath) == 1 && path.IsSeparator(inputPath[0])
	}

	if isRootFS(root) && len(folders) > 0 {
		root = folders[0].Name
		folders = folders[1:]
	}

	return root, folders
}

func (pt *Path) endWithSeparator(inputPath string) bool {
	if inputPath == "" {
		return false
	}

	return path.IsSeparator(inputPath[len(inputPath)-1])
}

func (pt *Path) normalize(inputPath string) string {
	normalized := inputPath

	if strings.HasPrefix(normalized, "~") && (len(normalized) == 1 || path.IsSeparator(normalized[1])) {
		normalized = pt.env.Home() + normalized[1:]
	}

	normalized = path.Clean(normalized)

	if pt.env.GOOS() == runtime.WINDOWS || pt.env.GOOS() == runtime.DARWIN {
		normalized = strings.ToLower(normalized)
	}

	if pt.cygPath {
		return strings.ReplaceAll(normalized, `\`, "/")
	}

	return normalized
}

func (pt *Path) colorizePath(root string, elements []string) string {
	cycle := pt.options.StringArray(Cycle, []string{})
	skipColorize := len(cycle) == 0
	folderSeparator := pt.getFolderSeparator()
	colorSeparator := pt.options.Bool(CycleFolderSeparator, false)
	folderFormat := pt.options.String(FolderFormat, "%s")

	edgeFormat := pt.options.String(EdgeFormat, folderFormat)
	leftFormat := pt.options.String(LeftFormat, edgeFormat)
	rightFormat := pt.options.String(RightFormat, edgeFormat)

	colorizeElement := func(element string) string {
		if skipColorize || element == "" {
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

	// Pre-calculate total capacity needed
	totalLen := len(root)
	for _, el := range elements {
		totalLen += len(el) + 20 // estimate for color codes
	}

	sb := text.NewBuilder()

	sb.Grow(totalLen)

	formattedRoot := fmt.Sprintf(leftFormat, root)
	sb.WriteString(colorizeElement(formattedRoot))

	if !pt.endWithSeparator(root) {
		sb.WriteString(colorizeSeparator())
	}

	for i, element := range elements {
		if element == "" {
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

	if pt.relative == "" {
		return folders
	}

	elements := strings.SplitSeq(pt.relative, pt.pathSeparator)
	folderFormatMap := pt.makeFolderFormatMap()
	currentPath := pt.root

	if !pt.endWithSeparator(pt.root) {
		currentPath += pt.pathSeparator
	}

	var display bool

	for element := range elements {
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

	if gitDirFormat := pt.options.String(GitDirFormat, ""); len(gitDirFormat) != 0 {
		dir, err := pt.env.HasParentFilePath(".git", false)
		if err == nil && dir.IsDir {
			// Make it consistent with the modified parent.
			parent := pt.join(pt.replaceMappedLocations(dir.ParentFolder))
			folderFormatMap[parent] = gitDirFormat
		}
	}

	return folderFormatMap
}
