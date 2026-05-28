package content

import (
	"embed"
	"sort"
	"strconv"

	"gopkg.in/yaml.v3"
)

//go:embed content.yaml
var siteData []byte

type Card struct {
	Title       string `yaml:"title"`
	Description string `yaml:"description"`
	Href        string `yaml:"href"`
	Order       int    `yaml:"order"`
}

type HomePage struct {
	Title   string `yaml:"title"`
	Tagline string `yaml:"tagline"`
	Cards   []Card `yaml:"cards"`
}

type AboutPage struct {
	Title string `yaml:"title"`
	Body  string `yaml:"body"`
}

type WorkEntry struct {
	Slug        string   `yaml:"-"`
	Title       string   `yaml:"title"`
	Type        string   `yaml:"type"`
	Client      string   `yaml:"client"`
	Year        int      `yaml:"year"`
	Tools       []string `yaml:"tools"`
	Description string   `yaml:"description"`
}

type SiteContent struct {
	Home  HomePage  `yaml:"home"`
	About AboutPage `yaml:"about"`
}

func Load() (SiteContent, error) {
	var s SiteContent
	err := yaml.Unmarshal(siteData, &s)
	return s, err
}

//go:embed work/*.yaml
var workFS embed.FS

func LoadWork() ([]WorkEntry, error) {
	files, err := workFS.ReadDir("work")
	if err != nil {
		return nil, err
	}
	var entries []WorkEntry
	for _, f := range files {
		name := f.Name()
		src, err := workFS.ReadFile("work/" + name)
		if err != nil {
			return nil, err
		}
		var e WorkEntry
		if err := yaml.Unmarshal(src, &e); err != nil {
			return nil, err
		}
		e.Slug = name[:len(name)-5]
		entries = append(entries, e)
	}
	return entries, nil
}

func UniqueYears(entries []WorkEntry) []string {
	seen := map[int]bool{}
	var out []string
	for _, e := range entries {
		if !seen[e.Year] {
			seen[e.Year] = true
			out = append(out, strconv.Itoa(e.Year))
		}
	}
	sort.Sort(sort.Reverse(sort.StringSlice(out)))
	return out
}

func UniqueTools(entries []WorkEntry) []string {
	seen := map[string]bool{}
	var out []string
	for _, e := range entries {
		for _, t := range e.Tools {
			if !seen[t] {
				seen[t] = true
				out = append(out, t)
			}
		}
	}
	sort.Strings(out)
	return out
}

func FilterWork(entries []WorkEntry, year, tool string) []WorkEntry {
	var out []WorkEntry
	for _, e := range entries {
		if year != "" && strconv.Itoa(e.Year) != year {
			continue
		}
		if tool != "" {
			match := false
			for _, t := range e.Tools {
				if t == tool {
					match = true
					break
				}
			}
			if !match {
				continue
			}
		}
		out = append(out, e)
	}
	return out
}
