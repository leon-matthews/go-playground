package common

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		cfg, err := LoadConfig("testdata/config.json")
		if err != nil {
			t.Fatal(err)
		}

		if len(cfg.Prefixes) != 2 {
			t.Fatalf("got %d prefixes, want 2", len(cfg.Prefixes))
		}
		// Should be lowercased
		if cfg.Prefixes[0].Text != "walmart" {
			t.Errorf("prefix[0].Text = %q, want %q", cfg.Prefixes[0].Text, "walmart")
		}
		if cfg.Prefixes[0].Category != "Shopping" {
			t.Errorf("prefix[0].Category = %q, want %q", cfg.Prefixes[0].Category, "Shopping")
		}

		sc, ok := cfg.Sources["anz"]
		if !ok {
			t.Fatal("missing anz source config")
		}
		if len(sc.Delete) != 1 || sc.Delete[0] != "4835-****-****-9371" {
			t.Errorf("anz delete = %v, want [4835-****-****-9371]", sc.Delete)
		}
	})

	t.Run("empty object", func(t *testing.T) {
		cfg, err := LoadConfig("testdata/config_empty.json")
		if err != nil {
			t.Fatal(err)
		}
		if len(cfg.Prefixes) != 0 {
			t.Fatalf("got %d prefixes, want 0", len(cfg.Prefixes))
		}
	})

	t.Run("file not found", func(t *testing.T) {
		_, err := LoadConfig("testdata/nonexistent.json")
		if err == nil {
			t.Fatal("expected error for missing file")
		}
	})

	t.Run("malformed json", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "bad.json")
		os.WriteFile(path, []byte("{bad"), 0644)
		_, err := LoadConfig(path)
		if err == nil {
			t.Fatal("expected error for malformed JSON")
		}
	})

	t.Run("lowercases and trims prefix text", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "trim.json")
		os.WriteFile(path, []byte(`{"prefixes": [{"prefix": "  WALMART  ", "category": " Shopping "}]}`), 0644)
		cfg, err := LoadConfig(path)
		if err != nil {
			t.Fatal(err)
		}
		if cfg.Prefixes[0].Text != "walmart" {
			t.Errorf("Text = %q, want %q", cfg.Prefixes[0].Text, "walmart")
		}
		if cfg.Prefixes[0].Category != "Shopping" {
			t.Errorf("Category = %q, want %q", cfg.Prefixes[0].Category, "Shopping")
		}
	})
}

func TestLoadConfigValidation(t *testing.T) {
	writeJSON := func(t *testing.T, content string) string {
		t.Helper()
		path := filepath.Join(t.TempDir(), "config.json")
		os.WriteFile(path, []byte(content), 0644)
		return path
	}

	t.Run("empty prefix text", func(t *testing.T) {
		path := writeJSON(t, `{"prefixes": [{"prefix": "", "category": "Shopping"}]}`)
		_, err := LoadConfig(path)
		if err == nil {
			t.Fatal("expected error for empty prefix text")
		}
		if !strings.Contains(err.Error(), "empty prefix text") {
			t.Errorf("error = %q, want mention of empty prefix text", err)
		}
	})

	t.Run("empty category", func(t *testing.T) {
		path := writeJSON(t, `{"prefixes": [{"prefix": "walmart", "category": ""}]}`)
		_, err := LoadConfig(path)
		if err == nil {
			t.Fatal("expected error for empty category")
		}
		if !strings.Contains(err.Error(), "empty category") {
			t.Errorf("error = %q, want mention of empty category", err)
		}
	})

	t.Run("duplicate prefix", func(t *testing.T) {
		path := writeJSON(t, `{"prefixes": [
			{"prefix": "walmart", "category": "Shopping"},
			{"prefix": "walmart", "category": "Retail"}
		]}`)
		_, err := LoadConfig(path)
		if err == nil {
			t.Fatal("expected error for duplicate prefix")
		}
		if !strings.Contains(err.Error(), "duplicate prefix") {
			t.Errorf("error = %q, want mention of duplicate prefix", err)
		}
	})

	t.Run("reports all errors at once", func(t *testing.T) {
		path := writeJSON(t, `{"prefixes": [
			{"prefix": "", "category": ""},
			{"prefix": "walmart", "category": "Shopping"},
			{"prefix": "walmart", "category": "Retail"}
		]}`)
		_, err := LoadConfig(path)
		if err == nil {
			t.Fatal("expected error")
		}
		// Should contain all three issues: empty text, empty category, duplicate
		s := err.Error()
		if !strings.Contains(s, "empty prefix text") {
			t.Errorf("missing 'empty prefix text' in %q", s)
		}
		if !strings.Contains(s, "empty category") {
			t.Errorf("missing 'empty category' in %q", s)
		}
		if !strings.Contains(s, "duplicate prefix") {
			t.Errorf("missing 'duplicate prefix' in %q", s)
		}
	})
}

func TestSaveConfig(t *testing.T) {
	t.Run("round-trip", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "config.json")
		cfg := &Config{
			Prefixes: []Prefix{
				{Text: "walmart", Category: "Shopping"},
				{Text: "amazon", Category: "Online"},
			},
			Sources: map[string]SourceConfig{
				"anz": {Delete: []string{"noise"}},
			},
		}

		if err := SaveConfig(path, cfg); err != nil {
			t.Fatal(err)
		}

		got, err := LoadConfig(path)
		if err != nil {
			t.Fatal(err)
		}

		// SaveConfig sorts, so expect alphabetical order
		if got.Prefixes[0].Text != "amazon" || got.Prefixes[1].Text != "walmart" {
			t.Errorf("prefixes not sorted: %+v", got.Prefixes)
		}

		if got.Sources["anz"].Delete[0] != "noise" {
			t.Errorf("sources not preserved: %+v", got.Sources)
		}
	})

	t.Run("does not mutate input", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "config.json")
		cfg := &Config{
			Prefixes: []Prefix{
				{Text: "walmart", Category: "Shopping"},
				{Text: "amazon", Category: "Online"},
			},
		}

		if err := SaveConfig(path, cfg); err != nil {
			t.Fatal(err)
		}

		if cfg.Prefixes[0].Text != "walmart" || cfg.Prefixes[1].Text != "amazon" {
			t.Errorf("input was mutated: %+v", cfg.Prefixes)
		}
	})

	t.Run("unwritable path", func(t *testing.T) {
		err := SaveConfig("/no/such/directory/config.json", &Config{})
		if err == nil {
			t.Fatal("expected error for unwritable path")
		}
	})
}

func TestCleanDetails(t *testing.T) {
	t.Run("removes single string", func(t *testing.T) {
		r := NewReplacer([]string{"4835-****-****-9371"})
		got := CleanDetails("Acme Corp 4835-****-****-9371 Auckland", r)
		if got != "Acme Corp Auckland" {
			t.Errorf("got %q, want %q", got, "Acme Corp Auckland")
		}
	})

	t.Run("removes multiple strings", func(t *testing.T) {
		r := NewReplacer([]string{"noise1", "noise2"})
		got := CleanDetails("Acme Corp noise1 noise2 Auckland", r)
		if got != "Acme Corp Auckland" {
			t.Errorf("got %q, want %q", got, "Acme Corp Auckland")
		}
	})

	t.Run("no-op when removals empty", func(t *testing.T) {
		r := NewReplacer(nil)
		got := CleanDetails("Acme Corp", r)
		if got != "Acme Corp" {
			t.Errorf("got %q, want %q", got, "Acme Corp")
		}
	})

	t.Run("no-op when no matches", func(t *testing.T) {
		r := NewReplacer([]string{"xyz"})
		got := CleanDetails("Acme Corp", r)
		if got != "Acme Corp" {
			t.Errorf("got %q, want %q", got, "Acme Corp")
		}
	})
}
