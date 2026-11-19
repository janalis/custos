package cli

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/janalis/custos/internal/diagnostic/markup"
	"github.com/janalis/custos/internal/fixing"
	"github.com/janalis/custos/internal/inspection/analysis"
	"github.com/janalis/custos/internal/php/syntax"
	phpversion "github.com/janalis/custos/internal/php/version"
	"github.com/janalis/custos/internal/platform/safeio"
	"github.com/janalis/custos/internal/project/config"
)

// fixtureSettings is the JSON sidecar of a rule fixture.
type fixtureSettings struct {
	PHP  string `json:"php,omitempty"`  // target PHP version
	Path string `json:"path,omitempty"` // project path the rule sees, for its paths globs
}

// cmdTestRules runs every rule that has a fixture directory over its
// marked-up fixtures and compares findings and fixes.
func cmdTestRules(args []string, stdout, stderr io.Writer) (int, error) {
	flags := newFlagSet("test-rules", stderr)
	configDir := flags.String("config", ".", "directory to look for custos.json from")
	if err := flags.Parse(args); err != nil {
		return 2, err
	}
	cfg, err := config.Load(*configDir)
	if err != nil {
		return 2, err
	}
	dir := filepath.Join(cfg.Root, filepath.FromSlash(cfg.CustomRuleTests))
	if flags.NArg() > 0 {
		// Absolute, so fixtures' project paths are relative to the root.
		dir = flags.Arg(0)
		if abs, err := filepath.Abs(dir); err == nil {
			dir = abs
		}
	}
	// Fixtures come with the project: read them only within their
	// directory, bounded, and never through a link out of it.
	root, err := os.OpenRoot(dir)
	if err != nil {
		return 2, fmt.Errorf("rule fixtures: %w", err)
	}
	defer root.Close()
	// The root opened, so the directory is readable; one that fails
	// meanwhile lists no fixtures, which is reported below.
	entries, _ := fs.ReadDir(root.FS(), ".")
	rules := cfg.Registry(rulesWithExtra())
	ran, failed := 0, 0
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		m, ok := analysis.Lookup(rules, e.Name())
		if !ok {
			return 2, fmt.Errorf("rule fixtures: %s: unknown rule", e.Name())
		}
		fixtures, _ := fs.Glob(root.FS(), e.Name()+"/*.php") // the pattern is valid
		for _, name := range fixtures {
			if strings.HasSuffix(name, ".fixed.php") {
				continue
			}
			ran++
			problems, err := checkFixture(cfg, rules, m.ID, root, filepath.Join(dir, filepath.FromSlash(name)), name)
			if err != nil {
				return 2, fmt.Errorf("%s: %w", name, err)
			}
			if len(problems) == 0 {
				fmt.Fprintf(stdout, "ok    %s\n", name)
				continue
			}
			failed++
			fmt.Fprintf(stdout, "FAIL  %s\n", name)
			for _, p := range problems {
				fmt.Fprintf(stdout, "      %s\n", strings.ReplaceAll(p, "\n", "\n      "))
			}
		}
	}
	if ran == 0 {
		return 2, fmt.Errorf("rule fixtures: none in %s (one directory per rule ID, holding .php fixtures)", dir)
	}
	fmt.Fprintf(stdout, "\n%d fixture(s), %d failed\n", ran, failed)
	if failed > 0 {
		return 1, nil
	}
	return 0, nil
}

// checkFixture runs rule id alone over the fixture name of root, at path
// on disk, and returns what differs from its markup and its .fixed.php.
func checkFixture(cfg *config.Config, rules []analysis.Rule, id string, root *os.Root, path, name string) ([]string, error) {
	read := func(name string) ([]byte, error) { return safeio.ReadRootFile(root, name, config.MaxFileSize) }
	marked, err := read(name)
	if err != nil {
		return nil, err
	}
	src, want, err := markup.Parse(marked)
	if err != nil {
		return nil, err
	}
	base := strings.TrimSuffix(name, ".php")
	var settings fixtureSettings
	if b, err := read(base + ".json"); err == nil {
		if err := json.Unmarshal(b, &settings); err != nil {
			return nil, fmt.Errorf("%s.json: %w", filepath.Base(base), err)
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return nil, err
	}
	acfg := cfg.Analysis()
	acfg.Only = []string{id}
	if settings.PHP != "" {
		if acfg.PHP, err = phpversion.Parse(settings.PHP); err != nil {
			return nil, err
		}
	}
	project := settings.Path
	if project == "" {
		rel, _ := filepath.Rel(cfg.Root, path)
		project = filepath.ToSlash(rel)
	}
	acfg.ProjectPath = func(string) string { return project }
	engine, err := analysis.NewEngine(rules, acfg)
	if err != nil {
		return nil, err
	}
	opts := cfg.Parse()
	opts.Version = acfg.PHP
	var got []markup.Expectation
	for _, f := range engine.Analyze(syntax.Parse(path, src, opts)) {
		got = append(got, markup.Expectation{Severity: f.Severity, Message: f.Message, Start: int(f.Span.Start), End: int(f.Span.End)})
	}
	markup.Sort(got)
	lines := syntax.NewLineIndex(src)
	at := func(e markup.Expectation) string {
		line, col := lines.RuneColumn(uint32(e.Start))
		return fmt.Sprintf("%d:%d %s %q: %q", line+1, col+1, e.Severity, src[e.Start:e.End], e.Message)
	}
	var problems []string
	missing, unexpected := markup.Diff(want, got, true)
	for _, e := range missing {
		problems = append(problems, "missing    "+at(e))
	}
	for _, e := range unexpected {
		problems = append(problems, "unexpected "+at(e))
	}
	if fixed, err := read(base + ".fixed.php"); err == nil {
		out := fixing.FixSource(engine, path, src, fixing.Options{Parse: opts}).Source
		if !bytes.Equal(markup.NormalizeWS(out), markup.NormalizeWS(fixed)) {
			problems = append(problems, fmt.Sprintf("fixed source differs from %s.fixed.php:\n%s", filepath.Base(base), out))
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return nil, err
	}
	return problems, nil
}
