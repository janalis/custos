package project

import (
	"bytes"
	"context"
	"encoding/binary"
	"os"
	"path/filepath"
	"testing"

	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
	"custos/internal/semantic/flow"
	"custos/internal/semantic/infer"
)

// flowStepsProject is a project of three analysed files and a dependency,
// with the cache directory its runs share.
type flowStepsProject struct {
	root, dir string
	files     []string
	opt       syntax.Options
}

func newFlowStepsProject(t *testing.T) *flowStepsProject {
	t.Helper()
	p := &flowStepsProject{root: t.TempDir(), dir: t.TempDir(), opt: syntax.Options{Version: phpversion.PHP84}}
	for name, src := range map[string]string{
		"a.php": "<?php function a($x) { return b($x); }",
		"b.php": "<?php function b($x) { return \\Lib\\run($x); }",
		"c.php": "<?php a($_GET['x']);",
	} {
		path := filepath.Join(p.root, "src", name)
		writeFile(t, path, src)
		p.files = append(p.files, path)
	}
	lib := filepath.Join(p.root, "vendor", "lib", "run.php")
	writeFile(t, lib, "<?php\nnamespace Lib;\nfunction run($c) { system($c); }\n")
	backdate(t, lib)
	return p
}

// run indexes and summarizes the project as prepare does, with srcs in
// place of the sources read where given, and returns the steps it read
// and the build it made, saved.
func (p *flowStepsProject) run(t *testing.T, srcs map[string][]byte) (*flowSteps, *FlowBuild) {
	t.Helper()
	c := OpenSymbolCache(p.dir, p.root, p.opt, infer.Options{})
	c.fingerprint(len(p.files))
	ix, read, _, _ := buildIndex(IndexSources(p.root, p.files, nil, nil), len(p.files), p.opt, infer.Options{}, c, false)
	for i, path := range p.files {
		if src, ok := srcs[path]; ok {
			read[i] = src
		}
	}
	options := WithVendorFlow(p.root, ix, p.opt, flow.Options{})
	s := c.flowSteps(p.root, p.files, read, options)
	if s == nil {
		t.Fatal("no steps")
	}
	build := buildFlow(context.Background(), p.files, read, p.opt, ix, options, nil, s.earlier, true, nil)
	if err := s.save(build); err != nil {
		t.Fatal(err)
	}
	return s, build
}

// priors lists the files the steps give priors for.
func (s *flowSteps) priors() []string {
	if s.earlier == nil {
		return nil
	}
	var out []string
	for _, path := range s.files {
		if _, ok := s.earlier.at[path]; ok {
			out = append(out, filepath.Base(path))
		}
	}
	return out
}

func TestFlowStepsCache(t *testing.T) {
	p := newFlowStepsProject(t)
	s, cold := p.run(t, nil)
	if s.earlier != nil || cold.made.Load() == 0 {
		t.Fatal("a first run takes no steps")
	}
	s, warm := p.run(t, nil)
	if got := s.priors(); len(got) != 3 || !s.current || warm.made.Load() != 0 {
		t.Fatalf("an unchanged run takes every step: %v, made %d", got, warm.made.Load())
	}
	// Its steps are those read back: it does not write them again.
	if err := os.Remove(s.path); err != nil {
		t.Fatal(err)
	}
	if err := s.save(warm); err != nil || fileExists(s.path) {
		t.Fatal("unchanged steps written again")
	}
	p.run(t, nil)
	// An edited source takes none of its steps; one unread neither.
	edited := []byte("<?php function a($x) { return b($x); }\n// edit")
	if s, _ = p.run(t, map[string][]byte{p.files[0]: edited}); s.current || len(s.priors()) != 2 {
		t.Fatalf("edited: %v", s.priors())
	}
	if s, _ = p.run(t, map[string][]byte{p.files[0]: edited, p.files[2]: nil}); s.current || len(s.priors()) != 2 {
		t.Fatalf("unread: %v", s.priors())
	}
	// Declarations that change take no step at all.
	writeFile(t, p.files[0], "<?php function a($x) { return b($x); } function more() {}")
	if s, _ = p.run(t, nil); s.earlier != nil {
		t.Fatalf("declarations changed: %v", s.priors())
	}
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// A dependency file modified within the racy window, and a run without
// fingerprints or without a cache, take and save no step.
func TestFlowStepsRefused(t *testing.T) {
	p := newFlowStepsProject(t)
	p.run(t, nil)
	writeFile(t, filepath.Join(p.root, "vendor", "lib", "run.php"), "<?php\nnamespace Lib;\nfunction run($c) { return $c; }\n")
	c := OpenSymbolCache(p.dir, p.root, p.opt, infer.Options{})
	c.fingerprint(len(p.files))
	ix, read, _, _ := buildIndex(IndexSources(p.root, p.files, nil, nil), len(p.files), p.opt, infer.Options{}, c, false)
	if s := c.flowSteps(p.root, p.files, read, WithVendorFlow(p.root, ix, p.opt, flow.Options{})); s != nil {
		t.Fatal("racy dependency")
	}
	if (*SymbolCache)(nil).flowSteps(p.root, p.files, read, flow.Options{}) != nil {
		t.Fatal("no cache")
	}
	c = OpenSymbolCache(p.dir, p.root, p.opt, infer.Options{})
	if c.flowSteps(p.root, p.files, read, flow.Options{}) != nil {
		t.Fatal("no fingerprints")
	}
	var none *flowSteps
	if none.save(nil) != nil {
		t.Fatal("saving no steps")
	}
}

// Steps saved otherwise, or that do not decode, are not taken.
func TestFlowStepsDecode(t *testing.T) {
	p := newFlowStepsProject(t)
	s, _ := p.run(t, nil)
	saved, err := os.ReadFile(s.path)
	if err != nil {
		t.Fatal(err)
	}
	srcs := make([][]byte, len(p.files))
	for i, path := range p.files {
		srcs[i], _ = os.ReadFile(path)
	}
	if earlier, current := s.decode(saved, srcs, flow.Options{}); earlier == nil || !current {
		t.Fatal("saved steps")
	}
	body := saved[len(flowCacheMagic):]
	_, n := binary.Uvarint(body)
	keyed := func(key string, rest []byte) []byte {
		out := binary.AppendUvarint([]byte(flowCacheMagic), uint64(len(key)))
		return append(append(out, key...), rest...)
	}
	encoded := body[n+len(s.key):]
	// The steps of a project of other files.
	e := flow.NewEncoder()
	e.Writer().Len(1, false)
	other := e.Bytes()
	e = flow.NewEncoder()
	e.Writer().Len(len(p.files), false)
	e.Writer().String("elsewhere.php")
	misnamed := e.Bytes()
	for name, b := range map[string][]byte{
		"magic":    body,
		"length":   append([]byte(flowCacheMagic), 0xff),
		"key":      keyed("other", encoded),
		"tables":   keyed(s.key, []byte{0x80}),
		"files":    keyed(s.key, other),
		"paths":    keyed(s.key, misnamed),
		"trailing": append(bytes.Clone(saved), 0),
	} {
		if earlier, _ := s.decode(b, srcs, flow.Options{}); earlier != nil {
			t.Errorf("%s: steps taken", name)
		}
	}
}

// A file whose steps cannot be encoded is saved without them, and
// summarized anew.
func TestFlowStepsUnencodable(t *testing.T) {
	p := newFlowStepsProject(t)
	s, build := p.run(t, nil)
	step := build.steps[0][0]
	// An overlay is no basis: the encoder refuses it.
	build.steps[0] = append(build.steps[0], flowStep{step.result, flow.NewSnapshot(step.result).WithFile(step.result)})
	s.current = false
	if err := s.save(build); err != nil {
		t.Fatal(err)
	}
	if s, _ = p.run(t, nil); len(s.priors()) != 2 {
		t.Fatalf("unencodable steps taken: %v", s.priors())
	}
}
