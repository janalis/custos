//go:build unix

// Command benchcmp compares custos with other PHP tools on real projects.
//
//	benchcmp bump     move corpus tags and tool versions to their latest release
//	benchcmp install  install the pinned tools and corpora under -work
//	benchcmp run      measure wall time, peak memory and findings into -results
//	benchcmp render   write the documentation page from -results
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func main() { os.Exit(run(os.Args[1:], os.Stdout, os.Stderr)) }

const usage = "usage: benchcmp bump|install|run|render [flags]"

func run(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		fmt.Fprintln(stderr, usage)
		return 2
	}
	sub := args[0]
	fs := flag.NewFlagSet("benchcmp "+sub, flag.ContinueOnError)
	fs.SetOutput(stderr)
	config := fs.String("config", "bench/competitors.json", "tools, versions and corpora")
	work := fs.String("work", ".cache/bench", "directory of installed tools and corpora")
	configs := fs.String("configs", "bench/configs", "tool configuration templates")
	custos := fs.String("custos", "bin/custos", "custos binary under test")
	results := fs.String("results", "bench/results.json", "measurements")
	page := fs.String("page", "docs/guide/benchmark.md", "generated documentation page")
	packagist := fs.String("packagist", "https://repo.packagist.org", "Packagist metadata server")
	toolFilter := fs.String("tool", "", "comma-separated tool IDs (default: all)")
	corpusFilter := fs.String("corpus", "", "comma-separated corpus names (default: all)")
	if err := fs.Parse(args[1:]); err != nil || fs.NArg() != 0 {
		fmt.Fprintln(stderr, usage)
		return 2
	}
	// Tools run in the corpus directory.
	for _, p := range []*string{work, configs, custos} {
		abs, _ := filepath.Abs(*p) // only fails without a working directory
		*p = abs
	}
	var err error
	switch sub {
	case "render":
		err = renderPage(*results, *page)
	case "bump", "install", "run":
		err = withConfig(*config, func(c *Config) error {
			tools, err := selectByName(c.Tools, func(t Tool) string { return t.ID }, *toolFilter)
			if err != nil {
				return err
			}
			corpora, err := selectByName(c.Corpora, func(c Corpus) string { return c.Name }, *corpusFilter)
			if err != nil {
				return err
			}
			if err := os.MkdirAll(*work, 0o755); err != nil {
				return err
			}
			timeout, _ := c.timeout() // checked by loadConfig
			b := &bench{
				cfg: c, work: *work, configs: *configs, custos: *custos, timeout: timeout,
				log: func(f string, a ...any) { fmt.Fprintf(stderr, f+"\n", a...) },
			}
			switch sub {
			case "bump":
				changes, err := bump(c, *packagist)
				if err != nil {
					return err
				}
				fmt.Fprint(stdout, strings.Join(append(changes, ""), "\n"))
				return writeJSON(*config, c)
			case "install":
				return b.install(tools, corpora, stderr)
			}
			r, err := b.measure(tools, corpora)
			if err != nil {
				return err
			}
			return writeJSON(*results, r)
		})
	default:
		fmt.Fprintln(stderr, usage)
		return 2
	}
	if err != nil {
		fmt.Fprintln(stderr, "benchcmp:", err)
		return 1
	}
	return 0
}

func withConfig(path string, f func(*Config) error) error {
	c, err := loadConfig(path)
	if err != nil {
		return err
	}
	return f(c)
}

func renderPage(results, page string) error {
	data, err := os.ReadFile(results)
	if err != nil {
		return err
	}
	var r Results
	if err := json.Unmarshal(data, &r); err != nil {
		return fmt.Errorf("%s: %w", results, err)
	}
	return os.WriteFile(page, []byte(render(&r)), 0o644)
}
