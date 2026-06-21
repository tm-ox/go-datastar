package stream

import "encoding/json"

type Event struct {
	Type       string
	Title      string
	User       string
	Bot        bool
	ServerName string
	Timestamp  int64
	ByteDelta  int
}

type wireEvent struct {
	Type       string `json:"type"`
	Title      string `json:"title"`
	User       string `json:"user"`
	Bot        bool   `json:"bot"`
	ServerName string `json:"server_name"`
	Timestamp  int64  `json:"timestamp"`
	Length     *struct {
		Old int `json:"old"`
		New int `json:"new"`
	} `json:"length"`
}

func ParseEvent(raw []byte) (Event, error) {
	var wire wireEvent
	if err := json.Unmarshal(raw, &wire); err != nil {
		return Event{}, err
	}
	ev := Event{
		Type:       wire.Type,
		Title:      wire.Title,
		User:       wire.User,
		Bot:        wire.Bot,
		ServerName: wire.ServerName,
		Timestamp:  wire.Timestamp,
	}
	if wire.Length != nil {
		ev.ByteDelta = wire.Length.New - wire.Length.Old
	}
	return ev, nil
}
