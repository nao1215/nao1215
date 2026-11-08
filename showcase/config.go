package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
)

// Config lists the repositories shown in the README showcase.
type Config struct {
	// Owner is the default owner for repositories written without "owner/".
	Owner    string    `json:"owner"`
	Sections []Section `json:"sections"`
}

// Section is a run of groups. A section with a Summary is folded into a
// <details> block; a section without one is shown as is.
type Section struct {
	Summary string  `json:"summary,omitempty"`
	Groups  []Group `json:"groups"`
}

// Group is one heading followed by its cards.
type Group struct {
	Title string     `json:"title"`
	Repos []RepoSpec `json:"repos"`
}

// RepoSpec is one card. Description and Language override what GitHub
// reports, for example when a game built with gozen is detected as Python
// because of its helper scripts.
type RepoSpec struct {
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	Language    *Language `json:"language,omitempty"`
	Highlight   bool      `json:"highlight,omitempty"`
}

// Language is the label and dot color shown in a card's footer.
type Language struct {
	Name  string `json:"name"`
	Color string `json:"color"`
}

// OwnerAndName splits Name into owner and repository, using defaultOwner when
// Name has no "owner/" prefix.
func (r RepoSpec) OwnerAndName(defaultOwner string) (string, string) {
	if owner, name, ok := strings.Cut(r.Name, "/"); ok {
		return owner, name
	}
	return defaultOwner, r.Name
}

func loadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	return &cfg, nil
}

func (c *Config) validate() error {
	if c.Owner == "" {
		return errors.New("owner is empty")
	}
	seen := map[string]bool{}
	for _, s := range c.Sections {
		for _, g := range s.Groups {
			if g.Title == "" {
				return errors.New("a group has no title")
			}
			for _, r := range g.Repos {
				owner, name := r.OwnerAndName(c.Owner)
				if owner == "" || name == "" || strings.Contains(name, "/") {
					return fmt.Errorf("invalid repository name %q", r.Name)
				}
				if l := r.Language; l != nil && (l.Name == "" || !isHexColor(l.Color)) {
					return fmt.Errorf("%s: language needs a name and a #rrggbb color", r.Name)
				}
				key := strings.ToLower(owner + "/" + name)
				if seen[key] {
					return fmt.Errorf("repository %q is listed twice", r.Name)
				}
				seen[key] = true
			}
		}
	}
	return nil
}

// Repos returns every repository spec in display order.
func (c *Config) Repos() []RepoSpec {
	var out []RepoSpec
	for _, s := range c.Sections {
		for _, g := range s.Groups {
			out = append(out, g.Repos...)
		}
	}
	return out
}

func isHexColor(s string) bool {
	if len(s) != 7 || s[0] != '#' {
		return false
	}
	for _, c := range s[1:] {
		if !strings.ContainsRune("0123456789abcdefABCDEF", c) {
			return false
		}
	}
	return true
}
