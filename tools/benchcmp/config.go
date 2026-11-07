//go:build unix

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// Config is bench/competitors.json: what is measured and at which versions.
type Config struct {
	Runs     int               `json:"runs"`
	Warmup   int               `json:"warmup"`
	Timeout  string            `json:"timeout"`
	Corpora  []Corpus          `json:"corpora"`
	Packages map[string]string `json:"packages"`
	Tools    []Tool            `json:"tools"`
}

// Corpus is a project analyzed at a pinned tag.
type Corpus struct {
	Name  string   `json:"name"`
	Repo  string   `json:"repo"`
	Tag   string   `json:"tag"`
	Paths []string `json:"paths"`
}

// Tool is one measured command. Package names a Composer package of
// Packages ("" for the custos binary under test).
type Tool struct {
	ID       string            `json:"id"`
	Label    string            `json:"label"`
	Kind     string            `json:"kind"`
	Package  string            `json:"package,omitempty"`
	Cmd      []string          `json:"cmd"`
	Env      map[string]string `json:"env,omitempty"`
	Findings string            `json:"findings"`
	// Stream is the output holding the findings: "stdout" (default) or
	// "stderr".
	Stream string `json:"stream,omitempty"`
}

func loadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var c Config
	if err := json.Unmarshal(data, &c); err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	if c.Runs < 1 || c.Warmup < 0 {
		return nil, fmt.Errorf("%s: runs must be at least 1 and warmup at least 0", path)
	}
	if _, err := c.timeout(); err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	for _, t := range c.Tools {
		if _, ok := parsers[t.Findings]; !ok {
			return nil, fmt.Errorf("%s: tool %s: unknown findings parser %q", path, t.ID, t.Findings)
		}
		if _, ok := c.Packages[t.Package]; t.Package != "" && !ok {
			return nil, fmt.Errorf("%s: tool %s: package %s is not pinned", path, t.ID, t.Package)
		}
	}
	return &c, nil
}

func (c *Config) timeout() (time.Duration, error) { return time.ParseDuration(c.Timeout) }

// writeJSON writes v indented, with a final newline.
func writeJSON(path string, v any) error {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetIndent("", "  ")
	enc.SetEscapeHTML(false)
	_ = enc.Encode(v) // only marshals our own types
	return os.WriteFile(path, buf.Bytes(), 0o644)
}
