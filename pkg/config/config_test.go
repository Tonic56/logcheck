package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Tonic56/logcheck/pkg/config"
)

func TestDefaults_AllEnabled(t *testing.T) {
	c := config.Defaults()
	for _, name := range []string{
		config.RuleLowercase,
		config.RuleEnglishOnly,
		config.RuleNoSpecial,
		config.RuleNoSensitive,
	} {
		if !c.IsEnabled(name) {
			t.Errorf("rule %q should be enabled by default", name)
		}
	}
}

func TestLoad_DisableRule(t *testing.T) {
	dir := t.TempDir()
	yamlContent := `
disabled:
  - no-sensitive-data
`
	path := filepath.Join(dir, "logcheck.yaml")
	if err := os.WriteFile(path, []byte(yamlContent), 0600); err != nil {
		t.Fatal(err)
	}

	c, err := config.Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if c.IsEnabled(config.RuleNoSensitive) {
		t.Error("no-sensitive-data should be disabled")
	}
	if !c.IsEnabled(config.RuleLowercase) {
		t.Error("lowercase should still be enabled")
	}
}

func TestLoad_ExtraKeywords(t *testing.T) {
	dir := t.TempDir()
	yamlContent := `
extra_keywords:
  - internal_token
  - priv_cert
`
	path := filepath.Join(dir, "logcheck.yaml")
	if err := os.WriteFile(path, []byte(yamlContent), 0600); err != nil {
		t.Fatal(err)
	}

	c, err := config.Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	found := false
	for _, kw := range c.ExtraKeywords {
		if kw == "internal_token" {
			found = true
		}
	}
	if !found {
		t.Error("extra keyword 'internal_token' not loaded")
	}
}

func TestLoad_AllowedChars(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "logcheck.yaml")
	if err := os.WriteFile(path, []byte("allowed_chars: \"!\"\n"), 0600); err != nil {
		t.Fatal(err)
	}
	c, err := config.Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if c.AllowedChars != "!" {
		t.Errorf("AllowedChars = %q; want %q", c.AllowedChars, "!")
	}
}

func TestLoad_MissingFile(t *testing.T) {
	c, err := config.Load("/nonexistent/logcheck.yaml")
	if err != nil {
		t.Fatalf("Load with missing file should not error, got: %v", err)
	}
	if !c.IsEnabled(config.RuleLowercase) {
		t.Error("defaults should be returned for missing file")
	}
}
