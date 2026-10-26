// Package config loads custos.json and derives defaults from composer.json.
package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"custos/internal/diagnostic"
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/meta"
	phpversion "custos/internal/php/version"
	"custos/internal/platform/safeio"
	"custos/internal/semantic/flow"
)

// FileName is the project configuration file.
const FileName = "custos.json"

// MaxFileSize bounds custos.json and composer.json (both come from the
// analysed repository: a huge file, or a symlink to a device, must fail
// instead of exhausting memory or blocking).
const MaxFileSize = 8 << 20

// RuleSettings is the JSON shape of one rule override.
type RuleSettings struct {
	Enabled  *bool          `json:"enabled,omitempty"`
	Severity string         `json:"severity,omitempty"`
	Options  map[string]any `json:"options,omitempty"`
}

// FlowSettings configures optional runtime assumptions and how cross-file
// flow is summarized.
type FlowSettings struct {
	Assertions *bool `json:"assertions,omitempty"`
	Order      *bool `json:"order,omitempty"`
}

// File is the JSON shape of custos.json (also accepted as LSP
// initializationOptions).
type File struct {
	Flow            *FlowSettings           `json:"flow,omitempty"`
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
	Flow flow.Options
	// FlowOrder summarizes cross-file flow in the order the last run saved
	// (see project.Options.FlowOrder).
	FlowOrder       bool
	Root            string // directory of custos.json / composer.json (or cwd)
	PHP             phpversion.Version
	PHPSource       string // where PHP came from: config | composer | .php-version | default
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
	if b, err := safeio.ReadFile(filepath.Join(root, FileName), MaxFileSize); err == nil {
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
	return resolve(root, f, func(name string) ([]byte, error) { return safeio.ReadFile(filepath.Join(root, name), MaxFileSize) })
}

func resolve(root string, f File, read func(string) ([]byte, error)) (*Config, error) {
	c := &Config{Root: root, Paths: f.Paths, Baseline: f.Baseline, Exclude: append(append([]string{}, DefaultExclude...), f.Exclude...), Rules: map[string]analysis.RuleConfig{}}
	// The configuration comes with the analysed code: it must not point
	// custos (and `custos fix`) at files outside the project.
	for _, p := range f.Paths {
		if !insideRoot(p) {
			return nil, fmt.Errorf("config: paths entry %q leaves the project root", p)
		}
	}
	if f.Baseline != "" && !insideRoot(f.Baseline) {
		return nil, fmt.Errorf("config: baseline %q leaves the project root", f.Baseline)
	}
	switch {
	case f.PHP != "":
		v, err := phpversion.Parse(f.PHP)
		if err != nil {
			return nil, err
		}
		c.PHP, c.PHPSource = v, "config"
	default:
		if v, ok := composerPHPRead(read); ok {
			c.PHP, c.PHPSource = v, "composer"
		} else if v, ok := phpVersionFileRead(read); ok {
			c.PHP, c.PHPSource = v, ".php-version"
		} else {
			c.PHP, c.PHPSource = phpversion.Default, "default"
		}
	}
	switch strings.ToLower(f.ComparisonStyle) {
	case "":
		var found bool
		if c.ComparisonStyle, found = csFixerStyleRead(read); !found {
			c.ComparisonStyle, _ = magoStyleRead(read)
		}
	case "regular":
	case "yoda":
		c.ComparisonStyle = analysis.StyleYoda
	default:
		return nil, fmt.Errorf("config: comparisonStyle must be regular or yoda, got %q", f.ComparisonStyle)
	}
	if f.Flow != nil && f.Flow.Assertions != nil {
		c.Flow.Assertions = *f.Flow.Assertions
	}
	if f.Flow != nil && f.Flow.Order != nil {
		c.FlowOrder = *f.Flow.Order
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
			rc.Severity = diagnostic.Severity(rs.Severity)
		default:
			return nil, fmt.Errorf("config: rule %s: invalid severity %q", id, rs.Severity)
		}
		c.Rules[m.ID] = rc
	}
	v := composerPHPUnitRead(read)
	if v == 0 {
		v = schemaPHPUnitRead(read)
	}
	if v > 0 {
		// Fallback for the PHPUnit rules when PHPUnit itself is not
		// indexed (see rules/phpunit detectedPHPUnitVersion).
		for _, id := range []string{"PhpUnitTests", "PhpUnitDeprecations"} {
			rc := c.Rules[id]
			if rc.Options == nil {
				rc.Options = map[string]any{}
			}
			rc.Options["@composerPHPUnit"] = v
			c.Rules[id] = rc
		}
	}
	if len(c.Paths) == 0 {
		c.Paths = []string{"."}
	}
	return c, nil
}

// composerPHPUnit returns the lowest PHPUnit version composer.json allows
// (require-dev, else require), as the PHPUnit rules number versions
// (9.1 → 91; 10 and later → 100); 0 when none is declared.
func composerPHPUnit(root string) int {
	return composerPHPUnitRead(func(name string) ([]byte, error) { return safeio.ReadFile(filepath.Join(root, name), MaxFileSize) })
}

func composerPHPUnitRead(read func(string) ([]byte, error)) int {
	b, err := read("composer.json")
	if err != nil {
		return 0
	}
	var c struct {
		Require    map[string]string `json:"require"`
		RequireDev map[string]string `json:"require-dev"`
	}
	if json.Unmarshal(b, &c) != nil {
		return 0
	}
	constraint := c.RequireDev["phpunit/phpunit"]
	if constraint == "" {
		constraint = c.Require["phpunit/phpunit"]
	}
	best := 0
	for _, m := range unitVersionRe.FindAllStringSubmatch(constraint, -1) {
		major, _ := strconv.Atoi(m[1])
		minor, _ := strconv.Atoi(m[2])
		v := major*10 + min(minor, 9)
		if major >= 10 {
			v = 100
		}
		if best == 0 || v < best {
			best = v
		}
	}
	return best
}

var unitVersionRe = regexp.MustCompile(`(\d+)(?:\.(\d+))?`)

// schemaPHPUnitRead returns the PHPUnit version that phpunit.xml (else
// phpunit.xml.dist) names in its schema URL, numbered as composerPHPUnit
// does; 0 when neither names one. A project whose PHPUnit lives outside
// vendor/ (symfony/phpunit-bridge) still declares the schema it targets.
func schemaPHPUnitRead(read func(string) ([]byte, error)) int {
	for _, name := range []string{"phpunit.xml", "phpunit.xml.dist"} {
		b, err := read(name)
		if err != nil {
			continue
		}
		if m := schemaVersionRe.FindSubmatch(b); m != nil {
			major, _ := strconv.Atoi(string(m[1]))
			minor, _ := strconv.Atoi(string(m[2]))
			if major >= 10 {
				return 100
			}
			return major*10 + min(minor, 9)
		}
	}
	return 0
}

var schemaVersionRe = regexp.MustCompile(`schema\.phpunit\.de/(\d+)\.(\d+)/phpunit\.xsd`)

// insideRoot reports whether the relative path p stays within the root
// directory (no absolute path, no leading "..", no volume name).
func insideRoot(p string) bool {
	if filepath.IsAbs(p) || filepath.VolumeName(p) != "" || strings.HasPrefix(p, "/") || strings.HasPrefix(p, `\`) {
		return false
	}
	c := filepath.ToSlash(filepath.Clean(p))
	return c != ".." && !strings.HasPrefix(c, "../")
}

// Analysis returns the analysis.Config view.
func (c *Config) Analysis() analysis.Config {
	return analysis.Config{Flow: c.Flow, PHP: c.PHP, ComparisonStyle: c.ComparisonStyle, Rules: c.Rules}
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

// composerPHPRead reads config.platform.php, else the lowest version in require.php.
// phpVersionFileRead reads the version that `.php-version` (phpenv, the
// Symfony CLI and similar tools) pins: its first line, `8.2` or `8.2.12`.
func phpVersionFileRead(read func(string) ([]byte, error)) (phpversion.Version, bool) {
	b, err := read(".php-version")
	if err != nil {
		return 0, false
	}
	line, _, _ := strings.Cut(string(b), "\n")
	if strings.TrimSpace(line) == "" {
		return 0, false
	}
	v, err := phpversion.Parse(line)
	return v, err == nil
}

func composerPHPRead(read func(string) ([]byte, error)) (phpversion.Version, bool) {
	b, err := read("composer.json")
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
func lowestVersion(constraint string) (phpversion.Version, bool) {
	var best phpversion.Version
	for _, m := range versionRe.FindAllStringSubmatch(constraint, -1) {
		v, err := phpversion.Parse(m[1] + "." + m[2])
		if err != nil {
			continue
		}
		if best == 0 || v < best {
			best = v
		}
	}
	return best, best != 0
}
