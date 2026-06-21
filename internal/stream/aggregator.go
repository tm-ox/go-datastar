package stream

import "sync"

type Stats struct {
	TotalEdits int
	BotEdits   int
	NetBytes   int
}

type Aggregator struct {
	mu    sync.Mutex
	stats Stats
}

func NewAggregator() *Aggregator {
	return &Aggregator{}
}

func (a *Aggregator) Update(ev Event) {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.stats.TotalEdits++
	if ev.Bot {
		a.stats.BotEdits++
	}
	a.stats.NetBytes += ev.ByteDelta
}

func (a *Aggregator) Snapshot() Stats {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.stats
}
