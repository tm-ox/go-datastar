package content

import (
	"embed"
	"sort"
	"strconv"
	"strings"

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

func UniqueTypes(entries []WorkEntry) []string {
	seen := map[string]bool{}
	var out []string
	for _, e := range entries {
		if e.Type != "" && !seen[e.Type] {
			seen[e.Type] = true
			out = append(out, e.Type)
		}
	}
	sort.Strings(out)
	return out
}

func UniqueClients(entries []WorkEntry) []string {
	seen := map[string]bool{}
	var out []string
	for _, e := range entries {
		if e.Client != "" && !seen[e.Client] {
			seen[e.Client] = true
			out = append(out, e.Client)
		}
	}
	sort.Strings(out)
	return out
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

func FilterWork(entries []WorkEntry, typeFilter, clientFilter, yearFilter, toolFilter string) []WorkEntry {
	var out []WorkEntry
	for _, e := range entries {
		if typeFilter != "" && e.Type != typeFilter {
			continue
		}
		if clientFilter != "" && e.Client != clientFilter {
			continue
		}
		if yearFilter != "" && strconv.Itoa(e.Year) != yearFilter {
			continue
		}
		if toolFilter != "" {
			match := false
			for _, t := range e.Tools {
				if t == toolFilter {
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

func SortWork(entries []WorkEntry, by string) []WorkEntry {
	out := make([]WorkEntry, len(entries))
	copy(out, entries)
	parts := strings.SplitN(by, "-", 2)
	if len(parts) != 2 {
		return out
	}
	col, dir := parts[0], parts[1]
	sort.Slice(out, func(i, j int) bool {
		switch col {
		case "title":
			if dir == "asc" {
				return out[i].Title < out[j].Title
			}
			return out[i].Title > out[j].Title
		case "type":
			if dir == "asc" {
				return out[i].Type < out[j].Type
			}
			return out[i].Type > out[j].Type
		case "client":
			if dir == "asc" {
				return out[i].Client < out[j].Client
			}
			return out[i].Client > out[j].Client
		case "year":
			if dir == "asc" {
				return out[i].Year < out[j].Year
			}
			return out[i].Year > out[j].Year
		case "tools":
			if dir == "asc" {
				return strings.Join(out[i].Tools, ",") < strings.Join(out[j].Tools, ",")
			}
			return strings.Join(out[i].Tools, ",") > strings.Join(out[j].Tools, ",")
		}
		return false
	})
	return out
}
