package cli

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"runtime/pprof"
	"strings"

	"github.com/janalis/custos/internal/inspection/analysis"
	"github.com/janalis/custos/internal/inspection/meta"
	phpversion "github.com/janalis/custos/internal/php/version"
	"github.com/janalis/custos/internal/project"
	"github.com/janalis/custos/internal/project/config"
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
	memProf   string
	noCache   bool
	flowOrder bool
	parseOnce bool
	// rules lists the rules the command may run; nil runs every registered
	// rule.
	rules func() []analysis.Rule
}

// cacheDir is where declarations of dependencies are kept between runs
// with the cross-file summaries of the analyzed files (see
// project.CacheDir); "" with --no-cache.
func (c *common) cacheDir() string {
	if c.noCache {
		return ""
	}
	return project.CacheDir()
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
	c.fs.StringVar(&c.memProf, "memprofile", "", "write an allocation profile to this file")
	c.fs.BoolVar(&c.noCache, "no-cache", false, "do not reuse or save declarations of dependencies and cross-file summaries between runs")
	c.fs.BoolVar(&c.flowOrder, "flow-order", false, "use the last saved cross-file summary order as a scheduling hint")
	c.fs.BoolVar(&c.parseOnce, "parse-once", false, "keep each analyzed file's syntax tree between phases: faster, more memory")
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
	acfg.LabelRules = c.profile != ""
	acfg.ProjectPath = project.DiskProjectPaths(cfg.Root)
	if c.only != "" {
		for _, id := range strings.Split(c.only, ",") {
			m, ok := meta.Lookup(strings.TrimSpace(id))
			if !ok {
				return nil, fmt.Errorf("unknown rule %q", id)
			}
			acfg.Only = append(acfg.Only, m.ID)
		}
	}
	var scope []string
	for _, p := range cfg.Paths {
		scope = append(scope, filepath.Join(cfg.Root, p))
	}
	if len(paths) == 0 {
		paths = scope
	}
	exclude := cfg.Exclude
	if c.exclude != "" {
		exclude = append(exclude, strings.Split(c.exclude, ",")...)
	}
	rules := registry
	if c.rules != nil {
		rules = c.rules
	}
	p, err := project.Open(rules(), project.Options{Root: cfg.Root, Paths: paths, Exclude: exclude, Scope: scope, Cache: c.cacheDir(), FlowOrder: c.flowOrder || cfg.FlowOrder, ParseOnce: c.parseOnce || cfg.ParseOnce, Analysis: acfg, Parse: cfg.Parse()})
	if err != nil {
		return nil, err
	}
	return &setup{project: p, cfg: cfg}, nil
}

func (c *common) startProfile() (func(), error) {
	stopCPU, err := c.startCPUProfile()
	if err != nil {
		return nil, err
	}
	if c.memProf == "" {
		return stopCPU, nil
	}
	f, err := os.Create(c.memProf)
	if err != nil {
		stopCPU()
		return nil, err
	}
	// Run turns allocation sampling off; sample from here, at the runtime's
	// default rate, so the profile covers the run as the CPU profile does.
	runtime.MemProfileRate = memProfileRate
	return func() {
		stopCPU()
		// The profile holds the allocations up to the last collection. Like
		// the CPU profile's, a failed write only leaves a truncated profile;
		// the command's own result stands.
		runtime.GC()
		_ = pprof.Lookup("allocs").WriteTo(f, 0)
		f.Close()
	}, nil
}

// memProfileRate is the runtime's default allocation sampling rate.
const memProfileRate = 512 * 1024

func (c *common) startCPUProfile() (func(), error) {
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
