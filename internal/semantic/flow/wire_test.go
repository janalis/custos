package flow

import (
	"errors"
	"reflect"
	"slices"
	"testing"

	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
	"custos/internal/semantic/index"
	"custos/internal/semantic/stubs"
	"custos/internal/semantic/types"
	"custos/internal/semantic/wire"
	"custos/internal/semantic/wire/wiretest"
)

// wireSkipped names the fields the codec does not write: options, which a
// decoder supplies; what makes a snapshot an overlay or a full collection,
// which no Basis holds; and syntax nodes, which no summary keeps (encoding
// one fails).
var wireSkipped = map[string]bool{
	"flow.File.options":     true,
	"flow.Snapshot.options": true,
	"flow.Snapshot.files":   true,
	"flow.Snapshot.base":    true,
	"flow.Snapshot.own":     true,
	"flow.Value.Expr":       true,
	"flow.Value.tested":     true,
	"flow.Value.unless":     true,
	"flow.Origin.Expr":      true,
}

// The codec writes every field: a field added to a type a file or a basis
// reaches fails this test until the codec writes it (or wireSkipped names
// it) and the list below names it.
func TestWireFields(t *testing.T) {
	want := map[string][]string{
		"flow.File":             {"options", "Path", "Summaries", "consulted", "held", "served", "callables"},
		"flow.Origin":           {"Expr", "NonFalse", "NonNull"},
		"flow.Sink":             {"Name", "Context", "Argument", "Path", "Span"},
		"flow.Snapshot":         {"options", "files", "summaries", "shared", "base", "own"},
		"flow.Source":           {"Kind", "Path", "Span", "Parameter", "ValuesOnly", "Keys", "htmlDecoded"},
		"flow.Summary":          {"pure", "defaults", "forward", "argumentSensitive", "returned", "resource", "effects", "dispatch", "Return", "Sinks", "Generator", "Never", "ByRef", "Fixed", "Quiet", "Spares"},
		"flow.Value":            {"record", "literal", "returned", "targets", "interval", "Sources", "Expr", "elements", "origins", "tested", "unless", "Length", "Identity", "storedIdentity", "Invalidated", "Safe", "Complete", "NonFalse", "NonNull", "LengthKnown", "headerMask", "upload", "scalars", "matched", "unlessFalse", "unlessNull", "skippedFalse", "skippedNull", "replaced", "entry", "plain", "deep", "truth"},
		"flow.callableSet":      {"items"},
		"flow.callableValue":    {"contract", "outer"},
		"flow.constantScalar":   {"k", "i", "f", "s", "b"},
		"flow.elementSet":       {"hasResources", "whole", "optional", "items", "order", "depth", "width"},
		"flow.forwardReturn":    {"parameter", "argument", "result", "weakReturn"},
		"flow.intInterval":      {"lo", "hi"},
		"flow.originSet":        {"items", "full", "compressed", "other"},
		"flow.pureContract":     {"params", "expr", "calls", "result", "weakReturn", "outer"},
		"flow.pureExpr":         {"kind", "weakArguments", "projects", "appendKey", "stopGuard", "prepared", "destruct", "slot", "op", "literal", "left", "right", "other", "call", "args", "cost", "depth"},
		"flow.pureField":        {"name", "key"},
		"flow.pureParam":        {"name", "typ", "initial", "optional", "nullable"},
		"flow.resourceEffect":   {"onTrue", "parameter", "operation", "checked"},
		"flow.resourceReturn":   {"token", "fresh", "parameter", "operation", "checked"},
		"flow.returnedCallable": {"pure", "outerParams", "forward", "argumentSensitive", "params", "defaults", "result", "sinks"},
		"flow.sinkClass":        {"Final", "Kind", "Modifiers", "Extends", "Traits", "Methods"},
		"flow.sinkMethod":       {"Count", "Modifiers", "Parameters"},
		"index.Avail":           {"From", "To"},
		"index.Param":           {"Name", "Type", "DocType", "DocValues", "Optional", "ByRef", "Variadic", "Promoted", "Default", "Out", "Gen", "Avail"},
		"syntax.Span":           {"Start", "End"},
	}
	// Types encode their own fields (see types.TestWireFields).
	stop := func(t reflect.Type) bool { return t == reflect.TypeFor[types.Type]() }
	got := wiretest.Fields(reflect.TypeFor[File](), wireSkipped, stop)
	for name, fields := range wiretest.Fields(reflect.TypeFor[Snapshot](), wireSkipped, stop) {
		got[name] = fields
	}
	if !reflect.DeepEqual(got, want) {
		for name, fields := range got {
			if !slices.Equal(fields, want[name]) {
				t.Errorf("%s: fields %v, the codec writes %v", name, fields, want[name])
			}
		}
		t.Fatal("the codec does not write every field")
	}
}

// wireRoundTrip encodes files and snapshots in turn and decodes them again
// with options.
func wireRoundTrip(t testing.TB, files []*File, snapshots []*Snapshot, options Options) ([]*File, []*Snapshot) {
	e := NewEncoder()
	for i := range files {
		e.File(files[i])
		e.Snapshot(snapshots[i])
	}
	if e.Err() != nil {
		t.Fatal(e.Err())
	}
	d, err := NewDecoder(e.Bytes(), options)
	if err != nil {
		t.Fatal(err)
	}
	outFiles, outSnapshots := make([]*File, len(files)), make([]*Snapshot, len(files))
	for i := range files {
		outFiles[i], outSnapshots[i] = d.File(), d.Snapshot()
	}
	if !d.Reader().Done() {
		t.Fatal("not read whole", d.Reader().Err())
	}
	return outFiles, outSnapshots
}

// Every field of a file and of a basis survives encoding, nil slices and
// maps staying nil and empty ones empty; the decoder supplies options.
func TestWireRoundTrip(t *testing.T) {
	vendor := &Vendor{}
	for depth := range 5 {
		var f File
		var s Snapshot
		wiretest.Fill(&f, depth, wireSkipped)
		wiretest.Fill(&s, depth, wireSkipped)
		files, snapshots := wireRoundTrip(t, []*File{&f}, []*Snapshot{&s}, Options{Assertions: true, Vendor: vendor})
		f.options, s.options = Options{Assertions: true, Vendor: vendor}, Options{Assertions: true, Vendor: vendor}
		if !reflect.DeepEqual(files[0], &f) {
			t.Fatalf("depth %d: the file differs", depth)
		}
		if !reflect.DeepEqual(snapshots[0], &s) {
			t.Fatalf("depth %d: the snapshot differs", depth)
		}
	}
	empty := &File{Summaries: map[string]Summary{"k": {pure: &pureContract{expr: &pureExpr{args: []*pureExpr{nil, {}}}}, defaults: []Value{}, ByRef: map[int]Value{}, Return: Value{Sources: []Source{}, targets: &callableSet{items: []syntax.Expr{}}, elements: &elementSet{items: map[string]Value{}}, origins: &originSet{items: []Origin{}}}}}, consulted: map[string]bool{}, held: []string{}, served: map[string]bool{}, callables: map[string]map[string]uint8{"k": {}}}
	basis := &Snapshot{summaries: map[string]*Summary{"none": nil}, shared: map[string]int{}}
	files, snapshots := wireRoundTrip(t, []*File{empty, nil}, []*Snapshot{basis, nil}, Options{})
	if !reflect.DeepEqual(files, []*File{empty, nil}) || !reflect.DeepEqual(snapshots, []*Snapshot{basis, nil}) {
		t.Fatal("empty and nil values")
	}
}

// A summary is written once however many snapshots hold it, and so is
// another encoded alike.
func TestWireSharesSummaries(t *testing.T) {
	held := &Summary{Generator: true}
	alike := &Summary{Generator: true}
	e := NewEncoder()
	for _, s := range []*Snapshot{{summaries: map[string]*Summary{"a": held}}, {summaries: map[string]*Summary{"a": held}}, {summaries: map[string]*Summary{"b": alike}}} {
		e.Snapshot(s)
	}
	e.File(&File{Summaries: map[string]Summary{"c": *held}})
	if e.count != 1 {
		t.Fatalf("%d summaries written", e.count)
	}
}

// A value that holds a syntax node, or a snapshot that is not a basis,
// fails the encoder; a rollback drops what the failed entry wrote.
func TestWireRefuses(t *testing.T) {
	node := &syntax.Binary{}
	for name, v := range map[string]Value{
		"expression":   {Expr: node},
		"tested":       {tested: node},
		"unless":       {unless: node},
		"target":       {targets: &callableSet{items: []syntax.Expr{node}}},
		"origin":       {origins: &originSet{items: []Origin{{Expr: node}}}},
		"nested":       {elements: &elementSet{items: map[string]Value{"k": {Expr: node}}}},
		"nested items": {returned: &callableValue{outer: []Value{{Expr: node}}}},
	} {
		e := NewEncoder()
		mark := e.Mark()
		e.File(&File{Summaries: map[string]Summary{"f": {Return: v}}})
		if e.Err() == nil {
			t.Fatalf("%s: encoded", name)
		}
		e.Rollback(mark)
		if e.Err() != nil || e.Mark() != mark {
			t.Fatalf("%s: rollback", name)
		}
		// Through a snapshot as well.
		e.Snapshot(&Snapshot{summaries: map[string]*Summary{"f": {Return: v}}})
		if e.Err() == nil {
			t.Fatalf("%s: encoded in a snapshot", name)
		}
	}
	full := NewSnapshot(&File{Path: "a", Summaries: map[string]Summary{"f": {}}})
	for name, s := range map[string]*Snapshot{"collection": full, "overlay": full.WithFile(&File{Path: "b"})} {
		e := NewEncoder()
		if e.Snapshot(s); e.Err() == nil {
			t.Fatalf("%s: encoded", name)
		}
	}
	// Too deep.
	x := &pureExpr{}
	for range wire.MaxDepth {
		x = &pureExpr{left: x}
	}
	e := NewEncoder()
	if e.File(&File{Summaries: map[string]Summary{"f": {pure: &pureContract{expr: x}}}}); e.Err() == nil {
		t.Fatal("too deep")
	}
	v := Value{}
	for range wire.MaxDepth {
		v = Value{returned: &callableValue{contract: &returnedCallable{result: v}}}
	}
	e = NewEncoder()
	if e.File(&File{Summaries: map[string]Summary{"f": {Return: v}}}); e.Err() == nil {
		t.Fatal("too deep value")
	}
	e = NewEncoder()
	if e.File(&File{Summaries: map[string]Summary{"f": {returned: v.returned.contract}}}); e.Err() == nil {
		t.Fatal("too deep callable")
	}
	c := &pureContract{}
	for range wire.MaxDepth {
		c = &pureContract{expr: &pureExpr{call: c}}
	}
	e = NewEncoder()
	if e.File(&File{Summaries: map[string]Summary{"f": {pure: c}}}); e.Err() == nil {
		t.Fatal("too deep contract")
	}
}

// A corrupt encoding is refused, or decodes to a failed reader.
func TestWireCorrupt(t *testing.T) {
	e := NewEncoder()
	e.File(&File{Path: "a", Summaries: map[string]Summary{"f": {Generator: true}}})
	good := e.Bytes()
	if _, err := NewDecoder(good, Options{}); err != nil {
		t.Fatal(err)
	}
	strs := wire.NewStrings().Append(nil)
	for name, b := range map[string][]byte{
		"strings":   {0x80},
		"header":    append(slices.Clone(strs), 0x80),
		"size":      append(slices.Clone(strs), 0, 9),
		"count":     append(slices.Clone(strs), 2, 1, 0),
		"summaries": append(slices.Clone(strs), 1, 2, 0, 0xff),
	} {
		if _, err := NewDecoder(b, Options{}); !errors.Is(err, wire.ErrCorrupt) {
			t.Fatalf("%s: %v", name, err)
		}
	}
	// A file naming a summary the table lacks.
	e = NewEncoder()
	w := e.Writer()
	w.Bool(true)
	w.String("a")
	w.Len(1, false)
	w.String("f")
	w.Uint(7)
	d, err := NewDecoder(e.Bytes(), Options{})
	if err != nil {
		t.Fatal(err)
	}
	if d.File(); d.Reader().Err() == nil {
		t.Fatal("a missing summary decoded")
	}
	// Nesting past the bound, which an encoder within it may still reach
	// at the start of a reader already nested: each chain alternates two
	// decoders, and one more level of nesting moves the failure to the
	// other.
	links := wire.MaxDepth/2 - 4
	var v Value
	for range links {
		v = Value{returned: &callableValue{contract: &returnedCallable{result: v}}}
	}
	var c *pureContract
	for range links {
		c = &pureContract{expr: &pureExpr{call: c}}
	}
	for name, summary := range map[string]Summary{"value": {Return: v}, "contract": {pure: c}} {
		table := wire.NewStrings()
		w := wire.NewWriter(table)
		encodeSummary(w, &summary)
		if w.Err() != nil {
			t.Fatal(w.Err())
		}
		strs, _, _ := wire.ReadStrings(table.Append(nil))
		for nested := 9; nested < 13; nested++ {
			r := wire.NewReader(w.Buf, strs)
			for range nested {
				r.Enter()
			}
			if decodeSummary(r); r.Err() == nil {
				t.Fatalf("%s: too deep decoded", name)
			}
		}
	}
}

// wireSample summarizes a file of a few dozen callables, and returns its
// summaries with their basis.
func wireSample(tb testing.TB) (*File, *Snapshot) {
	src := []byte(`<?php
final class Repo {
	public function find($id) { return $this->query("SELECT * FROM t WHERE id = " . $id); }
	public function query($sql) { return (new \PDO('sqlite::memory:'))->query($sql); }
	public function all(array $ids) { return array_map(fn($id) => $this->find($id), $ids); }
}
function shell($c) { return system($c); }
function wrap($c) { return shell(escapeshellarg($c)); }
function pick(array $a, $k) { return $a[$k] ?? null; }
function count_up(int $n) { $s = 0; for ($i = 0; $i < $n; $i++) { $s += $i; } return $s; }
function fill(&$out, $v) { $out = $v; }
function each_line($text) { foreach (explode("\n", $text) as $line) { yield trim($line); } }
function header_to($url) { header('Location: ' . $url); }
function read_file($p) { return file_get_contents(__DIR__ . '/' . basename($p)); }
`)
	f := syntax.ParseBest("sample.php", src, syntax.Options{Version: phpversion.PHP84})
	ix := index.New(stubs.Index())
	ix.Add(index.Extract(f))
	file := Extract(f, ix, phpversion.PHP84, nil, Options{})
	other := &File{Path: "other.php", Summaries: map[string]Summary{"shell": file.Summaries["shell"]}}
	basis := file.Basis(NewSnapshot(other, file))
	return file, basis
}

func TestWireSample(t *testing.T) {
	file, basis := wireSample(t)
	files, snapshots := wireRoundTrip(t, []*File{file}, []*Snapshot{basis}, Options{})
	if !reflect.DeepEqual(files[0], file) || !reflect.DeepEqual(snapshots[0], basis) {
		t.Fatal("a real file differs")
	}
}

func BenchmarkWireEncode(b *testing.B) {
	file, basis := wireSample(b)
	b.ReportAllocs()
	for b.Loop() {
		e := NewEncoder()
		e.File(file)
		e.Snapshot(basis)
		_ = e.Bytes()
	}
}

func BenchmarkWireDecode(b *testing.B) {
	file, basis := wireSample(b)
	e := NewEncoder()
	e.File(file)
	e.Snapshot(basis)
	encoded := e.Bytes()
	b.SetBytes(int64(len(encoded)))
	b.ReportAllocs()
	for b.Loop() {
		d, _ := NewDecoder(encoded, Options{})
		d.File()
		d.Snapshot()
	}
}
