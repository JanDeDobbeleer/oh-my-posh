package gitstatus

import (
	"container/heap"

	"github.com/go-git/go-git/v5/plumbing"
)

// Reachability flags used by aheadBehind's paint_down_to_common walk.
const (
	flagOurs   = 1
	flagTheirs = 2
	flagStale  = 4
)

type queuedCommit struct {
	info *commitInfo
	hash plumbing.Hash
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

func hasNonStale(queue commitPQ, flags map[plumbing.Hash]int) bool {
	for _, qc := range queue {
		if flags[qc.hash]&flagStale == 0 {
			return true
		}
	}

	return false
}

// aheadBehind implements git's paint_down_to_common: walk both tips by
// committer-date priority, flag reachability from each side, and count the
// commits reachable from only one side.
func aheadBehind(store *objectStore, ours, theirs plumbing.Hash) (int, int, error) {
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

	for hasNonStale(queue, flags) {
		commit := heap.Pop(&queue).(queuedCommit)
		flag := flags[commit.hash]

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
				// Re-queue so the new flags propagate further; duplicate
				// entries are harmless because flags are read at pop time.
				heap.Push(&queue, queuedCommit{hash: parent, info: infos[parent]})
				continue
			}

			if err := push(parent, propagate); err != nil {
				return 0, 0, err
			}
		}
	}

	// Flags only grow, and committer dates can tie or run backwards, so a
	// commit may gain its second flag after it was popped. Counting while
	// walking would keep the early, one-sided verdict.
	ahead, behind := 0, 0
	for _, flag := range flags {
		if flag&flagStale != 0 {
			continue
		}

		switch flag {
		case flagOurs:
			ahead++
		case flagTheirs:
			behind++
		}
	}

	return ahead, behind, nil
}
