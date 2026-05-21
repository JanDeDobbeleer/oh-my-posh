package segments

import "strings"

// DvcStatus represents part of the status of a DVC repository
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
	return " \ue8d1 {{.Status.String}} "
}

func (d *Dvc) Enabled() bool {
	if !d.hasCommand(DVCCOMMAND) {
		return false
	}

	// Check if we're in a DVC repository
	_, err := d.env.HasParentFilePath(".dvc", false)
	if err != nil {
		return false
	}

	// run dvc status command
	output, err := d.env.RunCommand(d.command, "status", "-q")
	if err != nil {
		return false
	}

	statusFormats := d.options.KeyValueMap(StatusFormats, map[string]string{})
	d.Status = &DvcStatus{ScmStatus: ScmStatus{Formats: statusFormats}}

	if output == "" {
		return true
	}

	lines := strings.SplitSeq(output, "\n")

	for line := range lines {
		if line == "" {
			continue
		}

		// DVC status output format:
		// data.xml: modified
		// or
		// data/: not in cache
		parts := strings.SplitN(line, ":", 2)
		if len(parts) < 2 {
			continue
		}

		status := strings.TrimSpace(parts[1])
		d.Status.add(status)
	}

	return true
}
