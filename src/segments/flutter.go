package segments
import "errors"

type Flutter struct {
	Language
}

func (f *Flutter) Template() string {
	return languageTemplate
}

func (f *Flutter) Enabled() bool {
	f.extensions = dartExtensions
	f.folders = dartFolders
	f.tooling = map[string]*cmd{
		"fvm": {
			regex:      `(?P<version>((?P<major>[0-9]+).(?P<minor>[0-9]+).(?P<patch>[0-9]+)))`,
			getVersion: f.fvmVersion,
		},
		"flutter": {
			executable: "flutter",
			args:       []string{"--version"},
			regex:      `Flutter (?P<version>((?P<major>[0-9]+).(?P<minor>[0-9]+).(?P<patch>[0-9]+)))`,
		},
	}
	f.defaultTooling = []string{"fvm", "flutter"}
	f.versionURLTemplate = "https://github.com/flutter/flutter/releases/tag/{{ .Major }}.{{ .Minor }}.{{ .Patch }}"

	return f.Language.Enabled()
}

func (f *Flutter) fvmVersion() (string, error) {
	if !f.env.HasCommand("fvm") {
		return "", errors.New(noVersion)
	}

	if !f.env.HasFolder(".fvm") {
		return "", errors.New(noVersion)
	}

	if version, _ := f.parseFvmrc(); version != "" {
		return version, nil
	}

	if version, _ := f.parseDartToolVersion(); version != "" {
		return version, nil
	}

	return "", errors.New(noVersion)
}

func (f *Flutter) parseFvmrc() (string, error) {
	fvmrcPath, err := f.env.HasParentFilePath(".fvmrc", false)
	if err != nil {
		return "", err
	}

	content := f.env.FileContent(fvmrcPath)
	if content == "" {
		return "", nil
	}

	return content, nil
}

func (f *Flutter) parseDartToolVersion() (string, error) {
	dartToolVersionPath, err := f.env.HasParentFilePath(".dart_tool/version", false)
	if err != nil {
		return "", err
	}

	content := f.env.FileContent(dartToolVersionPath)
	if content == "" {
		return "", nil
	}

	return content, nil
}
