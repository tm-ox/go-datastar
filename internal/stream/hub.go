package stream

import "sync"

type Hub struct {
	mu          sync.Mutex
	subscribers map[chan Event]struct{}
	recent      []Event
}

const recentSize = 20

func NewHub() *Hub {
	return &Hub{
		subscribers: make(map[chan Event]struct{}),
	}
}

func (h *Hub) Subscribe() (<-chan Event, []Event, func()) {
	ch := make(chan Event, 32)

	h.mu.Lock()
	h.subscribers[ch] = struct{}{}
	recent := make([]Event, len(h.recent))
	copy(recent, h.recent)
	h.mu.Unlock()

	cancel := func() {
		h.mu.Lock()
		delete(h.subscribers, ch)
		close(ch)
		h.mu.Unlock()
	}

	return ch, recent, cancel
}

func (h *Hub) Broadcast(ev Event) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.recent = append(h.recent, ev)
	if len(h.recent) > recentSize {
		h.recent = h.recent[1:]
	}

	for ch := range h.subscribers {
		select {
		case ch <- ev:
		default:
		}
	}
}
