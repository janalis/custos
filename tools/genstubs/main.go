// Command genstubs compiles JetBrains phpstorm-stubs (Apache-2.0) into the
// embedded builtin symbol index (internal/stubs/stubs.gob.gz).
//
//	go run ./tools/genstubs -src .cache/stubs-src/JetBrains-phpstorm-stubs-<rev>
package main

import (
	"compress/gzip"
	"encoding/gob"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"custos/internal/index"
	"custos/internal/phpver"
	"custos/internal/syntax"
)

func main() {
	src := flag.String("src", "", "phpstorm-stubs checkout")
	out := flag.String("out", "internal/stubs/stubs.gob.gz", "output file")
	rev := flag.String("rev", "", "stubs revision recorded in internal/stubs/VERSION")
	flag.Parse()
	if *src == "" {
		matches, _ := filepath.Glob(".cache/stubs-src/*phpstorm-stubs*")
		if len(matches) == 0 {
			fail(fmt.Errorf("no -src given and no .cache/stubs-src/*phpstorm-stubs* found"))
		}
		*src = matches[0]
	}
	var all []*index.FileSymbols
	files, errs := 0, 0
	err := filepath.WalkDir(*src, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(*src, p)
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
		}
		for _, fn := range fsyms.Functions {
			fn.File = fsyms.Path
		}
		all = append(all, fsyms)
		return nil
	})
	if err != nil {
		fail(err)
	}
	sort.Slice(all, func(i, j int) bool { return all[i].Path < all[j].Path })
	if err := write(*out, all); err != nil {
		fail(err)
	}
	if *rev == "" {
		*rev = filepath.Base(*src)
	}
	_ = os.WriteFile(filepath.Join(filepath.Dir(*out), "VERSION"), []byte(*rev+"\n"), 0o644)
	nc, nf, nk := 0, 0, 0
	for _, f := range all {
		nc += len(f.Classes)
		nf += len(f.Functions)
		nk += len(f.Constants)
	}
	st, _ := os.Stat(*out)
	fmt.Printf("stubs: %d files (%d with parse errors), %d classes, %d functions, %d constants -> %s (%d KB)\n",
		files, errs, nc, nf, nk, *out, st.Size()/1024)
}

func write(path string, all []*index.FileSymbols) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	zw, _ := gzip.NewWriterLevel(f, gzip.BestCompression)
	if err := gob.NewEncoder(zw).Encode(all); err != nil {
		return err
	}
	return zw.Close()
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, "genstubs:", err)
	os.Exit(1)
}
