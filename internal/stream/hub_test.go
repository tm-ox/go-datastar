package stream

import "testing"

func TestHubBroadcastReachesAllSubscribers(t *testing.T) {
	h := NewHub()
	a, _, cancelA := h.Subscribe()
	defer cancelA()
	b, _, cancelB := h.Subscribe()
	defer cancelB()

	h.Broadcast(Event{Title: "Hello"})

	if got := <-a; got.Title != "Hello" {
		t.Errorf("subscriber A got %q, want %q", got.Title, "Hello")
	}
	if got := <-b; got.Title != "Hello" {
		t.Errorf("subscriber B got %q, want %q", got.Title, "Hello")
	}
}

func TestHubDropsWhenSubscriberFull(t *testing.T) {
	h := NewHub()
	ch, _, cancel := h.Subscribe() // buffer is 32
	defer cancel()

	// Never drain ch; send far more than it can hold.
	for i := 0; i < 50; i++ {
		h.Broadcast(Event{Type: "edit"})
	}

	// Reaching this line at all proves Broadcast never blocked.
	if got := len(ch); got != 32 {
		t.Errorf("buffered events = %d, want 32 (extras dropped)", got)
	}
}

func TestHubRecentSeedsNewSubscriber(t *testing.T) {
	h := NewHub()
	for i := 0; i < 25; i++ {
		h.Broadcast(Event{Title: "e"})
	}

	_, recent, cancel := h.Subscribe()
	defer cancel()

	if len(recent) != recentSize {
		t.Errorf("recent = %d events, want %d", len(recent), recentSize)
	}
}
