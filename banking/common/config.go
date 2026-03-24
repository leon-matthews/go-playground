package common

import (
	"encoding/json"
	"fmt"
	"os"
	"slices"
	"strings"
)

// ComparePrefix orders prefixes by text for binary search.
func ComparePrefix(a, b Prefix) int {
	return strings.Compare(a.Text, b.Text)
}

// LoadConfig reads a JSON config file.
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	for i := range cfg.Prefixes {
		cfg.Prefixes[i].Text = strings.ToLower(strings.TrimSpace(cfg.Prefixes[i].Text))
		cfg.Prefixes[i].Category = strings.TrimSpace(cfg.Prefixes[i].Category)
	}

	return &cfg, nil
}

// SaveConfig sorts prefixes and writes the config as pretty-printed JSON.
func SaveConfig(path string, cfg *Config) error {
	out := *cfg
	out.Prefixes = make([]Prefix, len(cfg.Prefixes))
	copy(out.Prefixes, cfg.Prefixes)
	slices.SortFunc(out.Prefixes, ComparePrefix)

	data, err := json.MarshalIndent(&out, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}
	data = append(data, '\n')

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("write config: %w", err)
	}
	return nil
}

// NewReplacer builds a strings.Replacer that removes each literal string.
func NewReplacer(removals []string) *strings.Replacer {
	pairs := make([]string, 0, len(removals)*2)
	for _, r := range removals {
		pairs = append(pairs, r, "")
	}
	return strings.NewReplacer(pairs...)
}

// CleanDetails removes each literal string in removals from details,
// then re-normalizes whitespace.
func CleanDetails(details string, replacer *strings.Replacer) string {
	return CleanString(replacer.Replace(details))
}
