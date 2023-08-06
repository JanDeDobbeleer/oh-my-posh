package engine

import (
	"fmt"
	"time"

	"github.com/jandedobbeleer/oh-my-posh/src/log"
)

// debug will loop through your config file and output the timings for each segments
func (e *Engine) PrintDebug(startTime time.Time, version string) string {
	e.write(fmt.Sprintf("\n%s %s\n", log.Text("Version:").Green().Bold().Plain(), version))
	sh := e.Env.Shell()
	shellVersion := e.Env.Getenv("POSH_SHELL_VERSION")
	if len(shellVersion) != 0 {
		sh += fmt.Sprintf(" (%s)", shellVersion)
	}
	e.write(fmt.Sprintf("\n%s %s\n", log.Text("Shell:").Green().Bold().Plain(), sh))

	// console title timing
	titleStartTime := time.Now()
	e.Env.Debug("Segment: Title")
	title := e.getTitleTemplateText()
	consoleTitle := &Segment{
		name:       "ConsoleTitle",
		nameLength: 12,
		Enabled:    len(e.Config.ConsoleTitleTemplate) > 0,
		text:       title,
		duration:   time.Since(titleStartTime),
	}
	largestSegmentNameLength := consoleTitle.nameLength

	// render prompt
	e.write(log.Text("\nPrompt:\n\n").Green().Bold().Plain().String())
	e.write(e.Primary())

	e.write(log.Text("\n\nSegments:\n\n").Green().Bold().Plain().String())

	var segments []*Segment
	segments = append(segments, consoleTitle)

	for _, block := range e.Config.Blocks {
		for _, segment := range block.Segments {
			segments = append(segments, segment)
			if segment.nameLength > largestSegmentNameLength {
				largestSegmentNameLength = segment.nameLength
			}
		}
	}

	// 22 is the color for false/true and 7 is the reset color
	largestSegmentNameLength += 22 + 7
	for _, segment := range segments {
		duration := segment.duration.Milliseconds()
		var active log.Text
		if segment.Enabled {
			active = log.Text("true").Yellow()
		} else {
			active = log.Text("false").Purple()
		}
		segmentName := fmt.Sprintf("%s(%s)", segment.Name(), active.Plain())
		e.write(fmt.Sprintf("%-*s - %3d ms\n", largestSegmentNameLength, segmentName, duration))
	}

	e.write(fmt.Sprintf("\n%s %s\n", log.Text("Run duration:").Green().Bold().Plain(), time.Since(startTime)))
	e.write(fmt.Sprintf("\n%s %s\n", log.Text("Cache path:").Green().Bold().Plain(), e.Env.CachePath()))

	config := e.Env.Flags().Config
	if len(config) == 0 {
		config = "no --config set, using default built-in configuration"
	}
	e.write(fmt.Sprintf("\n%s %s\n", log.Text("Config path:").Green().Bold().Plain(), config))

	e.write(log.Text("\nLogs:\n\n").Green().Bold().Plain().String())
	e.write(e.Env.Logs())
	return e.string()
}
