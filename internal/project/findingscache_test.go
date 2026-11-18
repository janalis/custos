package project

import (
	"encoding/binary"
	"os"
	"path/filepath"
	"testing"

	"custos/internal/diagnostic"
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/catalogue"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
	"custos/internal/semantic/flow"
)

// findingsRun is one run of the findings oracle: its findings, rendered,
// and how many files it took from the cache.
type findingsRun struct {
	findings string
	kept     int64
}

// runFindings analyzes the src directory of root as the command line does,
// with the cache directory dir ("" for none); configure adjusts the
// project before the run.
func runFindings(t *testing.T, root, dir string, configure func(*Project)) findingsRun {
	t.Helper()
	var kept int64 = -1
	old := reported
	defer func() { reported = old }()
	reported = func(fc *findingsCache) {
		if fc != nil {
			kept = fc.kept.Load()
		}
	}
	src := filepath.Join(root, "src")
	opt := syntax.Options{Version: phpversion.PHP84}
	p, err := Open(catalogue.All(), Options{Root: root, Paths: []string{src}, Scope: []string{src}, Cache: dir, Analysis: analysis.Config{PHP: opt.Version, EnableAll: true, ProjectPath: DiskProjectPaths(root)}, Parse: opt})
	if err != nil {
		t.Fatal(err)
	}
	if configure != nil {
		configure(p)
	}
	return findingsRun{renderResults(p.AnalyzeReport()), kept}
}

// A run taking findings from the cache reports what a run without it does,
// and takes those of every file nothing it depends on changed in: after an
// edit that changes no contract only the edited file is analyzed again;
// after a change of contract, of declarations, of the configuration or of
// the baseline the checks see, every file is.
func TestFindingsCacheOracle(t *testing.T) {
	root := t.TempDir()
	src := func(name string) string { return filepath.Join(root, "src", name) }
	files := map[string]string{
		"helper.php":  "<?php function helper($x) { return $x; }",
		"shell.php":   "<?php function shell($c) { system(helper($c)); }",
		"caller.php":  "<?php shell($_GET['x']); $a = !!!$b;",
		"other.php":   "<?php final class Other { public function go(?Other $o) { return $o->go(null); } }",
		"pkg/lib.php": "<?php $p = !!!$q;",
	}
	for name, code := range files {
		writeFile(t, src(name), code)
	}
	dir := t.TempDir()
	check := func(label string, cached findingsRun, kept int64) {
		t.Helper()
		if fresh := runFindings(t, root, "", nil); cached.findings != fresh.findings {
			t.Fatalf("%s: findings differ from a run without the cache:\n%s\nwant\n%s", label, cached.findings, fresh.findings)
		}
		if cached.kept != kept {
			t.Fatalf("%s: %d files taken from the cache, want %d", label, cached.kept, kept)
		}
	}
	check("cold", runFindings(t, root, dir, nil), 0)
	check("unchanged", runFindings(t, root, dir, nil), 5)
	writeFile(t, src("caller.php"), files["caller.php"]+"\n// comment")
	check("comment", runFindings(t, root, dir, nil), 4)
	// A project of its own below src: path globs see its files otherwise
	// (and its composer.json is analyzed from now on).
	writeFile(t, src("pkg/composer.json"), "{}")
	check("nested project", runFindings(t, root, dir, nil), 0)
	writeFile(t, src("helper.php"), "<?php function helper($x) { return (int) $x; }")
	check("contract", runFindings(t, root, dir, nil), 0)
	check("settled", runFindings(t, root, dir, nil), 6)
	writeFile(t, src("other.php"), "<?php final class Other { public function go(Other $o) { return $o->go($o); } }")
	check("declarations", runFindings(t, root, dir, nil), 0)
	check("configuration", runFindings(t, root, dir, func(p *Project) {
		p.Engine, _ = analysis.NewEngine(catalogue.All(), analysis.Config{PHP: phpversion.PHP84})
	}), 0)
	check("back", runFindings(t, root, dir, nil), 0)
	// Composer's installed packages are part of the world rules see.
	writeFile(t, filepath.Join(root, "vendor", "composer", "installed.json"), `{"packages": []}`)
	check("installed", runFindings(t, root, dir, nil), 0)
	recording := func(p *Project) {
		p.Engine = p.Engine.WithRecorded(func(string, []byte) func(diagnostic.Finding, []diagnostic.Finding) bool { return nil })
	}
	if run := runFindings(t, root, dir, recording); run.kept != -1 {
		t.Fatal("an engine seeing a baseline of unknown key used the cache")
	}
	check("baseline", runFindings(t, root, dir, func(p *Project) {
		recording(p)
		p.Options.RecordedKey = "none"
	}), 0)
	check("same baseline", runFindings(t, root, dir, func(p *Project) {
		recording(p)
		p.Options.RecordedKey = "none"
	}), 6)
	// An unreadable file is reported, never kept.
	if err := os.Chmod(src("caller.php"), 0); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chmod(src("caller.php"), 0o644) }()
	if run := runFindings(t, root, dir, nil); run.kept > 5 {
		t.Fatalf("unreadable file taken from the cache (%d)", run.kept)
	}
}

// A findings file that does not decode, is keyed otherwise or was computed
// against another snapshot gives no entries; one that cannot be encoded is
// not saved.
func TestFindingsCacheDecode(t *testing.T) {
	snap := flow.NewSnapshotWithOptions(flow.Options{})
	encode := func(s *flow.Snapshot) []byte {
		e := flow.NewEncoder()
		e.Snapshot(s.Contracts())
		return e.Bytes()
	}
	field := func(b []byte) []byte { return append(binary.AppendUvarint(nil, uint64(len(b))), b...) }
	file := func(key string, snapshot, body []byte) []byte {
		out := append([]byte(findingsCacheMagic), field([]byte(key))...)
		return append(append(out, field(snapshot)...), body...)
	}
	fc := &findingsCache{key: "k", snapshot: snap}
	for label, b := range map[string][]byte{
		"magic":        []byte("other"),
		"key":          file("other", encode(snap), nil),
		"field":        append([]byte(findingsCacheMagic), field([]byte("k"))...),
		"decoder":      file("k", []byte{0xff}, nil),
		"trailing":     file("k", append(encode(snap), 0), nil),
		"no snapshot":  file("k", encode(nil), nil),
		"other":        file("k", encode(snap.WithFile(&flow.File{Path: "a.php", Summaries: map[string]flow.Summary{"f": {}}}).WithoutFile("b.php")), nil),
		"body":         file("k", encode(snap), []byte("not gob")),
		"long field":   append([]byte(findingsCacheMagic), 9, 'k'),
		"empty prefix": nil,
	} {
		if fc.decode(b) != nil {
			t.Fatalf("%s: entries taken", label)
		}
	}
	none := &findingsCache{key: "k"}
	if none.decode(file("k", encode(snap), nil)) != nil {
		t.Fatal("entries computed against a snapshot taken by a run without one")
	}
	node := syntax.Parse("a.php", []byte("<?php $x;"), syntax.Options{}).Stmts[0].(*syntax.ExprStmt).Expr
	held := flow.NewSnapshot(&flow.File{Path: "a.php", Summaries: map[string]flow.Summary{"f": {Return: flow.Value{Expr: node}}}})
	unencodable := &findingsCache{path: filepath.Join(t.TempDir(), "f.bin"), snapshot: held, files: []string{"a.php"}, fresh: []*cachedResult{{}}}
	if err := unencodable.save(); err == nil {
		t.Fatal("a snapshot holding a syntax node saved")
	}
	var nilCache *findingsCache
	if _, ok := nilCache.take(0, nil); ok || nilCache.save() != nil {
		t.Fatal("a nil cache keeps findings")
	}
	nilCache.store(0, diagnostic.FileResult{})
}
