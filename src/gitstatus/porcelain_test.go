package gitstatus

import (
	"strconv"
	"strings"
)

// parsePorcelainV2 replicates segments.Git.setStatus's parsing of
// `git status --porcelain=2 --branch`, without the merge-in-progress AA
// special case: that branch is unreachable at setStatus/setStatusNative time
// (Merge/Rebase are only known after setHEADStatus runs), so parity tests
// must not special-case it either.
func parsePorcelainV2(output string) *Result {
	result := &Result{UpstreamGone: true}

	const (
		hashPrefix     = "# branch.oid "
		refPrefix      = "# branch.head "
		upstreamPrefix = "# branch.upstream "
		abPrefix       = "# branch.ab "
	)

	for line := range strings.SplitSeq(output, "\n") {
		switch {
		case strings.HasPrefix(line, hashPrefix) && len(line) >= len(hashPrefix)+7:
			result.Hash = line[len(hashPrefix):]
		case strings.HasPrefix(line, refPrefix) && len(line) > len(refPrefix):
			result.Ref = line[len(refPrefix):]
		case strings.HasPrefix(line, upstreamPrefix) && len(line) > len(upstreamPrefix):
			result.Upstream = line[len(upstreamPrefix):]
			result.UpstreamGone = true
		case strings.HasPrefix(line, abPrefix) && len(line) > len(abPrefix):
			status := line[len(abPrefix):]
			splitted := strings.SplitN(status, " ", 3)
			if len(splitted) >= 2 {
				result.Ahead, _ = strconv.Atoi(splitted[0])
				behind, _ := strconv.Atoi(splitted[1])
				result.Behind = -behind
			}
			result.UpstreamGone = false
		default:
			addPorcelainLine(result, line)
		}
	}

	return result
}

func addPorcelainLine(result *Result, line string) {
	const untracked = "?"

	if strings.HasPrefix(line, untracked) {
		result.Working.Untracked++
		return
	}

	if len(line) <= 4 {
		return
	}

	// Mirrors segments.GitStatus.add's full switch (D/A/U/M plus the
	// R/C/m aliases for Modified and the "." no-op), independently of the
	// production addCode helper: this is what proves the native engine's
	// own rename-pairing/conflict-mapping output lines up with what git.go
	// would compute from real porcelain text.
	addPorcelainCode(&result.Staging, line[2])
	addPorcelainCode(&result.Working, line[3])
}

func addPorcelainCode(counts *Counts, code byte) {
	switch code {
	case '.':
		return
	case 'D':
		counts.Deleted++
	case 'A':
		counts.Added++
	case 'U':
		counts.Unmerged++
	case 'M', 'R', 'C', 'm':
		counts.Modified++
	}
}
