package stream

import (
	"bufio"
	"context"
	"log"
	"net/http"
	"strings"
	"time"
)

type Source struct {
	hub       *Hub
	url       string
	userAgent string
}

func NewSource(hub *Hub, userAgent string) *Source {
	return &Source{
		hub:       hub,
		url:       "https://stream.wikimedia.org/v2/stream/recentchange",
		userAgent: userAgent,
	}
}

func (s *Source) stream(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, s.url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", s.userAgent)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		data, ok := strings.CutPrefix(scanner.Text(), "data: ")
		if !ok {
			continue // skip event:/id:/blank/comment lines
		}
		ev, err := ParseEvent([]byte(data))
		if err != nil {
			continue // one bad line shouldn't kill the stream
		}
		s.hub.Broadcast(ev)
	}
	return scanner.Err()
}

func (s *Source) Run(ctx context.Context) {
	for {
		if err := s.stream(ctx); err != nil {
			log.Printf("stream: %v", err)
		}
		if ctx.Err() != nil {
			return // shutdown — don't reconnect
		}
		select {
		case <-ctx.Done():
			return
		case <-time.After(time.Second):
		}
	}
}
