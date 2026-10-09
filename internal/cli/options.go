package cli

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime/pprof"
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/meta"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
	"custos/internal/project"
	"custos/internal/project/config"
)

// common holds flags shared by analyse and fixing.
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
	c.fs.StringVar(&c.php, "php", "", "target PHP version (default: custos.json, composer.json, else "+phpversion.Default.String()+")")
	c.fs.StringVar(&c.configDir, "config", "", "directory to look for custos.json (default: first path)")
	c.fs.StringVar(&c.only, "rule", "", "comma-separated rule IDs to run (default: enabled rules)")
	c.fs.BoolVar(&c.all, "all", false, "enable every rule, including disabled-by-default ones")
	c.fs.StringVar(&c.exclude, "exclude", "", "comma-separated extra directories to exclude")
	c.fs.StringVar(&c.profile, "cpuprofile", "", "write a CPU profile to this file")
	return c
}

type setup struct {
	project *project.Project
	cfg     *config.Config
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
		v, err := phpversion.Parse(c.php)
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
	if len(paths) == 0 {
		for _, p := range cfg.Paths {
			paths = append(paths, filepath.Join(cfg.Root, p))
		}
	}
	exclude := cfg.Exclude
	if c.exclude != "" {
		exclude = append(exclude, strings.Split(c.exclude, ",")...)
	}
	p, err := project.Open(registry(), project.Options{Root: cfg.Root, Paths: paths, Exclude: exclude, Analysis: acfg, Parse: syntax.Options{Version: cfg.PHP, ShortOpenTag: cfg.ShortOpenTag}})
	if err != nil {
		return nil, err
	}
	return &setup{project: p, cfg: cfg}, nil
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
