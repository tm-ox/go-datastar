package stream

import "testing"

func TestAggregator(t *testing.T) {
	events := []Event{
		{Bot: true, ByteDelta: 100},
		{Bot: false, ByteDelta: -20},
		{Bot: true, ByteDelta: 5},
	}

	agg := NewAggregator()
	for _, ev := range events {
		agg.Update(ev)
	}

	got := agg.Snapshot()
	want := Stats{TotalEdits: 3, BotEdits: 2, NetBytes: 85}

	if got != want {
		t.Errorf("Snapshot() = %+v, want %+v", got, want)
	}
}
