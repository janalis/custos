// Package conformance checks rule implementations against annotated fixtures
// (see package markup for their format).
package conformance

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/janalis/custos/internal/diagnostic"
	"github.com/janalis/custos/internal/diagnostic/markup"
)

// Expectation is one expected finding of a fixture (see package markup).
type Expectation = markup.Expectation

var (
	// ParseMarkup strips the expectation tags of a fixture.
	ParseMarkup = markup.Parse
	// SortExpectations orders expectations by position.
	SortExpectations = markup.Sort
)

// Request is one analysis run over a fixture.
type Request struct {
	Path            string
	Source          []byte
	Rules           []string          // rule IDs to enable (all others off)
	PHP             string            // target PHP version, "" = default
	ComparisonStyle string            // "", "regular", "yoda"
	Options         map[string]string // "Rule.OPTION" -> raw value
	Calls           []string          // raw option calls (EA list options)
	// Companions are extra project files (path -> clean source) whose
	// symbols are visible to the analyzed file.
	Companions map[string][]byte
	// CompleteWorld marks the indexed files as everything the project can
	// load (see index.World).
	CompleteWorld bool
	// Vendor is the project's vendor directory (see index.World.Vendor).
	Vendor string
	// Extensions are the PHP extensions the project requires (see
	// index.World.Extensions).
	Extensions []string
	// Assertions trusts assert() calls as guards (flow.assertions).
	Assertions bool
}

// Finding is one reported problem, as seen by the harness.
type Finding struct {
	Rule     string
	Severity diagnostic.Severity
	Message  string
	Start    int
	End      int
}

// Engine is implemented by the analysis engine adapter.
type Engine interface {
	// Supports reports whether every given rule is implemented.
	Supports(rules []string) bool
	// Analyze returns findings for req.
	Analyze(req Request) ([]Finding, error)
	// Fix applies the available fixes in one pass and returns the result.
	Fix(req Request) ([]byte, error)
}

// EACase mirrors tools/extract's Case.
type EACase struct {
	Test            string            `json:"test"`
	Rules           []string          `json:"rules"`
	Fixture         string            `json:"fixture"`
	Fixed           string            `json:"fixed,omitempty"`
	PHP             string            `json:"php,omitempty"`
	ComparisonStyle string            `json:"comparisonStyle,omitempty"`
	Options         map[string]string `json:"options,omitempty"`
	Calls           []string          `json:"calls,omitempty"`
	Companions      []string          `json:"companions,omitempty"`
}

// EAIndex mirrors the parts of tools/extract's Index the harness needs.
type EAIndex struct {
	EAPath string   `json:"eaPath"`
	Cases  []EACase `json:"cases"`
}

// LoadEAIndex reads the local-only index produced by `make extract`.
func LoadEAIndex(path string) (*EAIndex, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var idx EAIndex
	if err := json.Unmarshal(b, &idx); err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	return &idx, nil
}

// Result of checking one case.
type Result struct {
	Missing    []Expectation // expected but not reported
	Unexpected []Expectation // reported but not expected
	FixDiff    string        // non-empty when the fixed output differs
}

// OK reports whether the case passed.
func (r Result) OK() bool {
	return len(r.Missing) == 0 && len(r.Unexpected) == 0 && r.FixDiff == ""
}

// CompareOptions selects what is compared.
type CompareOptions struct {
	Messages bool // compare message text (own fixtures only; EA text is never compared)
	// OnlyPrefixed keeps only expectations whose message has this prefix
	// (EA: "[EA] "), dropping IDE syntax-error markers.
	OnlyPrefixed string
	// HideExpected suppresses expected fix text in diffs (EA fixtures).
	HideExpected bool
	// NoFixWithoutFixed requires a fixture without expected fix output to
	// be left as it is by the fixes (own fixtures: a quick-fix is always
	// shown by a .fixed file).
	NoFixWithoutFixed bool
}

// Check runs engine on a marked-up fixture and compares the outcome.
// fixed may be nil when the case has no expected fix output.
func Check(engine Engine, req Request, marked, fixed []byte, opt CompareOptions) (Result, error) {
	clean, want, err := ParseMarkup(marked)
	if err != nil {
		return Result{}, err
	}
	req.Source = clean
	if opt.OnlyPrefixed != "" {
		// EA fixtures also contain IDE parse-error markers ("Expected: …");
		// only inspection results carry the prefix.
		kept := want[:0]
		for _, w := range want {
			if w.Message == "" || strings.HasPrefix(w.Message, opt.OnlyPrefixed) {
				kept = append(kept, w)
			}
		}
		want = kept
	}
	findings, err := engine.Analyze(req)
	if err != nil {
		return Result{}, err
	}
	got := make([]Expectation, 0, len(findings))
	for _, f := range findings {
		got = append(got, Expectation{Severity: f.Severity, Message: f.Message, Start: f.Start, End: f.End})
	}
	SortExpectations(got)
	var res Result
	res.Missing, res.Unexpected = markup.Diff(want, got, opt.Messages)
	if fixed == nil && opt.NoFixWithoutFixed && len(got) > 0 {
		out, err := engine.Fix(req)
		if err != nil {
			return res, err
		}
		if !bytes.Equal(out, clean) {
			res.FixDiff = fmt.Sprintf("a fix changes the fixture, which has no .fixed file to show it:\n%s", out)
		}
	}
	if fixed != nil {
		out, err := engine.Fix(req)
		if err != nil {
			return res, err
		}
		if g, w := markup.NormalizeWS(out), markup.NormalizeWS(fixed); !bytes.Equal(g, w) {
			if opt.HideExpected {
				// Clean room: never reveal upstream fixture text to implementers;
				// only say where custos' (whitespace-collapsed) output diverges.
				i := 0
				for i < len(g) && i < len(w) && g[i] == w[i] {
					i++
				}
				res.FixDiff = fmt.Sprintf("custos output diverges from the expected fix at collapsed offset %d; custos produced: %q", i, excerpt(g, i))
			} else {
				res.FixDiff = fmt.Sprintf("--- want\n%s\n--- got\n%s", fixed, out)
			}
		}
	}
	return res, nil
}

// ReadEA reads a file of the EA checkout (never copied into the repo).
func ReadEA(idx *EAIndex, rel string) ([]byte, error) {
	return os.ReadFile(filepath.Join(idx.EAPath, rel))
}

// excerpt returns a short window of b around offset i.
func excerpt(b []byte, i int) string {
	start := max(0, i-40)
	end := min(len(b), i+60)
	return string(b[start:end])
}
