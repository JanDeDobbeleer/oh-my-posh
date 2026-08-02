package gitstatus

import (
	"container/heap"

	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/storer"
)

// Reachability flags used by aheadBehind's paint_down_to_common walk.
const (
	flagOurs   = 1
	flagTheirs = 2
	flagStale  = 4
)

type queuedCommit struct {
	hash plumbing.Hash
	info *commitInfo
}

type commitPQ []queuedCommit

func (q commitPQ) Len() int      { return len(q) }
func (q commitPQ) Swap(i, j int) { q[i], q[j] = q[j], q[i] }
func (q commitPQ) Less(i, j int) bool {
	return q[i].info.CommitterWhen > q[j].info.CommitterWhen
}

func (q *commitPQ) Push(x any) { *q = append(*q, x.(queuedCommit)) }

func (q *commitPQ) Pop() any {
	old := *q
	n := len(old)
	item := old[n-1]
	*q = old[:n-1]
	return item
}

// aheadBehind implements git's paint_down_to_common: walk both tips by
// committer-date priority, flag reachability from each side, and count the
// commits reachable from only one side.
func aheadBehind(store storer.EncodedObjectStorer, ours, theirs plumbing.Hash) (int, int, error) {
	if ours == theirs {
		return 0, 0, nil
	}

	flags := map[plumbing.Hash]int{}
	infos := map[plumbing.Hash]*commitInfo{}
	var queue commitPQ

	push := func(h plumbing.Hash, flag int) error {
		if old, ok := flags[h]; ok {
			flags[h] = old | flag
			return nil
		}

		info, err := readCommit(store, h)
		if err != nil {
			return err
		}

		flags[h] = flag
		infos[h] = info
		heap.Push(&queue, queuedCommit{hash: h, info: info})
		return nil
	}

	if err := push(ours, flagOurs); err != nil {
		return 0, 0, err
	}
	if err := push(theirs, flagTheirs); err != nil {
		return 0, 0, err
	}

	ahead, behind := 0, 0
	interesting := 2

	for queue.Len() > 0 && interesting > 0 {
		commit := heap.Pop(&queue).(queuedCommit)
		flag := flags[commit.hash]

		if flag&flagStale == 0 {
			switch flag & (flagOurs | flagTheirs) {
			case flagOurs:
				ahead++
			case flagTheirs:
				behind++
			}
		}

		propagate := flag
		if flag&(flagOurs|flagTheirs) == flagOurs|flagTheirs {
			propagate |= flagStale
		}

		for _, parent := range commit.info.Parents {
			old, seen := flags[parent]
			if seen {
				merged := old | propagate
				if merged == old {
					continue
				}

				flags[parent] = merged
				// Re-queue so the stale flag propagates further; duplicate
				// entries are harmless because counting checks flags at pop
				// time, not at push time.
				heap.Push(&queue, queuedCommit{hash: parent, info: infos[parent]})
				continue
			}

			if err := push(parent, propagate); err != nil {
				return 0, 0, err
			}
		}

		// Termination heuristic: once every queued commit is flagged on
		// both sides, nothing one-sided remains to discover.
		allShared := true
		for _, qc := range queue {
			f := flags[qc.hash]
			if f&(flagOurs|flagTheirs) != flagOurs|flagTheirs {
				allShared = false
				break
			}
		}
		if allShared {
			interesting = 0
		}
	}

	return ahead, behind, nil
}
