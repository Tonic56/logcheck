// Package config handles loading and merging of logcheck.yaml configuration.
package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Rule name constants — used as keys in the Disabled map and in tests.
const (
	RuleLowercase   = "lowercase"
	RuleEnglishOnly = "english-only"
	RuleNoSpecial   = "no-special-chars"
	RuleNoSensitive = "no-sensitive-data"
)

// Config is the in-memory representation of logcheck.yaml.
type Config struct {
	// Disabled is a set of rule names that should be skipped.

	Disabled []string `yaml:"disabled"`

	// ExtraKeywords extends the built-in sensitive-data keyword list.

	ExtraKeywords []string `yaml:"extra_keywords"`

	// AllowedChars lists ASCII characters that the no-special-chars rule

	AllowedChars string `yaml:"allowed_chars"`

	// CustomPatterns holds user-defined regexp patterns for the

	CustomPatterns []string `yaml:"custom_patterns"`

	// disabledSet is an O(1) lookup set built from Disabled.
	disabledSet map[string]struct{}
}

// Defaults returns a Config with all rules enabled and no extra keywords.
func Defaults() *Config {
	c := &Config{}
	c.buildIndex()
	return c
}

// Load reads logcheck.yaml from path, falls back to Defaults() when the file
// does not exist, and returns an error for any other I/O or parse failure.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return Defaults(), nil
		}
		return nil, fmt.Errorf("logcheck: reading config %q: %w", path, err)
	}

	var c Config
	if err := yaml.Unmarshal(data, &c); err != nil {
		return nil, fmt.Errorf("logcheck: parsing config %q: %w", path, err)
	}

	c.buildIndex()
	return &c, nil
}

// IsEnabled reports whether ruleName should run.
func (c *Config) IsEnabled(ruleName string) bool {
	if c.disabledSet == nil {
		return true
	}
	_, off := c.disabledSet[ruleName]
	return !off
}

// SetDisabled replaces the disabled list and rebuilds the lookup index.
// Intended for use in tests.
func (c *Config) SetDisabled(names []string) {
	c.Disabled = names
	c.buildIndex()
}

func (c *Config) buildIndex() {
	c.disabledSet = make(map[string]struct{}, len(c.Disabled))
	for _, name := range c.Disabled {
		c.disabledSet[name] = struct{}{}
	}
}
