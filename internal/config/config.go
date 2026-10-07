// Package config loads custos.json and derives defaults from composer.json.
package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"custos/internal/analysis"
	"custos/internal/meta"
	"custos/internal/phpver"
)

// FileName is the project configuration file.
const FileName = "custos.json"

// RuleSettings is the JSON shape of one rule override.
type RuleSettings struct {
	Enabled  *bool          `json:"enabled,omitempty"`
	Severity string         `json:"severity,omitempty"`
	Options  map[string]any `json:"options,omitempty"`
}

// File is the JSON shape of custos.json (also accepted as LSP
// initializationOptions).
type File struct {
	PHP             string                  `json:"php,omitempty"`
	ComparisonStyle string                  `json:"comparisonStyle,omitempty"` // regular | yoda
	ShortOpenTag    *bool                   `json:"shortOpenTag,omitempty"`
	Paths           []string                `json:"paths,omitempty"`
	Exclude         []string                `json:"exclude,omitempty"`
	Baseline        string                  `json:"baseline,omitempty"` // baseline file, relative to the root
	Rules           map[string]RuleSettings `json:"rules,omitempty"`
}

// Config is the resolved configuration.
type Config struct {
	Root            string // directory of custos.json / composer.json (or cwd)
	PHP             phpver.Version
	PHPSource       string // where PHP came from: config | composer | default
	ComparisonStyle analysis.ComparisonStyle
	ShortOpenTag    bool
	Paths           []string
	Exclude         []string
	Baseline        string
	Rules           map[string]analysis.RuleConfig
}

// DefaultExclude are directory names skipped during discovery.
var DefaultExclude = []string{"vendor", "node_modules", ".git", ".idea", ".custos", "var/cache"}

// Load finds custos.json starting at dir and walking up; composer.json in
// the same root supplies the PHP version when the config does not.
func Load(dir string) (*Config, error) {
	root, err := findRoot(dir)
	if err != nil {
		return nil, err
	}
	var f File
	if b, err := os.ReadFile(filepath.Join(root, FileName)); err == nil {
		if err := json.Unmarshal(b, &f); err != nil {
			return nil, fmt.Errorf("%s: %w", filepath.Join(root, FileName), err)
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return nil, err
	}
	return Resolve(root, f)
}

// Resolve turns a File into a Config rooted at root.
func Resolve(root string, f File) (*Config, error) {
	c := &Config{Root: root, Paths: f.Paths, Baseline: f.Baseline, Exclude: append(append([]string{}, DefaultExclude...), f.Exclude...), Rules: map[string]analysis.RuleConfig{}}
	switch {
	case f.PHP != "":
		v, err := phpver.Parse(f.PHP)
		if err != nil {
			return nil, err
		}
		c.PHP, c.PHPSource = v, "config"
	default:
		if v, ok := composerPHP(root); ok {
			c.PHP, c.PHPSource = v, "composer"
		} else {
			c.PHP, c.PHPSource = phpver.Default, "default"
		}
	}
	switch strings.ToLower(f.ComparisonStyle) {
	case "", "regular":
	case "yoda":
		c.ComparisonStyle = analysis.StyleYoda
	default:
		return nil, fmt.Errorf("config: comparisonStyle must be regular or yoda, got %q", f.ComparisonStyle)
	}
	if f.ShortOpenTag != nil {
		c.ShortOpenTag = *f.ShortOpenTag
	}
	for id, rs := range f.Rules {
		m, ok := meta.Lookup(id)
		if !ok {
			return nil, fmt.Errorf("config: unknown rule %q", id)
		}
		rc := analysis.RuleConfig{Enabled: rs.Enabled, Options: rs.Options}
		switch rs.Severity {
		case "":
		case "error", "warning", "info":
			rc.Severity = meta.Severity(rs.Severity)
		default:
			return nil, fmt.Errorf("config: rule %s: invalid severity %q", id, rs.Severity)
		}
		c.Rules[m.ID] = rc
	}
	if len(c.Paths) == 0 {
		c.Paths = []string{"."}
	}
	return c, nil
}

// Analysis returns the analysis.Config view.
func (c *Config) Analysis() analysis.Config {
	return analysis.Config{PHP: c.PHP, ComparisonStyle: c.ComparisonStyle, Rules: c.Rules}
}

func findRoot(dir string) (string, error) {
	dir, err := filepath.Abs(dir)
	if err != nil {
		return "", err
	}
	for d := dir; ; d = filepath.Dir(d) {
		for _, marker := range []string{FileName, "composer.json"} {
			if _, err := os.Stat(filepath.Join(d, marker)); err == nil {
				return d, nil
			}
		}
		if parent := filepath.Dir(d); parent == d {
			return dir, nil
		}
	}
}

var versionRe = regexp.MustCompile(`(\d+)\.(\d+)`)

// composerPHP reads config.platform.php, else the lowest version in require.php.
func composerPHP(root string) (phpver.Version, bool) {
	b, err := os.ReadFile(filepath.Join(root, "composer.json"))
	if err != nil {
		return 0, false
	}
	var c struct {
		Require map[string]string `json:"require"`
		Config  struct {
			Platform map[string]string `json:"platform"`
		} `json:"config"`
	}
	if json.Unmarshal(b, &c) != nil {
		return 0, false
	}
	for _, s := range []string{c.Config.Platform["php"], c.Require["php"]} {
		if s == "" {
			continue
		}
		if v, ok := lowestVersion(s); ok {
			return v, true
		}
	}
	return 0, false
}

// lowestVersion returns the smallest supported version mentioned in a
// composer constraint such as "^7.4 || ^8.0" or ">=8.1".
func lowestVersion(constraint string) (phpver.Version, bool) {
	var best phpver.Version
	for _, m := range versionRe.FindAllStringSubmatch(constraint, -1) {
		v, err := phpver.Parse(m[1] + "." + m[2])
		if err != nil {
			continue
		}
		if best == 0 || v < best {
			best = v
		}
	}
	return best, best != 0
}
