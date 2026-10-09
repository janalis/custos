package cli

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"custos/internal/diagnostic"
	"custos/internal/output/report"
	"custos/internal/project/baseline"
)

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
	results := s.project.AnalyzeReport()
	stop()
	items := diagnostic.Items(results)
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
	if err := report.Write(stdout, *format, items, len(s.project.Files)); err != nil {
		return 2, err
	}
	if *stats {
		fmt.Fprintf(stderr, "custos: %d files, %d rules, PHP %s (%s), %v\n", len(s.project.Files), len(s.project.Engine.Rules()), s.cfg.PHP, s.cfg.PHPSource, time.Since(start).Round(time.Millisecond))
	}
	return exitCode(items, *failOn), nil
}

var failRank = map[string]int{"info": 1, "warning": 2, "error": 3}

func exitCode(items []diagnostic.Item, failOn string) int {
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
