package content

import (
	_ "embed"
	"sort"

	"bytes"

	"github.com/yuin/goldmark"
	"gopkg.in/yaml.v3"
)

//go:embed content.yaml
var siteData []byte

//go:embed CONTEXT.md
var contextData []byte

func LoadContext() ([]byte, error) {
	var buf bytes.Buffer
	if err := goldmark.Convert(contextData, &buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

type PageMeta struct {
	Title       string `yaml:"title"`
	Description string `yaml:"description"`
}

type Section struct {
	ID      string `yaml:"id"`
	Title   string `yaml:"title"`
	Tagline string `yaml:"tagline"`
	Cols    int    `yaml:"cols"`
	Cards   []Card `yaml:"cards"`
}

type HomePage struct {
	Meta     PageMeta  `yaml:"meta"`
	Title    string    `yaml:"title"`
	Tagline  string    `yaml:"tagline"`
	Cards    []Card    `yaml:"cards"`
	Sections []Section `yaml:"sections"`
}

type Button struct {
	Text   string `yaml:"text"`
	Href   string `yaml:"href"`
	Target string `yaml:"target"`
}

type Card struct {
	Title       string `yaml:"title"`
	Description string `yaml:"description"`
	Href        string `yaml:"href"`
	Order       int    `yaml:"order"`
	Button      Button `yaml:"button"`
	Icon        string `yaml:"icon"`
}

type AboutPage struct {
	Meta  PageMeta `yaml:"meta"`
	Title string   `yaml:"title"`
	Body  string   `yaml:"body"`
}

type SiteContent struct {
	Home  HomePage  `yaml:"home"`
	About AboutPage `yaml:"about"`
}

func Load() (SiteContent, error) {
	var s SiteContent
	err := yaml.Unmarshal(siteData, &s)
	if err != nil {
		return s, err
	}
	sort.Slice(s.Home.Cards, func(i, j int) bool {
		return s.Home.Cards[i].Order < s.Home.Cards[j].Order
	})
	return s, err
}
