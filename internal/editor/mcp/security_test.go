package mcp

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	sdk "github.com/modelcontextprotocol/go-sdk/mcp"

	"custos/internal/diagnostic"
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/meta"
	"custos/internal/php/syntax"
	"custos/internal/project"
	"custos/internal/project/config"
)

func TestStartupFailClosed(t *testing.T) {
	s, _, root := fixture(t)
	base := s.options
	for _, options := range []Options{{}, {Workspaces: make([]Workspace, 65)}, {Storage: base.Storage, Workspaces: []Workspace{{Root: root}}}, {Storage: base.Storage, Workspaces: []Workspace{{ID: "x", Root: root}, {ID: "x", Root: root}}}, {Storage: base.Storage, Default: "missing", Workspaces: base.Workspaces}, {Storage: filepath.Join(root, "state"), Workspaces: base.Workspaces}, {Storage: root, Workspaces: base.Workspaces}, {Storage: t.TempDir(), Workspaces: base.Workspaces}} {
		if created, err := New(options, s.rules); err == nil {
			created.Close()
			t.Fatal("unsafe options accepted", options)
		}
	}
	for _, fault := range []string{"abs-root", "canonical-root", "open-root", "cache", "abs-storage", "mkdir", "stat", "canonical-storage", "open-storage", "load"} {
		t.Run(fault, func(t *testing.T) {
			saved := startupIO
			defer func() { startupIO = saved }()
			sentinel := errors.New("filesystem fault")
			var opened []*os.Root
			startupIO.openRoot = func(p string) (*os.Root, error) {
				r, err := saved.openRoot(p)
				if r != nil {
					opened = append(opened, r)
				}
				return r, err
			}
			options := base
			switch fault {
			case "abs-root":
				startupIO.abs = func(string) (string, error) { return "", sentinel }
			case "canonical-root":
				startupIO.canonical = func(string) (string, error) { return "", sentinel }
			case "open-root":
				startupIO.openRoot = func(string) (*os.Root, error) { return nil, sentinel }
			case "cache":
				options.Storage = ""
				startupIO.cache = func() (string, error) { return "", sentinel }
			case "abs-storage":
				n := 0
				startupIO.abs = func(p string) (string, error) {
					n++
					if n == 2 {
						return "", sentinel
					}
					return saved.abs(p)
				}
			case "mkdir":
				startupIO.mkdir = func(string, os.FileMode) error { return sentinel }
			case "stat":
				startupIO.stat = func(string) (os.FileInfo, error) { return nil, sentinel }
			case "canonical-storage":
				n := 0
				startupIO.canonical = func(p string) (string, error) {
					n++
					if n == 2 {
						return "", sentinel
					}
					return saved.canonical(p)
				}
			case "open-storage":
				n := 0
				original := startupIO.openRoot
				startupIO.openRoot = func(p string) (*os.Root, error) {
					n++
					if n == 2 {
						return nil, sentinel
					}
					return original(p)
				}
			case "load":
				n := 0
				original := startupIO.openRoot
				startupIO.openRoot = func(p string) (*os.Root, error) {
					n++
					r, err := original(p)
					if n == 2 && r != nil {
						_ = r.Close()
					}
					return r, err
				}
			}
			if created, err := New(options, s.rules); err == nil {
				created.Close()
				t.Fatal("filesystem fault accepted")
			}
			for _, r := range opened {
				if _, err := r.Stat("."); err == nil {
					t.Fatal("startup leaked an open root")
				}
			}
		})
	}
	t.Setenv("HOME", t.TempDir())
	t.Setenv("XDG_CACHE_HOME", t.TempDir())
	options := base
	options.Storage = ""
	created, err := New(options, s.rules)
	if err != nil {
		t.Fatal(err)
	}
	created.Close()
	if validID("short") || validID(strings.Repeat("z", 48)) || !validID(randomID()) {
		t.Fatal("ID validation")
	}
	for _, data := range []string{`{"unknown":true}`, `{} {}`, `{} garbage`} {
		if _, err := ParseOptions([]byte(data)); err == nil {
			t.Fatal(data)
		}
	}
	if _, err := ParseOptions([]byte(`{}`)); err != nil {
		t.Fatal(err)
	}
}

func TestCorruptedStorageAndRetention(t *testing.T) {
	s, ctx, _ := fixture(t)
	cpID := invoke(ctx, t, s, "analyze", Input{})["checkpointId"].(string)
	var base checkpoint
	data, err := s.storage.ReadFile(cpID + ".json")
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(data, &base); err != nil {
		t.Fatal(err)
	}
	write := func(name string, b []byte) {
		t.Helper()
		if err := s.storage.WriteFile(name, b, 0o600); err != nil {
			t.Fatal(err)
		}
	}
	write("unrelated.tmp", []byte("ignored"))
	write("bad-id.json", []byte("{}"))
	write(randomID()+".json", []byte("broken"))
	for _, mutate := range []func(*checkpoint){func(c *checkpoint) { c.Identity = "unauthorized" }, func(c *checkpoint) { c.Root = "wrong" }, func(c *checkpoint) { c.Hashes = map[string]string{"../escape": "hash"} }, func(c *checkpoint) { c.Paths = []string{"../escape"} }, func(c *checkpoint) {
		c.Findings = []evidence{{Item: diagnostic.Item{Path: "../escape"}, Start: 9, End: 1}}
	}, func(c *checkpoint) { c.Created = time.Now().Add(time.Hour) }, func(c *checkpoint) { c.Buffer = true }} {
		copy := base
		copy.ID = randomID()
		mutate(&copy)
		b, _ := json.Marshal(copy)
		write(copy.ID+".json", b)
	}
	huge, err := s.storage.OpenFile(randomID()+".json", os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		t.Fatal(err)
	}
	if err := huge.Truncate((64 << 20) + 1); err != nil {
		t.Fatal(err)
	}
	_ = huge.Close()
	if err := os.Symlink(filepath.Join(t.TempDir(), "outside"), filepath.Join(s.storage.Name(), randomID()+".json")); err != nil {
		t.Fatal(err)
	}
	s.checkpoints = map[string]*checkpoint{}
	if err := s.load(); err != nil {
		t.Fatal(err)
	}
	if len(s.checkpoints) != 1 {
		t.Fatal("corrupt checkpoints loaded", len(s.checkpoints))
	}
	for range 25 {
		copy := base
		copy.ID = randomID()
		copy.Created = time.Now()
		s.checkpoints[copy.ID] = &copy
		s.snapshots[copy.ID] = &project.Snapshot{Sources: map[string][]byte{"a": []byte("<?php")}}
	}
	expired := base
	expired.ID = randomID()
	expired.Created = time.Now().Add(-8 * 24 * time.Hour)
	s.checkpoints[expired.ID] = &expired
	s.prune()
	if len(s.checkpoints) != 20 || len(s.snapshots) > 2 {
		t.Fatal("retention budgets", len(s.checkpoints), len(s.snapshots))
	}
	// Oversized evidence and snapshots are evicted rather than retained.
	copy := base
	copy.ID = randomID()
	copy.Created = time.Now()
	copy.Identity = "bob"
	copy.Findings = []evidence{{Item: diagnostic.Item{Message: strings.Repeat("x", 65<<20)}}}
	s.checkpoints[copy.ID] = &copy
	s.prune()
	if s.checkpoints[copy.ID] != nil {
		t.Fatal("global storage budget")
	}
	buffer := "<?php $x=!!!$a;"
	buf := invoke(ctx, t, s, "analyze", Input{Buffer: &buffer, Path: "unsaved.php"})["checkpointId"].(string)
	delete(s.snapshots, buf)
	rejected(ctx, t, s, "type", Input{CheckpointID: buf, Path: "unsaved.php", Offset: 10})
	if err := s.storage.Close(); err != nil {
		t.Fatal(err)
	}
	if err := s.load(); err == nil {
		t.Fatal("closed storage accepted")
	}
}

func TestBaselineUnicodeAndAmbiguity(t *testing.T) {
	s, ctx, root := fixture(t)
	a := invoke(ctx, t, s, "analyze", Input{})
	cp := s.checkpoints[a["checkpointId"].(string)]
	ev := cp.Findings[0]
	b, _ := json.Marshal(map[string]any{"entries": []any{map[string]any{"path": "a.php", "rule": ev.Rule, "message": ev.Message, "line": "$x = !!!$token; function sample(): int { return 1; }", "count": 1}}})
	writeFile(t, filepath.Join(root, "baseline.json"), string(b))
	writeFile(t, filepath.Join(root, "custos.json"), `{"php":"8.2","baseline":"baseline.json"}`)
	a = invoke(ctx, t, s, "analyze", Input{})
	if a["findings"] != 0 {
		t.Fatal("baseline not honored", a)
	}
	writeFile(t, filepath.Join(root, "baseline.json"), "invalid")
	rejected(ctx, t, s, "analyze", Input{})
	writeFile(t, filepath.Join(root, "custos.json"), `{"php":"8.2"}`)
	writeFile(t, filepath.Join(root, "a.php"), "<?php $é = !!!$token; class A {public function go(): int{return 1;}} const X = 1; function duplicate() {}")
	writeFile(t, filepath.Join(root, "b.php"), "<?php function duplicate() {}")
	if err := os.Mkdir(filepath.Join(root, "vendor"), 0o700); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(root, "vendor", "v.php"), "<?php function vendorDecl() {}")
	a = invoke(ctx, t, s, "analyze", Input{})
	id := a["checkpointId"].(string)
	ev = s.checkpoints[id].Findings[0]
	if ev.Column != 12 || ev.Start != 12 {
		t.Fatal("rune/byte position", ev)
	}
	if scoped := invoke(ctx, t, s, "symbols", Input{CheckpointID: id}); scoped["scope"] != "project" || scoped["total"] != 5 {
		t.Fatal(scoped)
	}
	result := invoke(ctx, t, s, "symbols", Input{CheckpointID: id, Scope: "all"})
	ambiguous := 0
	vendor := false
	for _, raw := range result["items"].([]any) {
		item := raw.(Output)
		if item["ambiguous"] == true {
			ambiguous++
		}
		if item["provenance"] == "vendor" {
			vendor = true
		}
	}
	if ambiguous != 2 || !vendor {
		t.Fatal(result)
	}
	invoke(ctx, t, s, "symbols", Input{CheckpointID: id, Path: "a.php", Query: "A"})
	rejected(ctx, t, s, "findings", Input{CheckpointID: id, Rule: "wrong"})
	rejected(ctx, t, s, "findings", Input{CheckpointID: id, Rule: "no-such-rule"})
	if out := invoke(ctx, t, s, "findings", Input{CheckpointID: id, Rule: "unnecessaryassertion"}); out["total"] != 0 {
		t.Fatal(out)
	}
	invoke(ctx, t, s, "rules", Input{Query: "no rule matches this"})
	unresolved := invoke(ctx, t, s, "type", Input{CheckpointID: id, Path: "a.php", Offset: 1})
	if unresolved["resolved"] != false {
		t.Fatal(unresolved)
	}
	buffer := "<?php function broken( {"
	invalid := invoke(ctx, t, s, "analyze", Input{Buffer: &buffer, Path: "broken.php"})
	if invalid["complete"] != false {
		t.Fatal(invalid)
	}
	snap := &project.Snapshot{Config: s.snapshots[id].Config, Sources: map[string][]byte{}, Results: []diagnostic.FileResult{{Path: "missing.php", Err: errors.New("unreadable")}}}
	tmp := &checkpoint{Complete: true}
	if err := collect(tmp, snap); err != nil || tmp.Complete {
		t.Fatal(tmp, err)
	}
}

func TestCommitFaultsAndPathSwaps(t *testing.T) {
	s, ctx, root := fixture(t)
	a := invoke(ctx, t, s, "analyze", Input{})
	id := a["checkpointId"].(string)
	cp := s.checkpoints[id]
	saved := openTemporary
	defer func() { openTemporary = saved; beforeCommit = nil }()
	openTemporary = func(*os.Root, string, os.FileMode) (*os.File, error) { return nil, errors.New("disk unavailable") }
	if err := s.persist(cp); err == nil {
		t.Fatal("persist opened failed")
	}
	rejected(ctx, t, s, "analyze", Input{})
	openTemporary = func(r *os.Root, name string, perm os.FileMode) (*os.File, error) {
		f, err := saved(r, name, perm)
		if f != nil {
			_ = f.Close()
		}
		return f, err
	}
	if err := s.persist(cp); err == nil {
		t.Fatal("closed temporary accepted")
	}
	openTemporary = saved
	p := invoke(ctx, t, s, "prepare_fixes", Input{CheckpointID: id, Iterative: true})
	planID := p["planId"].(string)
	openTemporary = func(*os.Root, string, os.FileMode) (*os.File, error) { return nil, errors.New("disk unavailable") }
	result := invoke(ctx, t, s, "apply_fixes", Input{PlanID: planID})
	if result["partialFailure"] != true {
		t.Fatal(result)
	}
	rejected(ctx, t, s, "apply_fixes", Input{PlanID: planID})
	openTemporary = saved
	a = invoke(ctx, t, s, "analyze", Input{})
	id = a["checkpointId"].(string)
	snap := s.snapshots[id]
	c := change{Path: "a.php", Source: snap.Sources["a.php"], Output: []byte("<?php $x=0;"), Perm: 0o600}
	secret := filepath.Join(t.TempDir(), "outside.php")
	writeFile(t, secret, "<?php $secret = 42;")
	beforeCommit = func(parent *os.Root, name string) {
		_ = parent.Remove(name)
		if err := os.Symlink(secret, filepath.Join(root, name)); err != nil {
			t.Fatal(err)
		}
	}
	if err := replace(s.workspaces["default"], c); err == nil {
		t.Fatal("symlink swap applied")
	}
	b, err := os.ReadFile(secret)
	if err != nil || string(b) != "<?php $secret = 42;" {
		t.Fatal("outside file changed", err)
	}
	beforeCommit = nil
	if err := replace(s.workspaces["default"], change{Path: "missing/a.php"}); err == nil {
		t.Fatal("missing parent")
	}
	if err := replace(s.workspaces["default"], change{Path: "missing.php"}); err == nil {
		t.Fatal("missing file")
	}
	if err := os.Remove(filepath.Join(root, "a.php")); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(root, "a.php"), string(c.Source))
	if err := os.Chmod(filepath.Join(root, "a.php"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := replace(s.workspaces["default"], c); err == nil {
		t.Fatal("changed permissions accepted")
	}
	// The persisted target becoming a directory must leave no temporary file.
	cp = s.checkpoints[id]
	_ = s.storage.Remove(cp.ID + ".json")
	if err := s.storage.Mkdir(cp.ID+".json", 0o700); err != nil {
		t.Fatal(err)
	}
	if err := s.persist(cp); err == nil {
		t.Fatal("persist directory target")
	}
}

func TestPaginationTamperingAndBusy(t *testing.T) {
	s, ctx, _ := fixture(t)
	a := invoke(ctx, t, s, "analyze", Input{})
	id := a["checkpointId"].(string)
	rejected(ctx, t, s, "unknown", Input{CheckpointID: id})
	// A held operation slot makes callers wait until their context ends;
	// beyond the queue bound they are turned away at once.
	s.active <- struct{}{}
	waiting, cancel := context.WithTimeout(ctx, 10*time.Millisecond)
	rejected(waiting, t, s, "workspace", Input{})
	cancel()
	s.queued.Store(maxQueued)
	rejected(ctx, t, s, "workspace", Input{})
	s.queued.Store(0)
	<-s.active
	pageInput := Input{}
	filters, _ := json.Marshal([]any{identity(ctx), "scope", pageInput})
	binding := project.Fingerprint(filters)
	for _, offset := range []string{"-1", "100", "bad"} {
		cursor := base64.RawURLEncoding.EncodeToString([]byte(binding + ":" + offset))
		if _, err := page(ctx, "scope", Input{Cursor: cursor}, []any{1}); err == nil {
			t.Fatal(offset)
		}
	}
	for _, action := range []string{"delete", "resume"} {
		rejected(ctx, t, s, "checkpoints", Input{CheckpointID: "missing", Action: action})
	}
	rejected(ctx, t, s, "prepare_fixes", Input{CheckpointID: id, PlanID: "missing"})
	for i := range 20 {
		s.plans[fmt.Sprint(i)] = &plan{Expires: time.Now().Add(time.Hour)}
	}
	rejected(ctx, t, s, "prepare_fixes", Input{CheckpointID: id, Iterative: true})
}

func TestWorkspaceIsolationAndInvalidation(t *testing.T) {
	s, ctx, root := fixture(t)
	a := invoke(ctx, t, s, "analyze", Input{})
	id := a["checkpointId"].(string)
	extra, err := os.OpenRoot(root)
	if err != nil {
		t.Fatal(err)
	}
	s.workspaces["extra"] = &workspace{Workspace: Workspace{ID: "extra", Root: s.workspaces["default"].Root, Readers: []string{"local"}}, root: extra}
	hidden, err := os.OpenRoot(root)
	if err != nil {
		t.Fatal(err)
	}
	s.workspaces["hidden"] = &workspace{Workspace: Workspace{ID: "hidden", Root: s.workspaces["default"].Root, Readers: []string{"alice"}}, root: hidden}
	result := invoke(ctx, t, s, "workspace", Input{})
	if result["total"] != 2 {
		t.Fatal(result)
	}
	rejected(ctx, t, s, "findings", Input{WorkspaceID: "extra", CheckpointID: id})
	writeFile(t, filepath.Join(root, "custos.json"), "invalid")
	rejected(ctx, t, s, "workspace", Input{})
	if _, err := s.ruleInfo(ctx, s.workspaces["default"], Input{}); err == nil {
		t.Fatal("invalid rule settings")
	}
	rejected(ctx, t, s, "findings", Input{CheckpointID: id})
	writeFile(t, filepath.Join(root, "custos.json"), `{"php":"8.2"}`)
	a = invoke(ctx, t, s, "analyze", Input{})
	id = a["checkpointId"].(string)
	delete(s.snapshots, id)
	saved := s.rules
	s.rules = []analysis.Rule{missingRule{}}
	rejected(ctx, t, s, "symbols", Input{CheckpointID: id})
	s.rules = saved
}

type missingRule struct{}

func (missingRule) ID() string                           { return "InvalidRule" }
func (missingRule) Kinds() []syntax.NodeKind             { return nil }
func (missingRule) Check(*analysis.Context, syntax.Node) {}

type editRule struct {
	edit   string
	before func()
}

func (editRule) ID() string               { return "NestedNotOperators" }
func (editRule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KUnary} }
func (r editRule) Check(ctx *analysis.Context, n syntax.Node) {
	if r.before != nil {
		r.before()
	}
	ctx.Report(n.Span(), "Replace expression.", diagnostic.Fix{Title: "Replace", Edits: func() []diagnostic.TextEdit { return []diagnostic.TextEdit{{Span: n.Span(), NewText: r.edit}} }})
}

func TestPreparedFixValidationAndBudgets(t *testing.T) {
	s, ctx, root := fixture(t)
	a := invoke(ctx, t, s, "analyze", Input{})
	id := a["checkpointId"].(string)
	cp := s.checkpoints[id]
	snap := s.snapshots[id]
	w := s.workspaces["default"]
	if _, err := s.prepare(ctx, w, cp, snap, Input{Choices: []Choice{{FindingID: "0"}, {FindingID: "0"}}}); err == nil {
		t.Fatal("overlapping selected fixes accepted")
	}
	altered := *cp
	altered.Findings = append([]evidence{}, cp.Findings...)
	altered.Findings[0].Rule = "unreproducible"
	if _, err := s.prepare(ctx, w, &altered, snap, Input{Choices: []Choice{{FindingID: "0"}}}); err == nil {
		t.Fatal("unreproducible finding")
	}
	invoke(ctx, t, s, "prepare_fixes", Input{CheckpointID: id, Iterative: true, Rule: "unnecessaryassertion", Path: "missing.php"})
	invoke(ctx, t, s, "prepare_fixes", Input{CheckpointID: id, Iterative: true, Rule: "UnnecessaryAssertion"})
	rejected(ctx, t, s, "prepare_fixes", Input{CheckpointID: id, Iterative: true, Rule: "wrong"})
	cancelled, cancel := context.WithCancel(ctx)
	cancel()
	if _, err := s.prepare(cancelled, w, cp, snap, Input{Iterative: true}); err == nil {
		t.Fatal("cancelled preparation")
	}
	if _, err := s.prepare(cancelled, w, cp, snap, Input{Choices: []Choice{{FindingID: "0"}}}); err == nil {
		t.Fatal("cancelled selected preparation")
	}
	engine, err := analysis.NewEngine([]analysis.Rule{editRule{edit: "("}}, snap.Config.Analysis())
	if err != nil {
		t.Fatal(err)
	}
	original := snap.Engine
	snap.Engine = engine
	if _, err := s.prepare(ctx, w, cp, snap, Input{Iterative: true}); err == nil {
		t.Fatal("unparsable fix")
	}
	snap.Engine = original
	// A file vanishing or becoming a link before permission capture fails closed.
	if err := os.Remove(filepath.Join(root, "a.php")); err != nil {
		t.Fatal(err)
	}
	if _, err := s.prepare(ctx, w, cp, snap, Input{Iterative: true}); err == nil {
		t.Fatal("missing fix input")
	}
	if err := os.Symlink(filepath.Join(t.TempDir(), "outside"), filepath.Join(root, "a.php")); err != nil {
		t.Fatal(err)
	}
	if _, err := s.prepare(ctx, w, cp, snap, Input{Iterative: true}); err == nil {
		t.Fatal("linked fix input")
	}
	if err := os.Remove(filepath.Join(root, "a.php")); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(root, "a.php"), string(snap.Sources["a.php"]))
	s.plans["huge"] = &plan{Changes: []change{{Source: make([]byte, 64<<20)}}, Expires: time.Now().Add(time.Hour)}
	if _, err := s.prepare(ctx, w, cp, snap, Input{Iterative: true}); err == nil {
		t.Fatal("global plan bytes")
	}
	delete(s.plans, "huge")
	// Preview text stays pageable for large replacements and large diffs.
	text := strings.Repeat("a", 10000)
	p := &plan{ID: randomID(), Expires: time.Now().Add(time.Hour), Changes: []change{{Path: "a.php", Source: []byte(text), Output: []byte(text + "x"), Edits: []diagnostic.TextEdit{{Span: syntax.Span{End: 10000}, NewText: text}}}}}
	result, err := s.planPage(ctx, p, Input{})
	if err != nil || result["total"].(int) < 3 {
		t.Fatal(result, err)
	}
	p.Changes[0].Source = []byte(strings.Repeat("a\n", 400))
	p.Changes[0].Output = []byte(strings.Repeat("b\n", 400))
	if _, err := s.planPage(ctx, p, Input{}); err != nil {
		t.Fatal(err)
	}
}

func TestPlanApplyErrorsAndAnalysisFailure(t *testing.T) {
	s, ctx, root := fixture(t)
	a := invoke(ctx, t, s, "analyze", Input{})
	id := a["checkpointId"].(string)
	p := invoke(ctx, t, s, "prepare_fixes", Input{CheckpointID: id, Iterative: true})
	planID := p["planId"].(string)
	existing := s.plans[planID]
	existing.Checkpoint = "missing"
	rejected(ctx, t, s, "apply_fixes", Input{PlanID: planID})
	existing.Checkpoint = id
	writeFile(t, filepath.Join(root, "a.php"), "<?php $x=0;")
	rejected(ctx, t, s, "apply_fixes", Input{PlanID: planID})
	writeFile(t, filepath.Join(root, "a.php"), "<?php $x=!!!$a;")
	a = invoke(ctx, t, s, "analyze", Input{})
	id = a["checkpointId"].(string)
	p = invoke(ctx, t, s, "prepare_fixes", Input{CheckpointID: id, Iterative: true})
	planID = p["planId"].(string)
	beforeCommit = func(*os.Root, string) { writeFile(t, filepath.Join(root, "custos.json"), "invalid") }
	defer func() { beforeCommit = nil }()
	result := invoke(ctx, t, s, "apply_fixes", Input{PlanID: planID})
	if result["analysisError"] == nil {
		t.Fatal("post-commit analysis failure missing", result)
	}
}

type validationContext struct {
	context.Context
	steps, stop int
}

func (c *validationContext) Err() error {
	c.steps++
	if c.stop > 0 && c.steps >= c.stop {
		return context.Canceled
	}
	return c.Context.Err()
}

func TestCheckpointBudgetsAndCancelledPublication(t *testing.T) {
	s, ctx, root := fixture(t)
	a := invoke(ctx, t, s, "analyze", Input{})
	cp := s.checkpoints[a["checkpointId"].(string)]
	w := s.workspaces["default"]
	counter := &validationContext{Context: ctx}
	if _, err := project.CaptureSnapshot(counter, w.root, s.rules, cp.Paths, nil); err != nil {
		t.Fatal(err)
	}
	delete(s.snapshots, cp.ID)
	if err := s.validate(&validationContext{Context: ctx, stop: counter.steps + 1}, w, cp, true); err == nil || s.snapshots[cp.ID] != nil {
		t.Fatal("cancelled rebuild published", err)
	}
	saved := ruleCatalogue
	ruleCatalogue = func() ([]meta.Rule, error) { return nil, errors.New("metadata unavailable") }
	defer func() { ruleCatalogue = saved }()
	if _, err := s.ruleInfo(ctx, w, Input{}); err == nil {
		t.Fatal("metadata failure hidden")
	}
	ruleCatalogue = saved
	large := strings.Repeat("x", 65<<20)
	huge := *cp
	huge.Findings = []evidence{{Item: diagnostic.Item{Message: large}}}
	if err := s.persist(&huge); err == nil {
		t.Fatal("oversized checkpoint persisted")
	}
	cfg, err := config.LoadConfined(w.root)
	if err != nil {
		t.Fatal(err)
	}
	over := &project.Snapshot{Config: cfg, Sources: map[string][]byte{}, Results: []diagnostic.FileResult{{Path: "a.php", Findings: []diagnostic.Finding{{Message: large}}}}}
	if err := collect(&checkpoint{}, over); err == nil {
		t.Fatal("oversized finding evidence")
	}
	block := strings.Repeat("x", 34<<20)
	for _, owner := range []string{"alice", "bob"} {
		copy := *cp
		copy.ID = randomID()
		copy.Identity = owner
		copy.Created = time.Now()
		copy.Findings = []evidence{{Item: diagnostic.Item{Message: block}}}
		copy.size = 0
		s.checkpoints[copy.ID] = &copy
	}
	s.prune()
	if len(s.checkpoints) > 2 {
		t.Fatal("global evidence limit not enforced")
	}
	for i := range 1001 {
		if err := s.storage.WriteFile(fmt.Sprintf("ignored-%d.tmp", i), nil, 0o600); err != nil {
			t.Fatal(err)
		}
	}
	if err := s.load(); err == nil {
		t.Fatal("checkpoint directory file limit")
	}
	// A stale plan cannot read an oversized replacement target.
	f, err := w.root.OpenFile("large.php", os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		t.Fatal(err)
	}
	if err := f.Truncate(syntax.MaxFileSize + 1); err != nil {
		t.Fatal(err)
	}
	_ = f.Close()
	if err := replace(w, change{Path: "large.php", Perm: 0o600}); err == nil {
		t.Fatal("oversized replacement target")
	}
	writeFile(t, filepath.Join(root, "a.php"), "<?php $x=!!!$a;")
	savedOpen := openTemporary
	defer func() { openTemporary = savedOpen }()
	openTemporary = func(r *os.Root, name string, perm os.FileMode) (*os.File, error) {
		f, err := savedOpen(r, name, perm)
		if f != nil {
			_ = f.Close()
		}
		return f, err
	}
	if err := replace(w, change{Path: "a.php", Perm: 0o600, Source: []byte("<?php $x=!!!$a;"), Output: []byte("<?php $x=!$a;")}); err == nil {
		t.Fatal("failed write reported success")
	}
}

func TestLongPlanPathsAndTruncatedPreview(t *testing.T) {
	s, ctx, root := fixture(t)
	for i := range 90 {
		writeFile(t, filepath.Join(root, strings.Repeat("a", 100)+fmt.Sprintf("%d.php", i)), "<?php $x=!!!$a;")
	}
	a := invoke(ctx, t, s, "analyze", Input{})
	rejected(ctx, t, s, "prepare_fixes", Input{CheckpointID: a["checkpointId"].(string), Iterative: true})
	p := &plan{ID: randomID(), Expires: time.Now().Add(time.Hour), Changes: []change{{Path: "a.php", Source: []byte(strings.Repeat("a", 4500)), Output: []byte(strings.Repeat("b", 4500))}}}
	result, err := s.planPage(ctx, p, Input{})
	if err != nil || !strings.Contains(result["items"].([]any)[0].(Output)["diff"].(string), "Preview truncated") {
		t.Fatal(result, err)
	}
}

func TestProtocolResultBudget(t *testing.T) {
	s, ctx, _ := fixture(t)
	buffer := "<?php $x = new " + strings.Repeat("A", 40000) + ";"
	a := invoke(ctx, t, s, "analyze", Input{Path: "unsaved.php", Buffer: &buffer})
	id := a["checkpointId"].(string)
	serverTransport, clientTransport := sdk.NewInMemoryTransports()
	ss, err := s.Protocol("test").Connect(ctx, serverTransport, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer ss.Close()
	client := sdk.NewClient(&sdk.Implementation{Name: "test", Version: "1"}, nil)
	cs, err := client.Connect(ctx, clientTransport, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer cs.Close()
	result, err := cs.CallTool(ctx, &sdk.CallToolParams{Name: "custos_type", Arguments: map[string]any{"checkpointId": id, "path": "unsaved.php", "offset": 10}})
	if err != nil || !result.IsError || !strings.Contains(result.Content[0].(*sdk.TextContent).Text, "budget") {
		t.Fatal(result, err)
	}
}

func TestPreviewPreservesBytesAndRuleUpdateInvalidates(t *testing.T) {
	s, ctx, _ := fixture(t)
	for _, text := range []string{strings.Repeat("a", 2047) + "é" + strings.Repeat("z", 3000), string([]byte{0xff, 0xfe, 0x80}) + strings.Repeat("\x00", 3000)} {
		p := &plan{ID: randomID(), Expires: time.Now().Add(time.Hour), Changes: []change{{Path: "a.php", Edits: []diagnostic.TextEdit{{Span: syntax.Span{End: 1}, NewText: text}}}}}
		result, err := s.planPage(ctx, p, Input{})
		if err != nil {
			t.Fatal(err)
		}
		var recovered []byte
		for _, raw := range result["items"].([]any) {
			item := raw.(Output)
			if encoded, ok := item["textBase64"].(string); ok {
				b, err := base64.StdEncoding.DecodeString(encoded)
				if err != nil {
					t.Fatal(err)
				}
				recovered = append(recovered, b...)
			}
		}
		if string(recovered) != text {
			t.Fatal("preview lost replacement bytes")
		}
	}
	a := invoke(ctx, t, s, "analyze", Input{})
	id := a["checkpointId"].(string)
	cp := s.checkpoints[id]
	cp.Findings[0].Message = "old rule behavior"
	delete(s.snapshots, id)
	rejected(ctx, t, s, "findings", Input{CheckpointID: id})
	if !cp.Stale {
		t.Fatal("recovered evidence was not revalidated")
	}
}

func TestPlanCursorAfterCreationAndCorruptManifest(t *testing.T) {
	s, ctx, root := fixture(t)
	a := invoke(ctx, t, s, "analyze", Input{})
	id := a["checkpointId"].(string)
	first := invoke(ctx, t, s, "prepare_fixes", Input{CheckpointID: id, Choices: []Choice{{FindingID: "0"}}, Limit: 1})
	second := invoke(ctx, t, s, "prepare_fixes", Input{CheckpointID: id, PlanID: first["planId"].(string), Cursor: first["nextCursor"].(string), Limit: 1})
	if second["nextCursor"] != nil || len(second["items"].([]any)) != 1 {
		t.Fatal(second)
	}
	writeFile(t, filepath.Join(root, "custos.json"), `{"php":"8.2","baseline":"baseline.json"}`)
	writeFile(t, filepath.Join(root, "baseline.json"), "invalid")
	captured, err := project.CaptureSnapshot(ctx, s.workspaces["default"].root, s.rules, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	cp := s.checkpoints[id]
	cp.Hashes = captured.Hashes
	delete(s.snapshots, id)
	rejected(ctx, t, s, "findings", Input{CheckpointID: id})
	if !cp.Stale {
		t.Fatal("corrupt persisted manifest was not marked stale")
	}
}

func TestCancellationBeforeCommit(t *testing.T) {
	s, ctx, root := fixture(t)
	source, err := os.ReadFile(filepath.Join(root, "a.php"))
	if err != nil {
		t.Fatal(err)
	}
	c := change{Path: "a.php", Source: source, Output: []byte("<?php $x=0;"), Perm: 0o600}
	cancelled, cancel := context.WithCancel(ctx)
	cancel()
	if err := replaceContext(cancelled, s.workspaces["default"], c); err == nil {
		t.Fatal("already cancelled write")
	}
	active, cancel := context.WithCancel(ctx)
	beforeCommit = func(*os.Root, string) { cancel() }
	defer func() { beforeCommit = nil }()
	if err := replaceContext(active, s.workspaces["default"], c); err == nil {
		t.Fatal("cancelled replacement committed")
	}
	after, err := os.ReadFile(filepath.Join(root, "a.php"))
	if err != nil || string(after) != string(source) {
		t.Fatal("cancelled write changed source", err)
	}
}

func TestQuotaUsesIdentityWorkspaceTuple(t *testing.T) {
	s, _, _ := fixture(t)
	for _, owner := range []struct{ identity, workspace string }{{"a\x00b", "c"}, {"a", "b\x00c"}} {
		for range 16 {
			cp := &checkpoint{ID: randomID(), Identity: owner.identity, Workspace: owner.workspace, Created: time.Now()}
			s.checkpoints[cp.ID] = cp
		}
	}
	s.prune()
	if len(s.checkpoints) != 32 {
		t.Fatal("quota identities collided", len(s.checkpoints))
	}
}

func TestBufferDeclarationsAndOverlayFailure(t *testing.T) {
	s, ctx, _ := fixture(t)
	buffer := "<?php class Fresh { public function answer(): int { return 42; } } $x=new Fresh(); $x->answer();"
	a := invoke(ctx, t, s, "analyze", Input{Path: "unsaved.php", Buffer: &buffer})
	id := a["checkpointId"].(string)
	symbols := invoke(ctx, t, s, "symbols", Input{CheckpointID: id, Query: "Fresh"})
	if symbols["total"] != 2 {
		t.Fatal(symbols)
	}
	result := invoke(ctx, t, s, "type", Input{CheckpointID: id, Path: "unsaved.php", Offset: uint32(strings.LastIndex(buffer, "answer"))})
	if result["native"] != "int" {
		t.Fatal(result)
	}
	// Cancellation after captured sources have been read must not publish a
	// partially indexed supplied buffer.
	root := s.workspaces["default"].root
	counter := &validationContext{Context: ctx}
	if _, err := project.CaptureSnapshot(counter, root, s.rules, nil, nil); err != nil {
		t.Fatal(err)
	}
	_, err := s.analyze(&validationContext{Context: ctx, stop: counter.steps + 1}, s.workspaces["default"], Input{Path: "unsaved.php", Buffer: &buffer})
	if err == nil {
		t.Fatal("cancelled overlay published")
	}
}

func TestGlobalCheckpointCountSupportsRestart(t *testing.T) {
	s, _, _ := fixture(t)
	for i := range 1001 {
		id := randomID()
		s.checkpoints[id] = &checkpoint{ID: id, Identity: id, Workspace: "default", Created: time.Now().Add(time.Duration(i) * -time.Second)}
	}
	s.prune()
	if len(s.checkpoints) != 1000 {
		t.Fatalf("retained %d checkpoints", len(s.checkpoints))
	}
}
