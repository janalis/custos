// Command custos inspects and fixes PHP code.
package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"runtime/pprof"
	"slices"
	"strings"
	"sync"
	"time"

	"custos/internal/analysis"
	"custos/internal/baseline"
	"custos/internal/config"
	"custos/internal/diff"
	"custos/internal/fix"
	"custos/internal/index"
	"custos/internal/lsp"
	"custos/internal/meta"
	"custos/internal/phpver"
	"custos/internal/report"
	"custos/internal/rules"
	"custos/internal/runner"
	"custos/internal/syntax"
)

var version = "dev"

const usage = `custos — PHP inspections and fixes

Usage:
  custos analyse [flags] [paths...]   report problems
  custos fix [flags] [paths...]       apply quick-fixes
  custos lsp                          run the language server (stdio)
  custos rules [--json]               list rules
  custos explain <rule>               describe a rule
  custos version                      print version

Run "custos <command> -h" for command flags.
`

// Seams for tests: the rule registry and the catalogue are embedded, so
// their failure paths can only be exercised by swapping them.
var (
	registry  = rules.All
	catalogue = meta.All
)

func main() { os.Exit(run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr)) }

func run(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		fmt.Fprint(stderr, usage)
		return 2
	}
	var err error
	code := 0
	switch args[0] {
	case "analyse", "analyze":
		code, err = cmdAnalyse(args[1:], stdout, stderr)
	case "fix":
		code, err = cmdFix(args[1:], stdout, stderr)
	case "rules":
		err = cmdRules(args[1:], stdout, stderr)
	case "explain":
		err = cmdExplain(args[1:], stdout)
	case "lsp":
		err = cmdLSP(args[1:], stdin, stdout, stderr)
	case "version", "--version":
		fmt.Fprintln(stdout, "custos", version)
	case "-h", "--help", "help":
		fmt.Fprint(stdout, usage)
	default:
		fmt.Fprintf(stderr, "custos: unknown command %q\n\n%s", args[0], usage)
		return 2
	}
	if errors.Is(err, flag.ErrHelp) {
		return 0
	}
	if err != nil {
		fmt.Fprintln(stderr, "custos:", err)
		if code == 0 {
			code = 2
		}
	}
	return code
}

// common holds flags shared by analyse and fix.
type common struct {
	fs        *flag.FlagSet
	php       string
	style     string
	configDir string
	only      string
	all       bool
	exclude   string
	profile   string
}

func newFlagSet(name string, stderr io.Writer) *flag.FlagSet {
	fs := flag.NewFlagSet(name, flag.ContinueOnError)
	fs.SetOutput(stderr)
	return fs
}

func newCommon(name string, stderr io.Writer) *common {
	c := &common{fs: newFlagSet(name, stderr)}
	c.fs.StringVar(&c.style, "comparison-style", "", "operand order in comparisons: regular or yoda (default: custos.json)")
	c.fs.StringVar(&c.php, "php", "", "target PHP version (default: custos.json, composer.json, else "+phpver.Default.String()+")")
	c.fs.StringVar(&c.configDir, "config", "", "directory to look for custos.json (default: first path)")
	c.fs.StringVar(&c.only, "rule", "", "comma-separated rule IDs to run (default: enabled rules)")
	c.fs.BoolVar(&c.all, "all", false, "enable every rule, including disabled-by-default ones")
	c.fs.StringVar(&c.exclude, "exclude", "", "comma-separated extra directories to exclude")
	c.fs.StringVar(&c.profile, "cpuprofile", "", "write a CPU profile to this file")
	return c
}

type setup struct {
	cfg    *config.Config
	engine *analysis.Engine
	files  []string
	parse  syntax.Options
}

func (c *common) prepare(args []string) (*setup, error) {
	if err := c.fs.Parse(args); err != nil {
		return nil, err
	}
	paths := c.fs.Args()
	dir := c.configDir
	if dir == "" {
		dir = "."
		if len(paths) > 0 {
			dir = paths[0]
			if st, err := os.Stat(dir); err == nil && !st.IsDir() {
				dir = filepath.Dir(dir)
			}
		}
	}
	cfg, err := config.Load(dir)
	if err != nil {
		return nil, err
	}
	if c.php != "" {
		v, err := phpver.Parse(c.php)
		if err != nil {
			return nil, err
		}
		cfg.PHP, cfg.PHPSource = v, "flag"
	}
	switch c.style {
	case "":
	case "regular":
		cfg.ComparisonStyle = analysis.StyleRegular
	case "yoda":
		cfg.ComparisonStyle = analysis.StyleYoda
	default:
		return nil, fmt.Errorf("--comparison-style must be regular or yoda, got %q", c.style)
	}
	acfg := cfg.Analysis()
	acfg.EnableAll = c.all
	if c.only != "" {
		for _, id := range strings.Split(c.only, ",") {
			m, ok := meta.Lookup(strings.TrimSpace(id))
			if !ok {
				return nil, fmt.Errorf("unknown rule %q", id)
			}
			acfg.Only = append(acfg.Only, m.ID)
		}
	}
	e, err := analysis.NewEngine(registry(), acfg)
	if err != nil {
		return nil, err
	}
	if len(paths) == 0 {
		for _, p := range cfg.Paths {
			paths = append(paths, filepath.Join(cfg.Root, p))
		}
	}
	exclude := cfg.Exclude
	if c.exclude != "" {
		exclude = append(exclude, strings.Split(c.exclude, ",")...)
	}
	files, err := runner.DiscoverWith(paths, exclude, e.FilePatterns())
	if err != nil {
		return nil, err
	}
	return &setup{cfg: cfg, engine: e, files: files, parse: syntax.Options{Version: cfg.PHP, ShortOpenTag: cfg.ShortOpenTag}}, nil
}

func (c *common) startProfile() (func(), error) {
	if c.profile == "" {
		return func() {}, nil
	}
	f, err := os.Create(c.profile)
	if err != nil {
		return nil, err
	}
	if err := pprof.StartCPUProfile(f); err != nil {
		f.Close()
		return nil, err
	}
	return func() { pprof.StopCPUProfile(); f.Close() }, nil
}

func cmdAnalyse(args []string, stdout, stderr io.Writer) (int, error) {
	c := newCommon("analyse", stderr)
	format := c.fs.String("format", "text", "output format: "+strings.Join(report.Formats, ", "))
	failOn := c.fs.String("fail-on", "warning", "exit 1 when a finding has at least this severity: info, warning, error, never")
	stats := c.fs.Bool("stats", false, "print timing to stderr")
	baselinePath := c.fs.String("baseline", "", "ignore findings recorded in this baseline file (default: custos.json \"baseline\")")
	genBaseline := c.fs.String("generate-baseline", "", "write all current findings to this baseline file and exit 0")
	s, err := c.prepare(args)
	if err != nil {
		return 2, err
	}
	if _, ok := failRank[*failOn]; !ok && *failOn != "never" {
		return 2, fmt.Errorf("--fail-on must be info, warning, error or never, got %q", *failOn)
	}
	if !slices.Contains(report.Formats, *format) {
		return 2, fmt.Errorf("unknown format %q (want one of %s)", *format, strings.Join(report.Formats, ", "))
	}
	stop, err := c.startProfile()
	if err != nil {
		return 2, err
	}
	start := time.Now()
	var srcs [][]byte
	if s.engine.NeedsIndex() {
		// The index pass reads the analysed files first; keep those bytes
		// so the analysis does not read every file a second time.
		var ix *index.Index
		ix, srcs = runner.BuildIndexKeep(runner.IndexSources(s.cfg.Root, s.files), len(s.files), s.parse)
		s.engine.SetIndex(ix)
	}
	results := runner.RunSources(s.engine, s.files, srcs, s.parse)
	stop()
	items := report.Items(results)
	if *genBaseline != "" {
		n, err := baseline.Write(*genBaseline, items)
		if err != nil {
			return 2, err
		}
		fmt.Fprintf(stderr, "custos: baseline %s written (%d entries)\n", *genBaseline, n)
		return 0, nil
	}
	fromConfig := false
	if *baselinePath == "" && s.cfg.Baseline != "" {
		*baselinePath = filepath.Join(s.cfg.Root, s.cfg.Baseline)
		fromConfig = true
	}
	if *baselinePath != "" {
		b, err := baseline.Load(*baselinePath)
		if err != nil && fromConfig && errors.Is(err, os.ErrNotExist) {
			// A baseline named in custos.json that was not generated yet
			// suppresses nothing (first run); an explicit --baseline must exist.
			b, err = baseline.Empty(), nil
		}
		if err != nil {
			return 2, err
		}
		var suppressed int
		items, suppressed = b.Filter(items)
		if *stats {
			fmt.Fprintf(stderr, "custos: %d finding(s) suppressed by baseline\n", suppressed)
		}
	}
	if err := report.Write(stdout, *format, items, len(s.files)); err != nil {
		return 2, err
	}
	if *stats {
		fmt.Fprintf(stderr, "custos: %d files, %d rules, PHP %s (%s), %v\n", len(s.files), len(s.engine.Rules()), s.cfg.PHP, s.cfg.PHPSource, time.Since(start).Round(time.Millisecond))
	}
	return exitCode(items, *failOn), nil
}

var failRank = map[string]int{"info": 1, "warning": 2, "error": 3}

func exitCode(items []report.Item, failOn string) int {
	threshold, ok := failRank[failOn]
	if !ok {
		return 0 // never
	}
	for _, it := range items {
		if failRank[it.Severity] >= threshold {
			return 1
		}
	}
	return 0
}

func cmdFix(args []string, stdout, stderr io.Writer) (int, error) {
	c := newCommon("fix", stderr)
	dry := c.fs.Bool("dry-run", false, "do not write files")
	showDiff := c.fs.Bool("diff", false, "print a unified diff of the changes")
	s, err := c.prepare(args)
	if err != nil {
		return 2, err
	}
	stop, err := c.startProfile()
	if err != nil {
		return 2, err
	}
	defer stop()
	if s.engine.NeedsIndex() {
		s.engine.SetIndex(runner.BuildIndex(runner.IndexSources(s.cfg.Root, s.files), s.parse))
	}
	outs := fixAll(s)
	changed, edits, failed := 0, 0, 0
	for i, path := range s.files {
		o := outs[i]
		if errors.Is(o.err, errSkipped) {
			fmt.Fprintf(stderr, "custos: %v, not fixed\n", o.err)
			continue
		}
		if o.err != nil {
			fmt.Fprintln(stderr, "custos:", o.err)
			failed++
			continue
		}
		if o.out == nil {
			continue
		}
		if *showDiff {
			fmt.Fprint(stdout, diff.Unified(filepath.ToSlash(path), string(o.src), string(o.out), 3))
		}
		if !*dry {
			if err := os.WriteFile(path, o.out, o.perm); err != nil {
				fmt.Fprintln(stderr, "custos:", err)
				failed++
				continue
			}
		}
		changed++
		edits += o.applied
	}
	verb := "fixed"
	if *dry {
		verb = "would fix"
	}
	fmt.Fprintf(stderr, "custos: %s %d file(s), %d edit(s)\n", verb, changed, edits)
	if failed > 0 {
		return 2, fmt.Errorf("%d file(s) could not be fixed", failed)
	}
	return 0, nil
}

// fixOutcome is the result of fixing one file; out is nil when nothing
// changed.
type fixOutcome struct {
	src, out []byte
	perm     os.FileMode
	applied  int
	err      error
}

// fixAll fixes every file with one worker per CPU (files are independent;
// the engine is safe for concurrent use). Outcomes are in input order so
// diffs and messages stay deterministic.
func fixAll(s *setup) []fixOutcome {
	outs := make([]fixOutcome, len(s.files))
	jobs := make(chan int)
	var wg sync.WaitGroup
	for w := 0; w < runtime.GOMAXPROCS(0); w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := range jobs {
				src, perm, err := readFixable(s.files[i])
				if err != nil {
					outs[i].err = err
					continue
				}
				res := fix.FixSource(s.engine, s.files[i], src, fix.Options{Parse: s.parse})
				if res.Applied == 0 || string(res.Source) == string(src) {
					continue
				}
				outs[i] = fixOutcome{src: src, out: res.Source, perm: perm, applied: res.Applied}
			}
		}()
	}
	for i := range s.files {
		jobs <- i
	}
	close(jobs)
	wg.Wait()
	return outs
}

// errSkipped marks files fix leaves alone without failing.
var errSkipped = errors.New("not a regular file")

// readFixable reads a file fix may rewrite and returns its permission bits.
// Symlinks (and other non-regular files) are skipped, in dry runs too:
// writing through a link may lead outside the project, and the analysed
// repository is untrusted.
func readFixable(path string) ([]byte, os.FileMode, error) {
	info, err := os.Lstat(path)
	if err == nil && !info.Mode().IsRegular() {
		err = fmt.Errorf("%s: %w", path, errSkipped)
	}
	if err != nil {
		return nil, 0, err
	}
	src, err := runner.ReadSource(path)
	return src, info.Mode().Perm(), err
}

func cmdRules(args []string, stdout, stderr io.Writer) error {
	fs := newFlagSet("rules", stderr)
	asJSON := fs.Bool("json", false, "JSON output")
	if err := fs.Parse(args); err != nil {
		return err
	}
	all, err := catalogue()
	if err != nil {
		return err
	}
	implemented := map[string]bool{}
	for _, r := range registry() {
		implemented[r.ID()] = true
	}
	if *asJSON {
		type row struct {
			meta.Rule
			Implemented bool `json:"implemented"`
		}
		var out []row
		for _, r := range all {
			d, _ := meta.Describe(r.ID)
			r.HasFix = d.Fix // what custos offers, not the upstream catalogue
			out = append(out, row{r, implemented[r.ID]})
		}
		enc := json.NewEncoder(stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(out)
	}
	for _, r := range all {
		mark := " "
		if implemented[r.ID] {
			mark = "✓"
		}
		def := "off"
		if r.EnabledByDefault {
			def = "on"
		}
		fmt.Fprintf(stdout, "%s %-45s %-26s %-8s %s\n", mark, r.ID, r.Group, r.Severity, def)
	}
	return nil
}

func cmdExplain(args []string, stdout io.Writer) error {
	if len(args) != 1 {
		return errors.New("usage: custos explain <rule>")
	}
	m, ok := meta.Lookup(args[0])
	if !ok {
		return fmt.Errorf("unknown rule %q", args[0])
	}
	def := "enabled by default"
	if !m.EnabledByDefault {
		def = "disabled by default"
	}
	if m.Experimental {
		def += ", experimental"
	}
	fix := ""
	if d, ok := meta.Describe(m.ID); ok && d.Fix {
		fix = ", has a quick-fix"
	}
	fmt.Fprintf(stdout, "%s (%s)\n  group: %s, severity: %s, %s%s\n  suppress with: @noinspection %s\n\n", m.ID, m.LegacyID, m.Group, m.Severity, def, fix, m.LegacyID)
	if d, ok := meta.Describe(m.ID); ok {
		fmt.Fprintln(stdout, d.Summary)
		if d.Options != "" {
			fmt.Fprintf(stdout, "\nOptions:\n%s\n", d.Options)
		}
	}
	return nil
}

func cmdLSP(args []string, stdin io.Reader, stdout, stderr io.Writer) error {
	fs := newFlagSet("lsp", stderr)
	_ = fs.Bool("stdio", true, "communicate over stdin/stdout (the only transport)")
	if err := fs.Parse(args); err != nil {
		return err
	}
	lsp.Version = version
	return lsp.Serve(context.Background(), stdin, stdout)
}
