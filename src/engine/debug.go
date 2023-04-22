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
	e.write(log.Text("\nSegments:\n\n").Green().Bold().Plain().String())
	// console title timing
	titleStartTime := time.Now()
	e.Env.Debug("Segment: Title")
	title := e.getTitleTemplateText()
	consoleTitleTiming := &SegmentTiming{
		name:       "ConsoleTitle",
		nameLength: 12,
		active:     len(e.Config.ConsoleTitleTemplate) > 0,
		text:       title,
		duration:   time.Since(titleStartTime),
	}
	largestSegmentNameLength := consoleTitleTiming.nameLength
	var segmentTimings []*SegmentTiming
	segmentTimings = append(segmentTimings, consoleTitleTiming)
	// cache a pointer to the color cycle
	cycle = &e.Config.Cycle
	// loop each segments of each blocks
	for _, block := range e.Config.Blocks {
		block.Init(e.Env, e.Writer)
		longestSegmentName, timings := block.Debug()
		segmentTimings = append(segmentTimings, timings...)
		if longestSegmentName > largestSegmentNameLength {
			largestSegmentNameLength = longestSegmentName
		}
	}

	// 22 is the color for false/true and 7 is the reset color
	largestSegmentNameLength += 22 + 7
	for _, segment := range segmentTimings {
		duration := segment.duration.Milliseconds()
		var active log.Text
		if segment.active {
			active = log.Text("true").Yellow()
		} else {
			active = log.Text("false").Purple()
		}
		segmentName := fmt.Sprintf("%s(%s)", segment.name, active.Plain())
		e.write(fmt.Sprintf("%-*s - %3d ms - %s\n", largestSegmentNameLength, segmentName, duration, segment.text))
	}
	e.write(fmt.Sprintf("\n%s %s\n", log.Text("Run duration:").Green().Bold().Plain(), time.Since(startTime)))
	e.write(fmt.Sprintf("\n%s %s\n", log.Text("Cache path:").Green().Bold().Plain(), e.Env.CachePath()))
	e.write(fmt.Sprintf("\n%s %s\n", log.Text("Config path:").Green().Bold().Plain(), e.Env.Flags().Config))
	e.write(log.Text("\nLogs:\n\n").Green().Bold().Plain().String())
	e.write(e.Env.Logs())
	return e.string()
}
