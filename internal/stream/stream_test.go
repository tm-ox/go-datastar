package stream

import (
	_ "embed"
	"testing"
)

//go:embed testdata/edit.json
var testEvent []byte

//go:embed testdata/categorize.json
var testCategorizeEvent []byte

func TestParseEvent(t *testing.T) {
	tests := []struct {
		name       string
		raw        []byte
		wantType   string
		wantDelta  int
		wantServer string
	}{
		{"edit has byte delta", testEvent, "edit", 3808, "commons.wikimedia.org"},
		{"categorize has no length", testCategorizeEvent, "categorize", 0, "commons.wikimedia.org"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ev, err := ParseEvent(tt.raw)
			if err != nil {
				t.Fatalf("ParseEvent returned error: %v", err)
			}
			if ev.Type != tt.wantType {
				t.Errorf("Type = %q, want %q", ev.Type, tt.wantType)
			}
			if ev.ByteDelta != tt.wantDelta {
				t.Errorf("ByteDelta = %d, want %d", ev.ByteDelta, tt.wantDelta)
			}
			if ev.ServerName != tt.wantServer {
				t.Errorf("ServerName = %q, want %q", ev.ServerName, tt.wantServer)
			}
		})
	}
}
