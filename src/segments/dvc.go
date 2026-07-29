package segments

import "encoding/json"

type DvcStatus struct {
	ScmStatus
}

func (s *DvcStatus) add(code string) {
	switch code {
	case "not in cache":
		s.Missing++
	case "deleted":
		s.Deleted++
	case "new":
		s.Added++
	case "modified":
		s.Modified++
	}
}

const (
	DVCCOMMAND = "dvc"
)

type Dvc struct {
	Status *DvcStatus
	Scm
}

func (d *Dvc) Template() string {
	return "  {{ .Status.String }} "
}

func (d *Dvc) Enabled() bool {
	if !d.hasCommand(DVCCOMMAND) {
		return false
	}

	// only display when we're actually inside a DVC repository
	if _, err := d.env.HasParentFilePath(".dvc", false); err != nil {
		return false
	}

	statusFormats := d.options.KeyValueMap(StatusFormats, map[string]string{})
	d.Status = &DvcStatus{ScmStatus: ScmStatus{Formats: statusFormats}}

	output, err := d.env.RunCommand(d.command, "status", "--json")
	if err != nil {
		return false
	}

	d.setStatus(output)

	return true
}

func (d *Dvc) CacheKey() (string, bool) {
	dir, err := d.env.HasParentFilePath(".dvc", true)
	if err != nil {
		return "", false
	}

	return dir.Path, true
}

// `dvc status --json` has the shape:
//
//	{"<stage>": [{"changed outs": {"<file>": "<state>"}}, {"changed deps": {"<file>": "<state>"}}], ...}
//
// a clean workspace returns an empty object ({}). Unknown states are ignored.
func (d *Dvc) setStatus(output string) {
	if output == "" {
		return
	}

	var status map[string][]map[string]map[string]string
	if err := json.Unmarshal([]byte(output), &status); err != nil {
		return
	}

	for _, sections := range status {
		for _, section := range sections {
			for _, files := range section {
				for _, state := range files {
					d.Status.add(state)
				}
			}
		}
	}
}
