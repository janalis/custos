//go:build unix

package main

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"maps"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"slices"
	"strconv"
	"strings"
	"sync/atomic"
	"syscall"
	"time"
)

// Results is bench/results.json, rendered into the documentation page.
type Results struct {
	Date    string         `json:"date"`
	Runner  Machine        `json:"runner"`
	PHP     string         `json:"php"`
	Runs    int            `json:"runs"`
	Corpora []CorpusResult `json:"corpora"`
}

// Machine describes the computer the benchmark ran on.
type Machine struct {
	OS       string  `json:"os"`
	CPU      string  `json:"cpu"`
	Cores    int     `json:"cores"`
	MemoryGB float64 `json:"memoryGB"`
}

// CorpusResult holds one row per tool for a corpus.
type CorpusResult struct {
	Name  string `json:"name"`
	Repo  string `json:"repo"`
	Tag   string `json:"tag"`
	Files int    `json:"files"`
	Rows  []Row  `json:"rows"`
}

// Row is a tool's measurements on a corpus: wall times in seconds over the
// measured runs, the peak memory of its whole process group and the
// findings of the last run. Status is "ok", "timeout" or an error.
type Row struct {
	ID       string  `json:"id"`
	Label    string  `json:"label"`
	Kind     string  `json:"kind"`
	Version  string  `json:"version"`
	Status   string  `json:"status"`
	Median   float64 `json:"median,omitempty"`
	Min      float64 `json:"min,omitempty"`
	Max      float64 `json:"max,omitempty"`
	PeakMB   float64 `json:"peakMB,omitempty"`
	Findings int     `json:"findings"`
}

// sample is one execution of a command.
type sample struct {
	wall     time.Duration
	peak     uint64 // bytes
	stdout   []byte
	stderr   []byte
	timedOut bool
	err      error // could not run, or non-zero exit
}

// execute runs argv in dir in its own process group, killed with all its
// workers after timeout, and measures its wall time and peak memory.
func execute(argv []string, dir string, env []string, timeout time.Duration) sample {
	cmd := exec.Command(argv[0], argv[1:]...)
	cmd.Dir = dir
	cmd.Env = env
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	// Workers left behind must not keep Wait blocked on the output pipes.
	cmd.WaitDelay = time.Second
	start := time.Now()
	if err := cmd.Start(); err != nil {
		return sample{err: err}
	}
	pid := cmd.Process.Pid
	stop, peakc := make(chan struct{}), make(chan uint64)
	go func() { peakc <- watch(pid, stop) }()
	var timedOut atomic.Bool
	timer := time.AfterFunc(timeout, func() {
		timedOut.Store(true)
		_ = syscall.Kill(-pid, syscall.SIGKILL)
	})
	err := cmd.Wait()
	s := sample{wall: time.Since(start), err: err, stdout: stdout.Bytes(), stderr: stderr.Bytes()}
	timer.Stop()
	close(stop)
	// The sampled group total when there is one (Linux, unless the tool
	// exited before the first sample), else the kernel's peak of the tool.
	if s.peak = <-peakc; s.peak == 0 {
		s.peak = maxRSS(cmd.ProcessState)
	}
	_ = syscall.Kill(-pid, syscall.SIGKILL) // workers that outlived the tool
	s.timedOut = timedOut.Load()
	return s
}

// bench holds what a run needs besides the configuration.
type bench struct {
	cfg     *Config
	work    string // installed tools and corpora (see install)
	configs string // tool configuration templates
	custos  string // custos binary under test
	timeout time.Duration
	log     func(format string, args ...any)
}

func (b *bench) packageDir(pkg string) string {
	return filepath.Join(b.work, "tools", strings.ReplaceAll(pkg, "/", "-"))
}

func (b *bench) corpusDir(c Corpus) string { return filepath.Join(b.work, "corpora", c.Name) }

// measure runs every selected tool on every selected corpus.
func (b *bench) measure(tools []Tool, corpora []Corpus) (*Results, error) {
	r := &Results{
		Date:   time.Now().UTC().Format(time.DateOnly),
		Runner: machine(),
		PHP:    phpVersion(),
		Runs:   b.cfg.Runs,
	}
	custosVersion := b.custosVersion()
	for _, c := range corpora {
		dir := b.corpusDir(c)
		files, err := countPHP(dir, c.Paths)
		if err != nil {
			return nil, fmt.Errorf("corpus %s: %w (run install first)", c.Name, err)
		}
		cr := CorpusResult{Name: c.Name, Repo: c.Repo, Tag: c.Tag, Files: files}
		for _, t := range tools {
			version := custosVersion
			if t.Package != "" {
				version = b.cfg.Packages[t.Package]
			}
			b.log("%s on %s %s (%d files)", t.Label, c.Name, c.Tag, files)
			row := b.measureTool(t, c)
			row.ID, row.Label, row.Kind, row.Version = t.ID, t.Label, t.Kind, version
			b.log("  %s %.2fs %.0f MB %d findings", row.Status, row.Median, row.PeakMB, row.Findings)
			cr.Rows = append(cr.Rows, row)
		}
		r.Corpora = append(r.Corpora, cr)
	}
	return r, nil
}

// measureTool runs one tool Warmup times, then Runs times, each from cold
// caches.
func (b *bench) measureTool(t Tool, c Corpus) Row {
	var walls []float64
	var row Row
	for i := range b.cfg.Warmup + b.cfg.Runs {
		s, findings, err := b.runOnce(t, c)
		switch {
		case err != nil:
			return Row{Status: err.Error()}
		case s.timedOut:
			return Row{Status: "timeout"}
		}
		if i < b.cfg.Warmup {
			continue
		}
		walls = append(walls, s.wall.Seconds())
		row.PeakMB = max(row.PeakMB, float64(s.peak)/(1<<20))
		row.Findings = findings
	}
	slices.Sort(walls)
	row.Status = "ok"
	row.Min, row.Max, row.Median = walls[0], walls[len(walls)-1], walls[len(walls)/2]
	if len(walls)%2 == 0 {
		row.Median = (walls[len(walls)/2-1] + walls[len(walls)/2]) / 2
	}
	return row
}

// runOnce runs a tool with a fresh cache directory and configuration.
func (b *bench) runOnce(t Tool, c Corpus) (sample, int, error) {
	cache, err := os.MkdirTemp(b.work, "run-")
	if err != nil {
		return sample{}, 0, err
	}
	defer os.RemoveAll(cache)
	corpus := b.corpusDir(c)
	if err := renderConfigs(b.configs, cache, corpus, c.Paths); err != nil {
		return sample{}, 0, err
	}
	vars := map[string]string{
		"{pkg}":     b.packageDir(t.Package),
		"{custos}":  b.custos,
		"{corpus}":  corpus,
		"{configs}": cache,
		"{cache}":   cache,
		"{cpus}":    strconv.Itoa(runtime.NumCPU()),
	}
	argv := expand(t.Cmd, vars, c.Paths)
	env := os.Environ()
	for _, k := range slices.Sorted(maps.Keys(t.Env)) {
		env = append(env, k+"="+expand([]string{t.Env[k]}, vars, nil)[0])
	}
	s := execute(argv, corpus, env, b.timeout)
	if s.timedOut {
		return s, 0, nil
	}
	out := s.stdout
	if t.Stream == "stderr" {
		out = s.stderr
	}
	n, perr := parsers[t.Findings](out)
	if perr != nil {
		msg := "error: " + perr.Error()
		if s.err != nil {
			msg += " (" + s.err.Error() + ")"
		}
		return s, 0, errors.New(msg + firstLine(s.stderr))
	}
	return s, n, nil
}

// firstLine returns the first line of b, shortened and prefixed by ": "
// ("" for empty output).
func firstLine(b []byte) string {
	l, _, _ := strings.Cut(strings.TrimSpace(string(b)), "\n")
	if l = strings.TrimSpace(l); l == "" {
		return ""
	}
	if r := []rune(l); len(r) > 160 {
		l = string(r[:160]) + "…"
	}
	return ": " + l
}

// expand substitutes vars in each argument; an argument "{paths}" becomes
// one argument per corpus path.
func expand(args []string, vars map[string]string, paths []string) []string {
	var out []string
	for _, a := range args {
		if a == "{paths}" {
			out = append(out, paths...)
			continue
		}
		for k, v := range vars {
			a = strings.ReplaceAll(a, k, v)
		}
		out = append(out, a)
	}
	return out
}

// renderConfigs copies the templates of src to the run's cache directory,
// replacing @CORPUS@ and @CACHE@, and repeating each line holding @PATH@
// once per corpus path.
func renderConfigs(src, cache, corpus string, paths []string) error {
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}
	vars := strings.NewReplacer("@CORPUS@", corpus, "@CACHE@", cache)
	for _, e := range entries {
		data, err := os.ReadFile(filepath.Join(src, e.Name()))
		if err != nil {
			return err
		}
		var out strings.Builder
		for _, line := range strings.SplitAfter(vars.Replace(string(data)), "\n") {
			if !strings.Contains(line, "@PATH@") {
				out.WriteString(line)
				continue
			}
			for _, p := range paths {
				out.WriteString(strings.ReplaceAll(line, "@PATH@", p))
			}
		}
		// The directory was just created for this run.
		_ = os.WriteFile(filepath.Join(cache, e.Name()), []byte(out.String()), 0o644)
	}
	return nil
}

// countPHP counts the .php files under the corpus paths.
func countPHP(dir string, paths []string) (int, error) {
	n := 0
	for _, p := range paths {
		err := filepath.WalkDir(filepath.Join(dir, p), func(_ string, d fs.DirEntry, err error) error {
			if err == nil && !d.IsDir() && strings.HasSuffix(d.Name(), ".php") {
				n++
			}
			return err
		})
		if err != nil {
			return 0, err
		}
	}
	return n, nil
}

func (b *bench) custosVersion() string {
	out, err := exec.Command(b.custos, "version").Output()
	if err != nil {
		return "unknown"
	}
	f := strings.Fields(string(out)) // "custos vX.Y.Z"
	return strings.TrimPrefix(f[len(f)-1], "v")
}

func phpVersion() string {
	out, err := exec.Command("php", "-r", "echo PHP_VERSION;").Output()
	if err != nil {
		return "unknown"
	}
	return strings.TrimSpace(string(out))
}

var errNoMatch = errors.New("no match")

// selectByName keeps the items whose name is in the comma-separated filter
// (all when it is empty).
func selectByName[T any](items []T, name func(T) string, filter string) ([]T, error) {
	if filter == "" {
		return items, nil
	}
	want := strings.Split(filter, ",")
	var out []T
	for _, it := range items {
		if slices.Contains(want, name(it)) {
			out = append(out, it)
		}
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("%w for %q", errNoMatch, filter)
	}
	return out, nil
}
