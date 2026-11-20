package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/janalis/custos/internal/diagnostic"
	"github.com/janalis/custos/internal/format"
	"github.com/janalis/custos/internal/output/diff"
	"github.com/janalis/custos/internal/output/report"
	"github.com/janalis/custos/internal/php/syntax"
	phpversion "github.com/janalis/custos/internal/php/version"
	"github.com/janalis/custos/internal/project"
	"github.com/janalis/custos/internal/project/config"
)

// formatOptions are the flags of custos format.
type formatOptions struct {
	check, dry, diff, stdin, showConfig, allowRisky, noCache bool
	stdinPath, report, preset, php, configDir, exclude       string
	set                                                      setFlag
}

// setFlag collects repeated --set key=value flags.
type setFlag map[string]any

func (s setFlag) String() string { return "" }

// Set reads key=value: the value as JSON when it parses (numbers, booleans,
// arrays), else as text, where a list setting takes comma-separated items.
func (s setFlag) Set(arg string) error {
	key, value, ok := strings.Cut(arg, "=")
	if !ok || key == "" {
		return fmt.Errorf("want key=value, got %q", arg)
	}
	var v any
	if json.Unmarshal([]byte(value), &v) != nil {
		v = value
	}
	s[strings.TrimSpace(key)] = v
	return nil
}

func cmdFormat(args []string, stdin io.Reader, stdout, stderr io.Writer) (int, error) {
	o := formatOptions{set: setFlag{}}
	fs := newFlagSet("format", stderr)
	fs.BoolVar(&o.check, "check", false, "write nothing; list what would change and exit 1 if a file would change")
	fs.BoolVar(&o.dry, "dry-run", false, "write nothing; list what would change (with --diff, print it) and exit 1 if a file would change")
	fs.BoolVar(&o.diff, "diff", false, "print a unified diff of the changes (files are still written unless --check or --dry-run)")
	fs.BoolVar(&o.stdin, "stdin", false, `format standard input to standard output (as the path "-")`)
	fs.StringVar(&o.stdinPath, "stdin-path", "", "file that standard input stands for: configuration, .editorconfig, exclusions and messages")
	fs.StringVar(&o.report, "report", "text", "listing format of --check and --dry-run: "+strings.Join(report.Formats, ", "))
	fs.BoolVar(&o.showConfig, "show-config", false, "print each setting of the first path (default .) with its value and source, then exit")
	fs.StringVar(&o.preset, "preset", "", "preset: "+config.AutoPreset+", "+strings.Join(format.Presets(), ", ")+` (default: custos.json, else auto)`)
	fs.Var(o.set, "set", "set a setting, key=value (repeatable; the value is JSON when it parses, else text; lists are comma-separated)")
	fs.BoolVar(&o.allowRisky, "allow-risky", false, "apply the preset's risky settings")
	fs.StringVar(&o.php, "php", "", "target PHP version (default: custos.json, composer.json, else "+phpversion.Default.String()+")")
	fs.StringVar(&o.configDir, "config", "", "directory to look for custos.json (default: first path)")
	fs.StringVar(&o.exclude, "exclude", "", "comma-separated extra directories to exclude")
	fs.BoolVar(&o.noCache, "no-cache", false, "do not skip or remember the files known to be formatted")
	if err := fs.Parse(args); err != nil {
		return 2, err
	}
	paths := fs.Args()
	if len(paths) == 1 && paths[0] == "-" {
		o.stdin, paths = true, nil
	}
	switch {
	case !slices.Contains(report.Formats, o.report):
		return 2, fmt.Errorf("unknown report format %q (want one of %s)", o.report, strings.Join(report.Formats, ", "))
	case o.diff && o.report != "text":
		return 2, errors.New("--diff prints text: it cannot be combined with --report " + o.report)
	case o.stdin && len(paths) > 0:
		return 2, errors.New(`standard input ("-" or --stdin) cannot be formatted with other paths`)
	}
	where := paths
	if o.stdin && o.stdinPath != "" {
		where = []string{o.stdinPath}
	}
	cfg, err := loadConfig(o.configDir, where, o.php)
	if err != nil {
		return 2, err
	}
	resolver, err := cfg.FormatResolver(config.FormatFlags{Preset: o.preset, Set: o.set, AllowRisky: o.allowRisky})
	if err != nil {
		return 2, err
	}
	switch {
	case o.showConfig:
		target := "."
		if len(where) > 0 {
			target = where[0]
		}
		return showFormatConfig(stdout, resolver, target)
	case o.stdin:
		return formatStdin(o, cfg, resolver, stdin, stdout, stderr)
	}
	tuneGC()
	return formatFiles(o, cfg, resolver, paths, stdout, stderr)
}

// showFormatConfig prints the settings of target, or of a PHP file in it
// when it is a directory, with the source of each value.
func showFormatConfig(w io.Writer, r *config.FormatResolver, target string) (int, error) {
	if st, err := os.Stat(target); err == nil && st.IsDir() {
		target = filepath.Join(target, "example.php")
	}
	res, err := r.Settings(target)
	if err != nil {
		return 2, err
	}
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	name, source := r.Preset()
	fmt.Fprintf(tw, "preset\t%s\t%s\n", name, source)
	for _, d := range format.Registry() {
		v, _ := json.Marshal(d.Get(&res.Settings))
		fmt.Fprintf(tw, "%s\t%s\t%s\n", d.ID, v, res.Source[d.ID])
	}
	return 0, tw.Flush()
}

// formatTally counts the outcomes of a run.
type formatTally struct {
	changed, unchanged, skipped, failed int
	reported                            int // files with nothing to rewrite but lines too long
	items                               []diagnostic.Item
}

// summary is the last line a run prints; benchmark tooling parses the
// --check form.
func (t *formatTally) summary(w io.Writer, writing bool) {
	verb := "would be reformatted"
	if writing {
		verb = "reformatted"
	}
	fmt.Fprintf(w, "%d file(s) %s, %d unchanged, %d skipped", t.changed, verb, t.unchanged, t.skipped)
	if t.reported > 0 {
		fmt.Fprintf(w, ", %d with lines too long", t.reported)
	}
	fmt.Fprintln(w)
}

// skip records a file formatting left alone: one ignored by a comment is
// intended; the engine refusing its own output fails the run; a file with
// syntax errors only warns.
func (t *formatTally) skip(stderr io.Writer, name, why string) {
	t.skipped++
	if strings.HasPrefix(why, "ignored") {
		return
	}
	if strings.Contains(why, "please report") {
		t.failed++
	}
	fmt.Fprintf(stderr, "custos: %s: %s, not formatted\n", name, why)
	t.items = append(t.items, diagnostic.Item{Path: name, Line: 1, Column: 1, EndLine: 1, EndColumn: 1, Rule: "format.skipped", Severity: string(diagnostic.SeverityError), Message: "Not formatted: " + why})
}

// add lists the changes of res, made on src: every one, or only those
// formatting cannot make when reportsOnly.
func (t *formatTally) add(name string, src []byte, res format.Result, reportsOnly bool) {
	changes := slices.Clone(res.Changes)
	sort.SliceStable(changes, func(i, j int) bool { return changes[i].Start < changes[j].Start })
	li := syntax.NewLineIndex(src)
	for _, c := range changes {
		if reportsOnly && !c.Report {
			continue
		}
		l1, c1 := li.RuneColumn(c.Start)
		l2, c2 := li.RuneColumn(c.End)
		t.items = append(t.items, diagnostic.Item{Path: name, Line: l1 + 1, Column: c1 + 1, EndLine: l2 + 1, EndColumn: c2 + 1, Rule: "format." + c.Rule, Severity: string(diagnostic.SeverityInfo), Message: c.Message()})
	}
}

// list prints the listing in the report format.
func (t *formatTally) list(w io.Writer, kind string, files int) error {
	if kind != "text" {
		return report.Write(w, kind, t.items, files)
	}
	last := ""
	for _, it := range t.items {
		if it.Path != last {
			last = it.Path
			if _, err := fmt.Fprintln(w, it.Path); err != nil {
				return err
			}
		}
		if _, err := fmt.Fprintf(w, "  %d:%d  %s  %s\n", it.Line, it.Column, it.Rule, it.Message); err != nil {
			return err
		}
	}
	return nil
}

// exit returns the exit status of a run.
func (t *formatTally) exit(checking bool) (int, error) {
	switch {
	case t.failed > 0:
		return 2, fmt.Errorf("%d file(s) could not be formatted", t.failed)
	case checking && t.hasChanges():
		return 1, nil
	}
	return 0, nil
}

// hasChanges reports whether the listing holds a change (a skipped file
// alone does not fail a check).
func (t *formatTally) hasChanges() bool {
	for _, it := range t.items {
		if it.Rule != "format.skipped" {
			return true
		}
	}
	return false
}

func formatFiles(o formatOptions, cfg *config.Config, r *config.FormatResolver, paths []string, stdout, stderr io.Writer) (int, error) {
	if len(paths) == 0 {
		for _, p := range cfg.Paths {
			paths = append(paths, filepath.Join(cfg.Root, p))
		}
	}
	exclude := cfg.Exclude
	if o.exclude != "" {
		exclude = append(slices.Clip(exclude), strings.Split(o.exclude, ",")...)
	}
	found, err := project.Discover(paths, exclude)
	if err != nil {
		return 2, err
	}
	files := found[:0]
	for _, f := range found {
		if !cfg.FormatExcluded(f) {
			files = append(files, f)
		}
	}
	var cache *project.FormatCache
	if !o.noCache {
		cache = project.OpenFormatCache(project.CacheDir(), cfg.Root)
	}
	outs := project.FormatFiles(files, project.FormatOptions{Settings: r.Settings, PHP: cfg.PHP, Cache: cache})
	checking := o.check || o.dry
	t := &formatTally{}
	for i, path := range files {
		out := outs[i]
		name := filepath.ToSlash(path)
		switch {
		case errors.Is(out.Err, project.ErrSkipped):
			fmt.Fprintf(stderr, "custos: %v, not formatted\n", out.Err)
			t.skipped++
			continue
		case out.Err != nil:
			fmt.Fprintln(stderr, "custos:", out.Err)
			t.skipped++
			t.failed++
			continue
		case out.Result.Skipped != "":
			t.skip(stderr, name, out.Result.Skipped)
			continue
		}
		res := out.Result
		rewritten := string(res.Output) != string(out.Source)
		if checking {
			t.add(name, out.Source, res, false)
			switch {
			case rewritten:
				t.changed++
			case res.Changed():
				t.reported++
			default:
				t.unchanged++
			}
		} else {
			t.add(name, out.Source, res, true)
			if rewritten {
				t.changed++
			} else {
				t.unchanged++
			}
		}
		if rewritten && o.diff {
			fmt.Fprint(stdout, diff.Unified(name, string(out.Source), string(res.Output), 3))
		}
		if rewritten && !checking {
			if err := os.WriteFile(path, res.Output, out.Perm); err != nil {
				fmt.Fprintln(stderr, "custos:", err)
				t.changed--
				t.skipped++
				t.failed++
			}
		}
	}
	if err := cache.Save(); err != nil {
		fmt.Fprintln(stderr, "custos: format cache not saved:", err)
	}
	if !o.diff {
		if err := t.list(stdout, o.report, len(files)); err != nil {
			return 2, err
		}
	}
	t.summary(stderr, !checking)
	return t.exit(checking)
}

// formatStdin formats standard input to standard output, or lists its
// changes with --check.
func formatStdin(o formatOptions, cfg *config.Config, r *config.FormatResolver, stdin io.Reader, stdout, stderr io.Writer) (int, error) {
	src, err := io.ReadAll(io.LimitReader(stdin, syntax.MaxFileSize+1))
	if err != nil {
		return 2, err
	}
	if len(src) > syntax.MaxFileSize {
		// Writing back a cut text would lose the rest of the input.
		return 2, fmt.Errorf("standard input exceeds %d bytes", syntax.MaxFileSize)
	}
	name := o.stdinPath
	if name == "" {
		name = "stdin.php"
	}
	t := &formatTally{}
	res := format.Result{Output: src}
	if !cfg.FormatExcluded(name) {
		s, err := r.Settings(name)
		if err != nil {
			return 2, err
		}
		res = format.Source(name, src, format.Options{Settings: &s.Settings, PHP: cfg.PHP})
	}
	display := filepath.ToSlash(name)
	rewritten := string(res.Output) != string(src)
	switch {
	case res.Skipped != "":
		t.skip(stderr, display, res.Skipped)
	case res.Changed():
		if rewritten {
			t.changed++
		} else {
			t.reported++
		}
		t.add(display, src, res, false)
	default:
		t.unchanged++
	}
	switch {
	case o.check:
		if err := t.list(stdout, o.report, 1); err != nil {
			return 2, err
		}
		t.summary(stderr, false)
	case o.diff:
		if rewritten {
			fmt.Fprint(stdout, diff.Unified(display, string(src), string(res.Output), 3))
		}
	default:
		if _, err := stdout.Write(res.Output); err != nil {
			return 2, err
		}
	}
	return t.exit(o.check || o.dry)
}
