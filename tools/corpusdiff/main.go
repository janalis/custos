// Command corpusdiff records the findings custos reports on the pinned
// benchmark corpora (every rule, on each corpus's src) and compares a later
// run with that record, so a change that adds or removes findings on real
// code is seen before it lands. `make corpus-baseline` records;
// `make corpus-diff` compares and fails on any difference, writing the
// added and removed findings of each corpus next to its record.
//
// A record names the corpus version it was taken on: comparing with a
// corpus installed at another version is an error, not a difference.
package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
)

func main() { os.Exit(run(os.Args[1:], os.Stdout, os.Stderr)) }

func run(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("corpusdiff", flag.ContinueOnError)
	fs.SetOutput(stderr)
	record := fs.Bool("record", false, "record the findings instead of comparing them")
	corpora := fs.String("corpora", "guzzle,composer,laravel", "comma-separated corpus names")
	bin := fs.String("bin", "bin/custos", "custos binary")
	dir := fs.String("dir", ".cache/bench/corpora", "directory of the installed corpora")
	out := fs.String("out", ".cache/corpus-diff", "directory of the records")
	if err := fs.Parse(args); err != nil || fs.NArg() != 0 {
		fmt.Fprintln(stderr, "usage: corpusdiff [-record] [-corpora list] [-bin path] [-dir dir] [-out dir]")
		return 2
	}
	// The binary runs from each corpus directory.
	binary, _ := filepath.Abs(*bin)
	status := 0
	for name := range strings.SplitSeq(*corpora, ",") {
		c := corpus{name: name, dir: filepath.Join(*dir, name), out: *out, bin: binary}
		var differs bool
		var err error
		if *record {
			err = c.record(stdout)
		} else {
			differs, err = c.compare(stdout)
		}
		switch {
		case err != nil:
			fmt.Fprintf(stderr, "corpusdiff: %s: %v\n", name, err)
			return 2
		case differs:
			status = 1
		}
	}
	return status
}

// corpus is one installed corpus and where its record lives.
type corpus struct {
	name, dir, out, bin string
}

// recordPath is the file holding the corpus's recorded findings.
func (c corpus) recordPath() string { return filepath.Join(c.out, c.name+".txt") }

// version is the installed version of the corpus, as benchcmp wrote it.
func (c corpus) version() (string, error) {
	b, err := os.ReadFile(c.dir + ".version")
	if err != nil {
		return "", fmt.Errorf("not installed (go run ./tools/benchcmp install -tool custos -corpus %s): %w", c.name, err)
	}
	return strings.TrimSpace(string(b)), nil
}

// findings runs custos on the corpus and returns its findings, sorted.
func (c corpus) findings() ([]string, error) {
	cmd := exec.Command(c.bin, "analyse", "-all", "-no-cache", "-fail-on", "never", "src")
	cmd.Dir = c.dir
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("custos: %w: %s", err, strings.TrimSpace(stderr.String()))
	}
	lines := strings.Split(strings.TrimRight(stdout.String(), "\n"), "\n")
	if lines[0] == "" {
		lines = nil
	}
	slices.Sort(lines)
	return lines, nil
}

func (c corpus) record(stdout io.Writer) error {
	version, err := c.version()
	if err != nil {
		return err
	}
	lines, err := c.findings()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(c.out, 0o755); err != nil {
		return err
	}
	body := header(c.name, version) + strings.Join(append(lines, ""), "\n")
	if err := os.WriteFile(c.recordPath(), []byte(body), 0o644); err != nil {
		return err
	}
	fmt.Fprintf(stdout, "%s %s: %d findings recorded\n", c.name, version, len(lines))
	return nil
}

// header is the first line of a record.
func header(name, version string) string { return "# " + name + " " + version + "\n" }

// compare runs custos and reports whether its findings differ from the
// record, writing the difference to <out>/<name>.diff.
func (c corpus) compare(stdout io.Writer) (bool, error) {
	version, err := c.version()
	if err != nil {
		return false, err
	}
	b, err := os.ReadFile(c.recordPath())
	if err != nil {
		return false, fmt.Errorf("no record (make corpus-baseline): %w", err)
	}
	head, body, _ := strings.Cut(string(b), "\n")
	if head+"\n" != header(c.name, version) {
		return false, errors.New("recorded on another version (" + strings.TrimPrefix(head, "# ") + "), installed " + version + ": record again")
	}
	var was []string
	if body != "" {
		was = strings.Split(strings.TrimRight(body, "\n"), "\n")
	}
	now, err := c.findings()
	if err != nil {
		return false, err
	}
	removed, added := difference(was, now), difference(now, was)
	diff := c.recordPath()[:len(c.recordPath())-len(".txt")] + ".diff"
	if len(removed)+len(added) == 0 {
		_ = os.Remove(diff)
		fmt.Fprintf(stdout, "%s %s: %d findings, unchanged\n", c.name, version, len(now))
		return false, nil
	}
	var text strings.Builder
	for _, line := range removed {
		text.WriteString("- " + line + "\n")
	}
	for _, line := range added {
		text.WriteString("+ " + line + "\n")
	}
	if err := os.WriteFile(diff, []byte(text.String()), 0o644); err != nil {
		return false, err
	}
	fmt.Fprintf(stdout, "%s %s: %d findings removed, %d added (%s)\n", c.name, version, len(removed), len(added), diff)
	for _, rule := range byRule(removed, added) {
		fmt.Fprintln(stdout, "  "+rule)
	}
	return true, nil
}

// difference returns the lines of a, sorted, that b lacks, counting
// repeated lines.
func difference(a, b []string) []string {
	left := map[string]int{}
	for _, line := range b {
		left[line]++
	}
	var out []string
	for _, line := range a {
		if left[line] > 0 {
			left[line]--
			continue
		}
		out = append(out, line)
	}
	return out
}

// ruleOf matches the rule ID at the end of a text-format finding.
var ruleOf = regexp.MustCompile(`\[([A-Za-z0-9_]+)\]( \(fixable\))?$`)

// byRule summarizes removed and added findings by rule, sorted.
func byRule(removed, added []string) []string {
	type count struct{ removed, added int }
	counts := map[string]*count{}
	tally := func(lines []string, add bool) {
		for _, line := range lines {
			rule := "(unknown rule)"
			if m := ruleOf.FindStringSubmatch(line); m != nil {
				rule = m[1]
			}
			if counts[rule] == nil {
				counts[rule] = &count{}
			}
			if add {
				counts[rule].added++
			} else {
				counts[rule].removed++
			}
		}
	}
	tally(removed, false)
	tally(added, true)
	var out []string
	for rule, n := range counts {
		out = append(out, fmt.Sprintf("%s: -%d +%d", rule, n.removed, n.added))
	}
	slices.Sort(out)
	return out
}
