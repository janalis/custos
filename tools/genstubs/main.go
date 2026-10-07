// Command genstubs compiles JetBrains phpstorm-stubs (Apache-2.0) into the
// embedded builtin symbol index (internal/stubs/stubs.gob.gz).
//
//	go run ./tools/genstubs -src .cache/stubs-src/JetBrains-phpstorm-stubs-<rev>
package main

import (
	"compress/gzip"
	"encoding/gob"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"custos/internal/index"
	"custos/internal/phpver"
	"custos/internal/syntax"
	"custos/internal/types"
)

// defaultSrc locates the stubs checkout when -src is not given (`make stubs`).
var defaultSrc = ".cache/stubs-src/*phpstorm-stubs*"

func main() { os.Exit(run(os.Args[1:], os.Stdout, os.Stderr)) }

func run(args []string, stdout, stderr io.Writer) int {
	fl := flag.NewFlagSet("genstubs", flag.ContinueOnError)
	fl.SetOutput(stderr)
	src := fl.String("src", "", "phpstorm-stubs checkout")
	out := fl.String("out", "internal/stubs/stubs.gob.gz", "output file")
	rev := fl.String("rev", "", "stubs revision recorded in internal/stubs/VERSION")
	if err := fl.Parse(args); err != nil {
		if err == flag.ErrHelp {
			return 0
		}
		return 2
	}
	if err := generate(*src, *out, *rev, stdout); err != nil {
		fmt.Fprintln(stderr, "genstubs:", err)
		return 1
	}
	return 0
}

func generate(src, out, rev string, stdout io.Writer) error {
	if src == "" {
		matches, _ := filepath.Glob(defaultSrc)
		if len(matches) == 0 {
			return fmt.Errorf("no -src given and no %s found", defaultSrc)
		}
		src = matches[0]
	}
	var all []*index.FileSymbols
	files, errs := 0, 0
	err := filepath.WalkDir(src, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(src, p)
		if d.IsDir() {
			switch d.Name() {
			case "tests", "meta", ".github", "vendor":
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(p, ".php") || d.Name() == "PhpStormStubsMap.php" {
			return nil
		}
		b, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		f := syntax.Parse(filepath.ToSlash(rel), b, syntax.Options{Version: phpver.Max})
		files++
		if len(f.Errors) > 0 {
			errs++
		}
		fsyms := index.Extract(f)
		fsyms.Path = "stubs/" + filepath.ToSlash(rel)
		for _, c := range fsyms.Classes {
			c.File = fsyms.Path
			for _, m := range c.Methods {
				noShapes(&m.DocReturn)
				noParamShapes(m.Params)
			}
			for _, p := range c.Props {
				noShapes(&p.DocType)
			}
		}
		for _, fn := range fsyms.Functions {
			fn.File = fsyms.Path
			noShapes(&fn.DocReturn)
			noParamShapes(fn.Params)
		}
		all = append(all, fsyms)
		return nil
	})
	if err != nil {
		return err
	}
	sort.Slice(all, func(i, j int) bool { return all[i].Path < all[j].Path })
	if err := write(out, all); err != nil {
		return err
	}
	if rev == "" {
		rev = filepath.Base(src)
	}
	_ = os.WriteFile(filepath.Join(filepath.Dir(out), "VERSION"), []byte(rev+"\n"), 0o644)
	nc, nf, nk := 0, 0, 0
	for _, f := range all {
		nc += len(f.Classes)
		nf += len(f.Functions)
		nk += len(f.Constants)
	}
	st, _ := os.Stat(out)
	fmt.Fprintf(stdout, "stubs: %d files (%d with parse errors), %d classes, %d functions, %d constants -> %s (%d KB)\n",
		files, errs, nc, nf, nk, out, st.Size()/1024)
	return nil
}

// noShapes drops the array facts (shapes, non-emptiness) from a stub doc
// type: the stubs' shapes are not reliable enough to type builtin results
// (pathinfo() documents its optional keys as required).
func noShapes(s *string) {
	if *s == "" {
		return
	}
	t := types.FromDoc(*s, nil)
	if !t.IsUnknown() {
		*s = t.WithoutArrayInfo().DocString()
	}
}

func noParamShapes(ps []index.Param) {
	for i := range ps {
		noShapes(&ps[i].DocType)
	}
}

func write(path string, all []*index.FileSymbols) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	zw, _ := gzip.NewWriterLevel(f, gzip.BestCompression)
	// Every step's error is reported, the file's Close included (a failed
	// final flush would otherwise leave a truncated index unnoticed).
	return errors.Join(gob.NewEncoder(zw).Encode(all), zw.Close(), f.Close())
}
