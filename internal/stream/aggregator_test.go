package stream

import (
	"reflect"
	"testing"
)

func TestAggregator(t *testing.T) {
	events := []Event{
		{ServerName: "commons.wikimedia.org", Bot: true, ByteDelta: 100, Timestamp: 100},
		{ServerName: "en.wikipedia.org", Bot: false, ByteDelta: -20, Timestamp: 100},
		{ServerName: "commons.wikimedia.org", Bot: true, ByteDelta: 5, Timestamp: 101},
	}

	agg := NewAggregator()
	for _, ev := range events {
		agg.Update(ev)
	}

	wantSpark := make([]int, 60)
	wantSpark[58] = 2 // second 100 → two events
	wantSpark[59] = 1 // second 101 → one event

	got := agg.Snapshot()
	want := Stats{
		TotalEdits: 3,
		BotEdits:   2,
		NetBytes:   85,
		TopWikis: []WikiCount{
			{"commons.wikimedia.org", 2},
			{"en.wikipedia.org", 1},
		},
		Sparkline: wantSpark,
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("Snapshot() = %+v, want %+v", got, want)
	}
}
