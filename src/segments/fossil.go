package segments

import "strings"

// FossilStatus represents part of the status of a Svn repository
type FossilStatus struct {
	ScmStatus
}

func (s *FossilStatus) add(code string) {
	switch code {
	case "CONFLICT":
		s.Conflicted++
	case "DELETED":
		s.Deleted++
	case "ADDED":
		s.Added++
	case "EDITED", "UPDATED", "CHANGED":
		s.Modified++
	case "RENAMED":
		s.Moved++
	}
}

const (
	FOSSILCOMMAND = "fossil"
)

type Fossil struct {
	scm

	Status *FossilStatus
	Branch string
}

func (f *Fossil) Template() string {
	return " \ue725 {{.Branch}} {{.Status.String}} "
}

func (f *Fossil) Enabled() bool {
	if !f.hasCommand(FOSSILCOMMAND) {
		return false
	}
	// run fossil command
	output, err := f.env.RunCommand(f.command, "status")
	if err != nil {
		return false
	}
	f.Status = &FossilStatus{}
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if len(line) == 0 {
			continue
		}
		context := strings.SplitN(line, " ", 2)
		if len(context) < 2 {
			continue
		}
		switch context[0] {
		case "tags:":
			f.Branch = strings.TrimSpace(context[1])
		default:
			f.Status.add(context[0])
		}
	}
	return true
}
