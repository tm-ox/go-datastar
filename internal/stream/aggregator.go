package stream

import (
	"sort"
	"sync"
)

type WikiCount struct {
	Name  string
	Count int
}

type Stats struct {
	TotalEdits int
	BotEdits   int
	NetBytes   int
	TopWikis   []WikiCount
	Sparkline  []int
}

type Aggregator struct {
	mu      sync.Mutex
	stats   Stats
	wikis   map[string]int
	buckets [60]int
	lastSec int64
}

func NewAggregator() *Aggregator {
	return &Aggregator{
		wikis: make(map[string]int),
	}
}

func (a *Aggregator) Update(ev Event) {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.stats.TotalEdits++
	if ev.Bot {
		a.stats.BotEdits++
	}
	a.stats.NetBytes += ev.ByteDelta
	a.wikis[ev.ServerName]++

	sec := ev.Timestamp
	if sec > a.lastSec {
		gap := sec - a.lastSec
		if gap >= 60 {
			a.buckets = [60]int{} // jumped past the whole window — wipe
		} else {
			for s := a.lastSec + 1; s <= sec; s++ {
				a.buckets[s%60] = 0 // clear each second we move into
			}
		}
		a.lastSec = sec
	}
	a.buckets[sec%60]++
}

func (a *Aggregator) Snapshot() Stats {
	a.mu.Lock()
	defer a.mu.Unlock()

	top := make([]WikiCount, 0, len(a.wikis))
	for name, count := range a.wikis {
		top = append(top, WikiCount{name, count})
	}
	sort.Slice(top, func(i, j int) bool {
		return top[i].Count > top[j].Count
	})
	if len(top) > 5 {
		top = top[:5]
	}

	spark := make([]int, 60)
	for j := 0; j < 60; j++ {
		sec := a.lastSec - 59 + int64(j)
		if sec < 0 {
			continue // before any data — leave 0
		}
		spark[j] = a.buckets[sec%60]
	}

	out := a.stats
	out.TopWikis = top
	out.Sparkline = spark
	return out
}
