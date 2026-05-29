package content

import (
	_ "embed"
	"sort"

	"gopkg.in/yaml.v3"
)

//go:embed content.yaml
var siteData []byte

type HomePage struct {
	Title   string `yaml:"title"`
	Tagline string `yaml:"tagline"`
	Cards   []Card `yaml:"cards"`
}

type Card struct {
	Title       string `yaml:"title"`
	Description string `yaml:"description"`
	Href        string `yaml:"href"`
	Order       int    `yaml:"order"`
}

type AboutPage struct {
	Title string `yaml:"title"`
	Body  string `yaml:"body"`
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
