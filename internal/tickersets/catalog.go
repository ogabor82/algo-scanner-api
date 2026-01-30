package tickersets

import (
	"errors"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/go-yaml/yaml"
)

type FileCatalog struct {
	Group string              `yaml:"group"`
	Title string              `yaml:"title"`
	Sets  map[string]SetEntry `yaml:"sets"`
}

type SetEntry struct {
	Title   string   `yaml:"title"`
	Tickers []string `yaml:"tickers"`
	Include []string `yaml:"include"`
}

type Catalog struct {
	Groups map[string]FileCatalog
}

func LoadDir(dir string) (*Catalog, error) {
	matches, err := filepath.Glob(filepath.Join(dir, "*.yaml"))
	if err != nil {
		return nil, err
	}
	if len(matches) == 0 {
		return nil, errors.New("no yaml files found in " + dir)
	}

	out := &Catalog{Groups: map[string]FileCatalog{}}

	for _, path := range matches {
		b, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}
		var fc FileCatalog
		if err := yaml.Unmarshal(b, &fc); err != nil {
			return nil, err
		}
		if fc.Group == "" {
			return nil, errors.New("missing 'group' in " + filepath.Base(path))
		}
		if fc.Sets == nil {
			fc.Sets = map[string]SetEntry{}
		}
		if _, exists := out.Groups[fc.Group]; exists {
			return nil, errors.New("duplicate group '" + fc.Group + "' (file: " + filepath.Base(path) + ")")
		}
		out.Groups[fc.Group] = fc
	}

	return out, nil
}

func (c *Catalog) Resolve(id string) (title string, tickers []string, err error) {
	group, set, ok := strings.Cut(id, ".")
	if !ok || group == "" || set == "" {
		return "", nil, errors.New("invalid id, expected 'group.set'")
	}
	g, ok := c.Groups[group]
	if !ok {
		return "", nil, errors.New("unknown group: " + group)
	}
	s, ok := g.Sets[set]
	if !ok {
		return "", nil, errors.New("unknown set: " + id)
	}

	stack := map[string]bool{}
	out, err := expand(g, set, stack)
	if err != nil {
		return "", nil, err
	}
	return s.Title, dedupeSort(out), nil
}

func expand(g FileCatalog, setKey string, stack map[string]bool) ([]string, error) {
	if stack[setKey] {
		return nil, errors.New("include cycle detected at set: " + setKey)
	}
	stack[setKey] = true
	defer delete(stack, setKey)

	s := g.Sets[setKey]

	out := append([]string{}, s.Tickers...)
	for _, child := range s.Include {
		child = strings.TrimSpace(child)
		if child == "" {
			continue
		}
		if _, ok := g.Sets[child]; !ok {
			return nil, errors.New("include references unknown set: " + child)
		}
		childOut, err := expand(g, child, stack)
		if err != nil {
			return nil, err
		}
		out = append(out, childOut...)
	}
	return out, nil
}

func dedupeSort(in []string) []string {
	seen := map[string]struct{}{}
	out := make([]string, 0, len(in))
	for _, t := range in {
		t = strings.TrimSpace(t)
		if t == "" {
			continue
		}
		if _, ok := seen[t]; ok {
			continue
		}
		seen[t] = struct{}{}
		out = append(out, t)
	}
	sort.Strings(out)
	return out
}

type SetSummary struct {
	ID      string   `json:"id"`
	Title   string   `json:"title"`
	Count   int      `json:"count"`
	Tickers []string `json:"tickers"`
}

func (c *Catalog) ListSummaries() []SetSummary {
	out := make([]SetSummary, 0)

	groupKeys := make([]string, 0, len(c.Groups))
	for gk := range c.Groups {
		groupKeys = append(groupKeys, gk)
	}
	sort.Strings(groupKeys)

	for _, gk := range groupKeys {
		g := c.Groups[gk]

		setKeys := make([]string, 0, len(g.Sets))
		for sk := range g.Sets {
			setKeys = append(setKeys, sk)
		}
		sort.Strings(setKeys)

		for _, sk := range setKeys {
			id := gk + "." + sk

			title, tickers, err := c.Resolve(id)
			if err != nil {
				out = append(out, SetSummary{
					ID:    id,
					Title: g.Sets[sk].Title,
					Count: 0,
				})
				continue
			}

			out = append(out, SetSummary{
				ID:      id,
				Title:   title,
				Count:   len(tickers),
				Tickers: tickers,
			})
		}
	}

	return out
}
