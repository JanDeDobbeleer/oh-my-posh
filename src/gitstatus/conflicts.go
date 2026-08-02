package gitstatus

// conflictXY maps the set of index stages present for an unmerged path to
// the porcelain XY conflict code, mirroring git's wt-status.c:
//
//	{1,2,3} -> UU   {2,3} -> AA   {1,2} -> UD
//	{1,3}   -> DU   {1}   -> DD   {2}   -> AU   {3} -> UA
//
// stages is a bitmask of (1<<Stage) for the stages present on that path
// (Stage 1 = base, 2 = ours, 3 = theirs).
func conflictXY(stages int) (staging, working byte) {
	switch stages {
	case 1 << 1:
		return 'D', 'D'
	case 1 << 2:
		return 'A', 'U'
	case 1 << 3:
		return 'U', 'A'
	case 1<<1 | 1<<2:
		return 'U', 'D'
	case 1<<1 | 1<<3:
		return 'D', 'U'
	case 1<<2 | 1<<3:
		return 'A', 'A'
	default:
		return 'U', 'U'
	}
}

// addCode applies a single porcelain status letter to the matching counter,
// mirroring segments.GitStatus.add's D/A/U/M mapping.
func addCode(counts *Counts, code byte) {
	switch code {
	case 'D':
		counts.Deleted++
	case 'A':
		counts.Added++
	case 'U':
		counts.Unmerged++
	case 'M':
		counts.Modified++
	}
}
