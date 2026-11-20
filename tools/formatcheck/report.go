package main

import (
	"cmp"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path"
	"slices"
	"strings"
	"text/tabwriter"
)

// divergence is a reference rule custos deliberately does not satisfy. Rule
// is a name or a path.Match pattern (`phpdoc_*`); File, when set, limits it
// to the findings in files whose path ends with it, for a reference rule
// that is right elsewhere.
type divergence struct {
	Rule   string `json:"rule"`
	File   string `json:"file,omitempty"`
	Reason string `json:"reason"`
}

// ruleWide returns the divergences that hold in every file.
func ruleWide(div []divergence) []divergence {
	var out []divergence
	for _, d := range div {
		if d.File == "" {
			out = append(out, d)
		}
	}
	return out
}

// inFile reports whether a divergence limited to a file covers f. A
// finding of Pint names every fixer it ran on the file, some with no net
// change: one of them is enough.
func inFile(div []divergence, f finding) bool {
	for _, d := range div {
		if d.File == "" || !strings.HasSuffix(f.File, d.File) {
			continue
		}
		for rule := range strings.SplitSeq(f.Rule, ",") {
			if ok, _ := path.Match(d.Rule, rule); ok {
				return true
			}
		}
	}
	return false
}

func isPattern(rule string) bool { return strings.ContainsAny(rule, `*?[\`) }

// loadDivergences reads the divergences of each preset.
func loadDivergences(file string) (map[string][]divergence, error) {
	data, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}
	var div map[string][]divergence
	if err := json.Unmarshal(data, &div); err != nil {
		return nil, fmt.Errorf("%s: %w", file, err)
	}
	for preset, ds := range div {
		if _, ok := checkers[preset]; !ok {
			return nil, fmt.Errorf("%s: unknown preset %q", file, preset)
		}
		for _, d := range ds {
			if _, err := path.Match(d.Rule, ""); err != nil || d.Rule == "" || d.Reason == "" {
				return nil, fmt.Errorf("%s: %s: each divergence needs a valid rule and a reason: %+v", file, preset, d)
			}
		}
	}
	return div, nil
}

// divergent returns the reason why rule is a divergence ("" when it is not).
func divergent(div []divergence, rule string) string {
	for _, d := range ruleWide(div) {
		if ok, _ := path.Match(d.Rule, rule); ok { // patterns are checked on load
			return d.Reason
		}
	}
	return ""
}

// ruleCount is the number of findings of a reference rule before and after
// formatting.
type ruleCount struct {
	Rule      string `json:"rule"`
	Before    int    `json:"before"`
	After     int    `json:"after"`
	Divergent string `json:"divergent,omitempty"` // the reason, for a divergence
}

// outcome is the result of a preset on a corpus. Remaining is the number
// of findings left after formatting outside the divergences, listed in
// Findings.
type outcome struct {
	Corpus    string      `json:"corpus"`
	Preset    string      `json:"preset"`
	Checker   string      `json:"checker"`
	Before    int         `json:"before"`
	After     int         `json:"after"`
	Divergent int         `json:"divergent"`
	Remaining int         `json:"remaining"`
	Rules     []ruleCount `json:"rules"`
	Findings  []finding   `json:"findings"`
}

// summarize counts the findings by rule. Without a gate, the findings after
// formatting that are not divergences remain; with one, gated does.
func summarize(corpus, preset string, ck checker, before, after, gated []finding, div []divergence) outcome {
	oc := outcome{Corpus: corpus, Preset: preset, Checker: ck.label, Before: len(before), After: len(after), Findings: []finding{}}
	counts := map[string]*ruleCount{}
	count := func(rule string) *ruleCount {
		if counts[rule] == nil {
			counts[rule] = &ruleCount{Rule: rule, Divergent: divergent(div, rule)}
		}
		return counts[rule]
	}
	for _, f := range before {
		count(f.Rule).Before++
	}
	for _, f := range after {
		rc := count(f.Rule)
		rc.After++
		switch {
		case rc.Divergent != "" || inFile(div, f):
			oc.Divergent++
		case ck.gate == nil:
			oc.Findings = append(oc.Findings, f)
		}
	}
	if ck.gate != nil {
		for _, f := range gated {
			if !inFile(div, f) {
				oc.Findings = append(oc.Findings, f)
			}
		}
	}
	oc.Remaining = len(oc.Findings)
	for _, rc := range counts {
		oc.Rules = append(oc.Rules, *rc)
	}
	slices.SortFunc(oc.Rules, func(a, b ruleCount) int {
		return cmp.Or(cmp.Compare(b.After, a.After), cmp.Compare(b.Before, a.Before), strings.Compare(a.Rule, b.Rule))
	})
	slices.SortFunc(oc.Findings, func(a, b finding) int {
		return cmp.Or(strings.Compare(a.Rule, b.Rule), strings.Compare(a.File, b.File), cmp.Compare(a.Line, b.Line))
	})
	return oc
}

// printText writes the table of outcomes, then the remaining findings of
// each, by rule, with up to examples of each rule.
func printText(w io.Writer, outcomes []outcome, examples int) {
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "corpus\tpreset\tchecker\tbefore\tafter\tdivergent\tremaining")
	var notes []string
	for _, oc := range outcomes {
		fmt.Fprintf(tw, "%s\t%s\t%s\t%d\t%d\t%d\t%d\n", oc.Corpus, oc.Preset, oc.Checker, oc.Before, oc.After, oc.Divergent, oc.Remaining)
		if note := checkers[oc.Preset].note; note != "" && !slices.Contains(notes, note) {
			notes = append(notes, note)
		}
	}
	_ = tw.Flush() // writes to w, whose errors the caller cannot act on
	if len(notes) > 0 {
		fmt.Fprintf(w, "\n%s\n", strings.Join(notes, "\n"))
	}
	for _, oc := range outcomes {
		if oc.Remaining == 0 {
			continue
		}
		fmt.Fprintf(w, "\n%s %s: %d remaining\n", oc.Corpus, oc.Preset, oc.Remaining)
		for fs := oc.Findings; len(fs) > 0; {
			n := 1
			for n < len(fs) && fs[n].Rule == fs[0].Rule {
				n++
			}
			fmt.Fprintf(w, "  %s: %d\n", fs[0].Rule, n)
			for _, f := range fs[:min(examples, n)] {
				fmt.Fprintf(w, "    %s:%d: %s\n", f.File, f.Line, f.Message)
			}
			fs = fs[n:]
		}
	}
}
