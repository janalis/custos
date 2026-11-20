package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"maps"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
)

// coderVersion is the drupal/coder release formatcheck installs: the last
// 8.3 release, whose Drupal standard runs on PHP_CodeSniffer 3.
const coderVersion = "8.3.31"

// command is one process to run.
type command struct {
	dir  string
	env  []string // added to the environment
	name string
	args []string
}

// result is what a finished process left: its outputs and exit status.
type result struct {
	stdout, stderr []byte
	code           int
}

// runner runs a command to completion. An error means it could not run; a
// non-zero exit status is not an error.
type runner func(command) (result, error)

func execRunner(c command) (result, error) {
	cmd := exec.Command(c.name, c.args...)
	cmd.Dir = c.dir
	cmd.Env = append(os.Environ(), c.env...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	err := cmd.Run()
	r := result{stdout: stdout.Bytes(), stderr: stderr.Bytes()}
	if ee, ok := errors.AsType[*exec.ExitError](err); ok {
		r.code = ee.ExitCode()
		err = nil
	}
	return r, err
}

// toolError describes a tool that failed or printed what could not be read.
func toolError(name string, r result, err error) error {
	msg := strings.TrimSpace(string(r.stderr))
	if len(msg) > 2000 {
		msg = msg[:2000] + "…"
	}
	return fmt.Errorf("%s (exit status %d): %w: %s", filepath.Base(name), r.code, err, msg)
}

// session holds what the checkers need.
type session struct {
	run   runner
	bench string // installed benchmark tools and corpora
	tools string // tools formatcheck installs
	log   io.Writer
	cpus  int
	// fixerRules caches the rule names of the installed PHP-CS-Fixer.
	fixerRules []string
}

func (s *session) benchBin(pkg, bin string) string {
	return filepath.Join(s.bench, "tools", pkg, "vendor", "bin", bin)
}

func (s *session) phpcsBin() string  { return s.benchBin("squizlabs-php_codesniffer", "phpcs") }
func (s *session) fixerBin() string  { return s.benchBin("friendsofphp-php-cs-fixer", "php-cs-fixer") }
func (s *session) pintBin() string   { return s.benchBin("laravel-pint", "pint") }
func (s *session) coderDir() string  { return filepath.Join(s.tools, "drupal-coder") }
func (s *session) drupalBin() string { return filepath.Join(s.coderDir(), "vendor", "bin", "phpcs") }

// finding is one report of a checker: the rule (sniff code or fixer name),
// the file relative to the copy, its line (0 when the checker gives none)
// and its message.
type finding struct {
	Rule    string `json:"rule"`
	File    string `json:"file"`
	Line    int    `json:"line"`
	Message string `json:"message"`
}

// checker is the reference checker of a preset.
type checker struct {
	label string
	// bin is the checker's executable; install, when set, installs it.
	bin     func(*session) string
	install func(*session) error
	// check reports the findings on the copy at root.
	check func(s *session, root string, paths []string) ([]finding, error)
	// gate, when set, finds what remains on the formatted copy instead of
	// subtracting the divergences from check's findings.
	gate func(s *session, root string, paths []string, div []divergence) ([]finding, error)
	// note explains the remaining column of a gated checker.
	note string
}

const fixerNote = "php-cs-fixer: remaining comes from a second run with the divergent fixers off; the rest of after - divergent\n" +
	"  are fixers reported only for tidying up after a divergent one."

var checkers = map[string]checker{
	"psr12": {
		label: "phpcs --standard=PSR12",
		bin:   (*session).phpcsBin,
		check: func(s *session, root string, paths []string) ([]finding, error) {
			return s.phpcs(s.phpcsBin(), nil, "PSR12", root, paths)
		},
	},
	"per": {
		label: "php-cs-fixer --rules=@PER-CS",
		bin:   (*session).fixerBin,
		check: func(s *session, root string, paths []string) ([]finding, error) {
			return s.fixer("@PER-CS", root, paths)
		},
		gate: fixerGate("@PER-CS"),
		note: fixerNote,
	},
	"symfony": {
		label: "php-cs-fixer --rules=@Symfony",
		bin:   (*session).fixerBin,
		check: func(s *session, root string, paths []string) ([]finding, error) {
			return s.fixer("@Symfony", root, paths)
		},
		gate: fixerGate("@Symfony"),
		note: fixerNote,
	},
	"laravel": {
		label: "pint --preset laravel",
		bin:   (*session).pintBin,
		check: (*session).pintTest,
		gate:  (*session).pintGate,
		note: "pint: remaining counts the files whose code (docblocks aside) Pint's fix changes with the divergent fixers off;\n" +
			"  the other fixers Pint reports there change nothing in the end.",
	},
	"drupal": {
		label:   "phpcs --standard=Drupal",
		bin:     (*session).drupalBin,
		install: (*session).installCoder,
		check: func(s *session, root string, paths []string) ([]finding, error) {
			// Composer's installer plugin, which would register the
			// standards, does not run: name them on the command line.
			v := filepath.Join(s.coderDir(), "vendor")
			standards := strings.Join([]string{
				filepath.Join(v, "drupal", "coder", "coder_sniffer"),
				filepath.Join(v, "slevomat", "coding-standard"),
				filepath.Join(v, "sirbrillig", "phpcs-variable-analysis"),
			}, ",")
			return s.phpcs(s.drupalBin(), []string{"--runtime-set", "installed_paths", standards}, "Drupal", root, paths)
		},
	},
}

// presetNames lists the presets formatcheck knows, in style-guide order.
func presetNames() []string { return []string{"psr12", "per", "symfony", "laravel", "drupal"} }

// ready installs the checker if formatcheck owns it, then requires it.
func (c checker) ready(s *session) error {
	if c.install != nil {
		if err := c.install(s); err != nil {
			return err
		}
	}
	if _, err := os.Stat(c.bin(s)); err != nil {
		return fmt.Errorf("%s: %w (go run ./tools/benchcmp install -tool phpcs,php-cs-fixer,pint)", c.label, err)
	}
	return nil
}

// installCoder installs drupal/coder at coderVersion under -tools, unless
// its stamp says it is there.
func (s *session) installCoder() error {
	dir := s.coderDir()
	stamp := dir + ".version"
	if got, err := os.ReadFile(stamp); err == nil && string(got) == coderVersion {
		return nil
	}
	fmt.Fprintf(s.log, "installing drupal/coder %s\n", coderVersion)
	if err := errors.Join(os.RemoveAll(dir), os.MkdirAll(dir, 0o755)); err != nil {
		return err
	}
	// No package script or plugin runs.
	r, err := s.run(command{dir: dir, name: "composer", args: []string{
		"require", "--no-interaction", "--no-progress", "--no-scripts", "--no-plugins", "drupal/coder:" + coderVersion,
	}})
	if err != nil {
		return err
	}
	if r.code != 0 {
		return toolError("composer", r, errors.New("installing drupal/coder failed"))
	}
	return os.WriteFile(stamp, []byte(coderVersion), 0o644)
}

// phpcs runs PHP_CodeSniffer with a standard. Its exit status only says
// whether it found something, so the output decides.
func (s *session) phpcs(bin string, extra []string, standard, root string, paths []string) ([]finding, error) {
	args := append([]string{"-q", "-d", "memory_limit=-1", "--standard=" + standard, "--report=json", "--parallel=" + strconv.Itoa(s.cpus)}, extra...)
	r, err := s.run(command{dir: root, name: bin, args: append(args, paths...)})
	if err != nil {
		return nil, err
	}
	fs, err := parsePHPCS(r.stdout, root)
	if err != nil {
		return nil, toolError(bin, r, err)
	}
	return fs, nil
}

// parsePHPCS reads the JSON report of PHP_CodeSniffer, whose files are
// absolute paths: they are made relative to root.
func parsePHPCS(out []byte, root string) ([]finding, error) {
	var report struct {
		Files map[string]struct {
			Messages []struct {
				Message string `json:"message"`
				Source  string `json:"source"`
				Line    int    `json:"line"`
			} `json:"messages"`
		} `json:"files"`
	}
	if err := json.Unmarshal(out, &report); err != nil {
		return nil, err
	}
	if report.Files == nil {
		return nil, errors.New("no files section")
	}
	var fs []finding
	for path, f := range report.Files {
		rel := strings.TrimPrefix(path, root+string(filepath.Separator))
		for _, m := range f.Messages {
			fs = append(fs, finding{Rule: m.Source, File: rel, Line: m.Line, Message: m.Message})
		}
	}
	return fs, nil
}

// fixer runs `php-cs-fixer check` with a rule set. Exit status 8 means it
// would change files; 64 that a file errored inside PHP-CS-Fixer (it does
// on some unformatted sources), which the report still covers, so it only
// warns; any other bit is a failure (4: a file does not parse).
func (s *session) fixer(rules, root string, paths []string) ([]finding, error) {
	bin := s.fixerBin()
	r, err := s.run(command{dir: root, env: []string{"PHP_CS_FIXER_IGNORE_ENV=1"}, name: bin, args: append([]string{
		"check", "-v", "--diff", "--rules=" + rules, "--format=json", "--using-cache=no", "--show-progress=none",
	}, paths...)})
	if err != nil {
		return nil, err
	}
	if r.code&^(8|64) != 0 {
		return nil, toolError(bin, r, errors.New("check failed"))
	}
	if r.code&64 != 0 {
		fmt.Fprintf(s.log, "formatcheck: %s: some files errored inside PHP-CS-Fixer (exit status %d)\n", root, r.code)
	}
	fs, err := parseFixer(r.stdout)
	if err != nil {
		return nil, toolError(bin, r, err)
	}
	return fs, nil
}

// fixerGate checks the formatted copy with the rule set, the divergent
// fixers turned off. PHP-CS-Fixer names every fixer that touched a file, so
// a divergent fixer's change (removing an import, say) also has the fixers
// that tidy up after it reported; with it off, they are not.
func fixerGate(set string) func(*session, string, []string, []divergence) ([]finding, error) {
	return func(s *session, root string, paths []string, div []divergence) ([]finding, error) {
		known, err := s.listFixerRules()
		if err != nil {
			return nil, err
		}
		rules := map[string]bool{set: true}
		for _, d := range ruleWide(div) {
			if !isPattern(d.Rule) {
				rules[d.Rule] = false // an unknown name fails the run
				continue
			}
			for _, name := range known {
				if ok, _ := path.Match(d.Rule, name); ok { // patterns are checked on load
					rules[name] = false
				}
			}
		}
		data, _ := json.Marshal(rules) // plain values
		return s.fixer(string(data), root, paths)
	}
}

// listFixerRules returns the names of every rule of the installed
// PHP-CS-Fixer (`list-rules --format=json`), once.
func (s *session) listFixerRules() ([]string, error) {
	if s.fixerRules != nil {
		return s.fixerRules, nil
	}
	bin := s.fixerBin()
	r, err := s.run(command{dir: s.bench, env: []string{"PHP_CS_FIXER_IGNORE_ENV=1"}, name: bin, args: []string{"list-rules", "--format=json"}})
	if err != nil {
		return nil, err
	}
	var list struct {
		Rules map[string]json.RawMessage `json:"rules"`
	}
	if err := json.Unmarshal(r.stdout, &list); err != nil || len(list.Rules) == 0 {
		return nil, toolError(bin, r, errors.New("no rule list"))
	}
	s.fixerRules = slices.Sorted(maps.Keys(list.Rules))
	return s.fixerRules, nil
}

// parseFixer reads the JSON report of `php-cs-fixer check -v --diff`: each
// file lists its fixers (only with -v) and one diff. A fixer is reported at
// the first change of its file's diff, as the diff does not say which
// fixer made which change.
func parseFixer(out []byte) ([]finding, error) {
	var report struct {
		Files *[]struct {
			Name          string   `json:"name"`
			AppliedFixers []string `json:"appliedFixers"`
			Diff          string   `json:"diff"`
		} `json:"files"`
	}
	if err := json.Unmarshal(out, &report); err != nil {
		return nil, err
	}
	if report.Files == nil {
		return nil, errors.New("no files section")
	}
	var fs []finding
	for _, f := range *report.Files {
		line, old, repl := firstChange(f.Diff)
		msg := fmt.Sprintf("first change: %q → %q", old, repl)
		for _, fixer := range f.AppliedFixers {
			fs = append(fs, finding{Rule: fixer, File: f.Name, Line: line, Message: msg})
		}
	}
	return fs, nil
}

// firstChange returns the line of the first change of a unified diff in the
// old file, with the old and new text of that line ("" for an added or
// removed line), indentation included. The line is 0 without a hunk.
func firstChange(diff string) (int, string, string) {
	line := 0
	inHunk := false
	lines := strings.Split(diff, "\n")
	for i, l := range lines {
		switch {
		case strings.HasPrefix(l, "@@ -"):
			start, _, _ := strings.Cut(strings.Fields(l)[1][1:], ",")
			line, _ = strconv.Atoi(start)
			inHunk = true
		case !inHunk:
		case strings.HasPrefix(l, " "):
			line++
		case strings.HasPrefix(l, "-"):
			repl := ""
			if i+1 < len(lines) && strings.HasPrefix(lines[i+1], "+") {
				repl = lines[i+1][1:]
			}
			return line, l[1:], repl
		case strings.HasPrefix(l, "+"):
			return line, "", l[1:]
		}
	}
	return 0, "", ""
}

// pintFile is a file of Pint's JSON report.
type pintFile struct {
	Path   string   `json:"path"`
	Fixers []string `json:"fixers"`
}

// parsePint reads Pint's JSON report: its result, and the files it would
// change or changed with their fixers.
func parsePint(out []byte) ([]pintFile, error) {
	var report struct {
		Result string     `json:"result"`
		Files  []pintFile `json:"files"`
	}
	if err := json.Unmarshal(out, &report); err != nil {
		return nil, err
	}
	if report.Result == "" {
		return nil, errors.New("no result")
	}
	return report.Files, nil
}

// pint runs Pint in root with args and reads its report.
func (s *session) pint(root string, args, paths []string) ([]pintFile, error) {
	bin := s.pintBin()
	// A cache file of the copy's own, not shared with other runs.
	args = append(args, "--format=json", "--cache-file="+root+".pint-cache")
	r, err := s.run(command{dir: root, name: bin, args: append(args, paths...)})
	if err != nil {
		return nil, err
	}
	files, err := parsePint(r.stdout)
	if err != nil {
		return nil, toolError(bin, r, err)
	}
	return files, nil
}

// pintTest reports each fixer `pint --test` names for a file; Pint gives no
// line.
func (s *session) pintTest(root string, paths []string) ([]finding, error) {
	files, err := s.pint(root, []string{"--test", "--no-config", "--preset", "laravel"}, paths)
	if err != nil {
		return nil, err
	}
	var fs []finding
	for _, f := range files {
		for _, fixer := range f.Fixers {
			fs = append(fs, finding{Rule: fixer, File: f.Path})
		}
	}
	return fs, nil
}

// pintGate fixes a copy of the formatted code at root with Pint, the
// divergent fixers turned off (patterns cannot be: they are left on), and
// reports each file whose code, docblocks aside, Pint changed. Pint names
// some fixers on files whose net change is nil; this ignores them.
func (s *session) pintGate(root string, paths []string, div []divergence) ([]finding, error) {
	fixed := root + "-pint"
	if err := copyCorpus(root, fixed, paths); err != nil {
		return nil, err
	}
	rules := map[string]bool{}
	for _, d := range ruleWide(div) {
		if !isPattern(d.Rule) {
			rules[d.Rule] = false
		}
	}
	config := fixed + ".pint.json"
	data, _ := json.Marshal(map[string]any{"preset": "laravel", "rules": rules}) // plain values
	if err := os.WriteFile(config, data, 0o644); err != nil {
		return nil, err
	}
	files, err := s.pint(fixed, []string{"--config", config}, paths)
	if err != nil {
		return nil, err
	}
	var fs []finding
	for _, f := range files {
		before, err := os.ReadFile(filepath.Join(root, f.Path))
		if err != nil {
			return nil, err
		}
		after, err := os.ReadFile(filepath.Join(fixed, f.Path))
		if err != nil {
			return nil, err
		}
		if line, old, repl, ok := codeChange(before, after); ok {
			fs = append(fs, finding{
				Rule:    strings.Join(slices.Sorted(slices.Values(f.Fixers)), ","),
				File:    f.Path,
				Line:    line,
				Message: fmt.Sprintf("Pint changes %q → %q", old, repl),
			})
		}
	}
	return fs, nil
}

// numbered is a line and its number.
type numbered struct {
	n    int
	text string
}

// codeLines returns the lines of src outside docblocks.
func codeLines(src []byte) []numbered {
	var out []numbered
	doc := false
	for i, l := range strings.Split(string(src), "\n") {
		t := strings.TrimSpace(l)
		if !doc && strings.HasPrefix(t, "/**") {
			doc = true
			t = t[3:] // `/**/` closes on its own line
		}
		if doc {
			doc = !strings.Contains(t, "*/")
			continue
		}
		out = append(out, numbered{i + 1, l})
	}
	return out
}

// codeChange compares two versions of a file outside docblocks: the first
// line of before that differs, with its old and new text ("" for an added
// or removed line), trimmed.
func codeChange(before, after []byte) (int, string, string, bool) {
	a, b := codeLines(before), codeLines(after)
	for i := range min(len(a), len(b)) {
		if a[i].text != b[i].text {
			return a[i].n, strings.TrimSpace(a[i].text), strings.TrimSpace(b[i].text), true
		}
	}
	switch {
	case len(a) > len(b):
		return a[len(b)].n, strings.TrimSpace(a[len(b)].text), "", true
	case len(b) > len(a):
		line := 1
		if len(a) > 0 {
			line = a[len(a)-1].n
		}
		return line, "", strings.TrimSpace(b[len(a)].text), true
	}
	return 0, "", "", false
}
