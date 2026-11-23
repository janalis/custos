// Command gendeprecations writes the table of PHP's own functions, methods
// and constants that DeprecatedFeature reports, each with the PHP version
// that deprecates it (`make deprecations`).
//
// It runs one PHP binary per release from 8.0 on: the newest lists every
// deprecated name it declares (reflection, and reading each constant), and
// each binary tells which of those names it deprecates. A name dates from
// the oldest binary that deprecates it, so every release must be given.
package main

import (
	_ "embed"
	"errors"
	"flag"
	"fmt"
	"go/format"
	"io"
	"maps"
	"os"
	"os/exec"
	"slices"
	"strings"

	phpversion "github.com/janalis/custos/internal/php/version"
)

//go:embed probe.php
var probe string

// skipped are aliases AliasFunctionsUsage reports with their replacement.
var skipped = map[string]bool{
	"f:is_double": true, "f:is_integer": true, "f:is_long": true, "f:doubleval": true,
	"f:socket_set_timeout": true, "f:mysqli_execute": true,
}

// binaries is the repeatable -php VERSION=PATH flag.
type binaries map[phpversion.Version]string

func (b binaries) String() string { return fmt.Sprint(map[phpversion.Version]string(b)) }

func (b binaries) Set(s string) error {
	ver, path, ok := strings.Cut(s, "=")
	if !ok || path == "" {
		return fmt.Errorf("want VERSION=PATH, got %q", s)
	}
	v, err := phpversion.Parse(ver)
	if err != nil {
		return err
	}
	b[v] = path
	return nil
}

func main() { os.Exit(run(os.Args[1:], os.Stdout, os.Stderr)) }

func run(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("gendeprecations", flag.ContinueOnError)
	fs.SetOutput(stderr)
	bins := binaries{}
	fs.Var(bins, "php", "VERSION=PATH of a PHP binary; one per release from 8.0 on, repeated")
	out := fs.String("out", "internal/inspection/rules/deprecatedfeature/data.go", "Go file to write")
	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return 0
		}
		return 2
	}
	n, err := generate(bins, *out, stderr)
	if err != nil {
		fmt.Fprintln(stderr, "gendeprecations:", err)
		return 1
	}
	fmt.Fprintf(stdout, "gendeprecations: %d deprecations -> %s\n", n, *out)
	return 0
}

// generate dates every deprecation and writes the table to out, warning on
// warn about binaries that lack an extension declaring deprecations: those
// may be dated late.
func generate(bins binaries, out string, warn io.Writer) (int, error) {
	versions := slices.Sorted(maps.Keys(bins))
	// Every release up to the newest supported one: a partial set would
	// date deprecations late or drop the newest.
	newest := phpversion.Max
	for _, r := range phpversion.Releases {
		if r.AtLeast(phpversion.PHP80) && bins[r] == "" {
			return 0, fmt.Errorf("no PHP %s binary: every release from 8.0 to %s dates the deprecations", r, newest)
		}
	}
	listed, err := probeRun(bins[newest], "list", nil)
	if err != nil {
		return 0, err
	}
	names, extension := make([]string, 0, len(listed)), map[string]string{}
	for _, line := range listed {
		name, ext, _ := strings.Cut(line, "\t")
		names = append(names, name)
		extension[name] = ext
	}
	since := map[string]phpversion.Version{}
	for _, v := range versions {
		if v.Below(phpversion.PHP80) {
			continue
		}
		got, err := probeRun(bins[v], "check", names)
		if err != nil {
			return 0, err
		}
		loaded := map[string]bool{}
		for _, n := range got {
			if ext, ok := strings.CutPrefix(n, "x:"); ok {
				loaded[ext] = true
			} else if _, seen := since[n]; !seen {
				since[n] = v
			}
		}
		missing := map[string]bool{}
		for _, ext := range extension {
			if ext != "" && !loaded[ext] {
				missing[ext] = true
			}
		}
		if len(missing) > 0 {
			fmt.Fprintf(warn, "gendeprecations: PHP %s lacks %s: their deprecations may be dated late\n", v, strings.Join(slices.Sorted(maps.Keys(missing)), ", "))
		}
	}
	for _, n := range names {
		if _, ok := since[n]; !ok {
			return 0, fmt.Errorf("PHP %s lists %s as deprecated but does not deprecate it", newest, n)
		}
	}
	for n := range skipped {
		delete(since, n)
	}
	src, err := formatSource([]byte(render(since)))
	if err != nil {
		return 0, err
	}
	return len(since), os.WriteFile(out, src, 0o644)
}

// formatSource formats the generated table (a variable for tests).
var formatSource = format.Source

// probeRun runs the probe in mode with names on stdin.
func probeRun(bin, mode string, names []string) ([]string, error) {
	code := strings.TrimPrefix(probe, "<?php")
	cmd := exec.Command(bin, "-d", "error_reporting=-1", "-d", "display_errors=0", "-r", code, "--", mode)
	cmd.Stdin = strings.NewReader(strings.Join(names, "\n") + "\n")
	b, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("%s %s: %w", bin, mode, err)
	}
	var out []string
	for _, line := range strings.Split(string(b), "\n") {
		if len(line) > 2 && line[1] == ':' {
			out = append(out, line)
		}
	}
	return out, nil
}

// render writes the Go source of the table.
func render(since map[string]phpversion.Version) string {
	funcs, consts, classConsts := map[string]string{}, map[string]string{}, map[string]string{}
	methods := map[string][]string{}
	for _, n := range slices.Sorted(maps.Keys(since)) {
		v := "phpversion.PHP" + strings.ReplaceAll(since[n].String(), ".", "")
		name := n[2:]
		switch n[0] {
		case 'f':
			funcs[strings.ToLower(name)] = v
		case 'c':
			consts[name] = v
		case 'k':
			class, constant, _ := strings.Cut(name, "::")
			classConsts[strings.ToLower(class)+"::"+constant] = v
		default: // m
			class, method, _ := strings.Cut(name, "::")
			key := strings.ToLower(method)
			methods[key] = append(methods[key], fmt.Sprintf("{%q, %s}", class, v))
		}
	}
	var b strings.Builder
	b.WriteString("// Code generated by tools/gendeprecations (make deprecations). DO NOT EDIT.\n\npackage deprecatedfeature\n\n")
	b.WriteString("import phpversion \"github.com/janalis/custos/internal/php/version\"\n")
	table(&b, "deprecatedFunctions are PHP's own functions by lower-case name, with the version that deprecates them.", "deprecatedFunctions", "phpversion.Version", funcs)
	lists := map[string]string{}
	for k, v := range methods {
		lists[k] = "{" + strings.Join(v, ", ") + "}"
	}
	table(&b, "deprecatedMethods are PHP's own methods by lower-case name: the classes declaring them deprecated, with the version.", "deprecatedMethods", "[]deprecatedMethod", lists)
	table(&b, "deprecatedClassConstants are PHP's own class constants by lower-case class and constant name.", "deprecatedClassConstants", "phpversion.Version", classConsts)
	table(&b, "deprecatedConstants are PHP's own global constants.", "deprecatedConstants", "phpversion.Version", consts)
	return b.String()
}

func table(b *strings.Builder, doc, name, typ string, entries map[string]string) {
	fmt.Fprintf(b, "\n// %s\nvar %s = map[string]%s{\n", doc, name, typ)
	for _, k := range slices.Sorted(maps.Keys(entries)) {
		fmt.Fprintf(b, "\t%q: %s,\n", k, entries[k])
	}
	b.WriteString("}\n")
}
