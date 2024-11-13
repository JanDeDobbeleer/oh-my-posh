package prompt

import (
	"fmt"
	"time"

	"github.com/jandedobbeleer/oh-my-posh/src/cache"
	"github.com/jandedobbeleer/oh-my-posh/src/config"
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
	log.Debug("segment: Title")
	consoleTitle := &config.Segment{
		Alias:      "ConsoleTitle",
		NameLength: 12,
		Enabled:    len(e.Config.ConsoleTitleTemplate) > 0,
		Duration:   time.Since(titleStartTime),
		Type:       config.TEXT,
	}
	_ = consoleTitle.MapSegmentWithWriter(e.Env)
	consoleTitle.SetText(e.getTitleTemplateText())

	largestSegmentNameLength := consoleTitle.NameLength

	// render prompt
	e.write(log.Text("\nPrompt:\n\n").Green().Bold().Plain().String())
	e.write(e.Primary())

	e.write(log.Text("\n\nSegments:\n\n").Green().Bold().Plain().String())

	var segments []*config.Segment
	segments = append(segments, consoleTitle)

	for _, block := range e.Config.Blocks {
		for _, segment := range block.Segments {
			segments = append(segments, segment)
			if segment.NameLength > largestSegmentNameLength {
				largestSegmentNameLength = segment.NameLength
			}
		}
	}

	// 22 is the color for false/true and 7 is the reset color
	largestSegmentNameLength += 22 + 7
	for _, segment := range segments {
		duration := segment.Duration.Milliseconds()
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
	e.write(fmt.Sprintf("\n%s %s\n", log.Text("Cache path:").Green().Bold().Plain(), cache.Path()))

	cfg := e.Env.Flags().Config
	if len(cfg) == 0 {
		cfg = "no --config set, using default built-in configuration"
	}

	e.write(fmt.Sprintf("\n%s %s\n", log.Text("Config path:").Green().Bold().Plain(), cfg))

	e.write(log.Text("\nLogs:\n\n").Green().Bold().Plain().String())
	e.write(e.Env.Logs())
	return e.string()
}
