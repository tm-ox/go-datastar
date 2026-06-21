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
}

type Aggregator struct {
	mu    sync.Mutex
	stats Stats
	wikis map[string]int
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

	out := a.stats
	out.TopWikis = top
	return out
}
