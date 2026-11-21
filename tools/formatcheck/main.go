// Command formatcheck measures how far the output of custos format is from
// the style checkers its presets follow. For each pinned benchmark corpus it
// copies the corpus's source into a temporary directory, formats a copy with
// each preset, and runs the preset's reference checker on the unformatted
// copy (before) and on the formatted one (after):
//
//	psr12    PHP_CodeSniffer, standard PSR12
//	per      PHP-CS-Fixer, rule set @PER-CS
//	symfony  PHP-CS-Fixer, rule set @Symfony
//	laravel  Laravel Pint, preset laravel
//	drupal   PHP_CodeSniffer, standard Drupal (drupal/coder)
//
// The reference rules custos deliberately does not satisfy are listed per
// preset, with a reason, in internal/format/testdata/divergences.json; any
// other finding after formatting remains, and makes the command exit 1, so
// `make formatcheck` works as a gate. PHP-CS-Fixer and Pint name every
// fixer that touched a file, including the fixers that only tidy up after
// a divergent one (blank lines left by a removed import), and Pint names
// some whose net change is nil; so for their presets the remaining findings
// come from a second run with the divergent fixers turned off: its findings
// for PHP-CS-Fixer, and for Pint the files its fix changes outside
// docblocks.
//
// The checkers are the benchmark's pinned packages (`go run ./tools/benchcmp
// install -tool phpcs,php-cs-fixer,pint`), except drupal/coder, which
// formatcheck installs itself under -tools. The corpora are never modified.
package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
)

func main() { os.Exit(run(os.Args[1:], os.Stdout, os.Stderr, execRunner)) }

const usage = "usage: formatcheck [-corpora list] [-presets list] [-json] [flags]"

// options are the command-line settings of one run.
type options struct {
	corpora, presets             []string
	config, bench, tools, custos string
	divergences, tmp             string
	examples                     int
	json                         bool
}

func run(args []string, stdout, stderr io.Writer, exe runner) int {
	fs := flag.NewFlagSet("formatcheck", flag.ContinueOnError)
	fs.SetOutput(stderr)
	corpora := fs.String("corpora", "", "comma-separated corpus names (default: all of -config)")
	presets := fs.String("presets", "", "comma-separated presets (default: "+strings.Join(presetNames(), ",")+")")
	var o options
	fs.StringVar(&o.config, "config", "bench/competitors.json", "benchmark configuration listing the corpora")
	fs.StringVar(&o.bench, "bench", ".cache/bench", "directory of the installed benchmark tools and corpora")
	fs.StringVar(&o.tools, "tools", ".cache/formatcheck/tools", "directory of the tools formatcheck installs")
	fs.StringVar(&o.custos, "custos", "bin/custos", "custos binary")
	fs.StringVar(&o.divergences, "divergences", "internal/format/testdata/divergences.json", "reference rules custos deliberately does not satisfy")
	fs.StringVar(&o.tmp, "tmp", "", "directory for the copies (default: the system temporary directory)")
	fs.IntVar(&o.examples, "examples", 3, "examples printed per remaining rule")
	fs.BoolVar(&o.json, "json", false, "print the results as JSON")
	if err := fs.Parse(args); err != nil || fs.NArg() != 0 {
		fmt.Fprintln(stderr, usage)
		return 2
	}
	o.corpora, o.presets = list(*corpora), list(*presets)
	// Tools run in the copies' directories.
	for _, p := range []*string{&o.bench, &o.tools, &o.custos} {
		abs, _ := filepath.Abs(*p) // only fails without a working directory
		*p = abs
	}
	outcomes, err := check(o, exe, stderr)
	if err != nil {
		fmt.Fprintf(stderr, "formatcheck: %v\n", err)
		return 2
	}
	if o.json {
		enc := json.NewEncoder(stdout)
		enc.SetIndent("", "  ")
		enc.SetEscapeHTML(false)
		_ = enc.Encode(outcomes) // only our own types
	} else {
		printText(stdout, outcomes, o.examples)
	}
	for _, oc := range outcomes {
		if oc.Remaining > 0 {
			return 1
		}
	}
	return 0
}

// list splits a comma-separated flag value, dropping empty items.
func list(s string) []string {
	var out []string
	for item := range strings.SplitSeq(s, ",") {
		if item = strings.TrimSpace(item); item != "" {
			out = append(out, item)
		}
	}
	return out
}

// corpus is a benchmark corpus: its name and the source paths formatted.
type corpus struct {
	Name  string   `json:"name"`
	Paths []string `json:"paths"`
	// PHP marks a check corpus for an old PHP version, only formatted
	// when named.
	PHP string `json:"php"`
}

// selectCorpora reads the corpora of the benchmark configuration and keeps
// the named ones (all benchmark corpora when names is empty), in the
// configuration's order.
func selectCorpora(config string, names []string) ([]corpus, error) {
	data, err := os.ReadFile(config)
	if err != nil {
		return nil, err
	}
	var c struct {
		Corpora []corpus `json:"corpora"`
	}
	if err := json.Unmarshal(data, &c); err != nil {
		return nil, fmt.Errorf("%s: %w", config, err)
	}
	var out []corpus
	for _, cp := range c.Corpora {
		if (len(names) == 0 && cp.PHP == "") || slices.Contains(names, cp.Name) {
			out = append(out, cp)
		}
	}
	for _, n := range names {
		if !slices.ContainsFunc(out, func(cp corpus) bool { return cp.Name == n }) {
			return nil, fmt.Errorf("unknown corpus %q in %s", n, config)
		}
	}
	return out, nil
}

// check runs every selected preset on every selected corpus.
func check(o options, exe runner, log io.Writer) ([]outcome, error) {
	corpora, err := selectCorpora(o.config, o.corpora)
	if err != nil {
		return nil, err
	}
	presets := o.presets
	if len(presets) == 0 {
		presets = presetNames()
	}
	for _, p := range presets {
		if _, ok := checkers[p]; !ok {
			return nil, fmt.Errorf("unknown preset %q (want %s)", p, strings.Join(presetNames(), ", "))
		}
	}
	divergences, err := loadDivergences(o.divergences)
	if err != nil {
		return nil, err
	}
	if _, err := os.Stat(o.custos); err != nil {
		return nil, fmt.Errorf("custos binary: %w (run make build)", err)
	}
	s := &session{run: exe, bench: o.bench, tools: o.tools, log: log, cpus: runtime.NumCPU()}
	for _, p := range presets {
		if err := checkers[p].ready(s); err != nil {
			return nil, err
		}
	}
	tmp, err := os.MkdirTemp(o.tmp, "formatcheck-")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(tmp)
	// phpcs reports real paths: resolve the temporary directory's symlinks
	// (/tmp and /var are links on macOS) so that they can be made relative.
	tmp, _ = filepath.EvalSymlinks(tmp) // it was just created
	var outcomes []outcome
	for _, c := range corpora {
		src := filepath.Join(o.bench, "corpora", c.Name)
		if _, err := os.Stat(src); err != nil {
			return nil, fmt.Errorf("corpus %s is not installed (go run ./tools/benchcmp install -tool custos -corpus %s): %w", c.Name, c.Name, err)
		}
		orig := filepath.Join(tmp, c.Name, "orig")
		if err := copyCorpus(src, orig, c.Paths); err != nil {
			return nil, fmt.Errorf("%s: %w", c.Name, err)
		}
		for _, p := range presets {
			fmt.Fprintf(log, "formatcheck: %s %s\n", c.Name, p)
			oc, err := s.checkPreset(o.custos, c, p, orig, filepath.Join(tmp, c.Name, p), divergences[p])
			if err != nil {
				return nil, fmt.Errorf("%s %s: %w", c.Name, p, err)
			}
			outcomes = append(outcomes, oc)
		}
	}
	return outcomes, nil
}

// checkPreset formats a copy of orig into dir with preset p and checks the
// copies before and after.
func (s *session) checkPreset(custos string, c corpus, p, orig, dir string, div []divergence) (outcome, error) {
	ck := checkers[p]
	if err := copyCorpus(orig, dir, c.Paths); err != nil {
		return outcome{}, err
	}
	res, err := s.run(command{dir: dir, name: custos, args: append([]string{"format", "--no-cache", "--preset", p}, c.Paths...)})
	if err != nil {
		return outcome{}, err
	}
	refused := refusals(res.stderr)
	if res.code != 0 && len(refused) == 0 {
		return outcome{}, toolError(custos, res, errors.New("formatting failed"))
	}
	before, err := ck.check(s, orig, c.Paths)
	if err != nil {
		return outcome{}, err
	}
	after, err := ck.check(s, dir, c.Paths)
	if err != nil {
		return outcome{}, err
	}
	var gated []finding
	if ck.gate != nil {
		if gated, err = ck.gate(s, dir, c.Paths, div); err != nil {
			return outcome{}, err
		}
	}
	oc := summarize(c.Name, p, ck, before, after, gated, div)
	oc.Findings = append(refused, oc.Findings...) // files custos left as written remain to fix
	oc.Remaining = len(oc.Findings)
	return oc, nil
}

// refusals reads the files custos format refused to change from its
// standard error, `custos: <file>: <reason>`, as findings of rule custos.
func refusals(stderr []byte) []finding {
	var fs []finding
	for line := range strings.Lines(string(stderr)) {
		rest, ok := strings.CutPrefix(strings.TrimRight(line, "\r\n"), "custos: ")
		file, reason, found := strings.Cut(rest, ": ")
		if ok && found && strings.HasSuffix(file, ".php") {
			fs = append(fs, finding{Rule: "custos", File: file, Message: reason})
		}
	}
	return fs
}

// copyCorpus copies the given paths of src into dst. The corpus's
// composer.json is left out: custos then targets its default PHP version,
// like the checkers, whose rules follow the PHP version they run on.
func copyCorpus(src, dst string, paths []string) error {
	for _, p := range paths {
		if err := os.CopyFS(filepath.Join(dst, p), os.DirFS(filepath.Join(src, p))); err != nil {
			return err
		}
	}
	return nil
}
