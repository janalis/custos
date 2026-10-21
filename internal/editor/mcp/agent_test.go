package mcp

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	sdk "github.com/modelcontextprotocol/go-sdk/mcp"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/meta"
	"custos/internal/inspection/rules/nestednotoperators"
	"custos/internal/semantic/index"
)

const declarations = `<?php
namespace App;

interface Shape { public function area(): float; }
trait Named { protected static function label(string ...$parts) {} }
enum Suit: string { case Hearts = 'h'; }
/** @method int magic(array $criteria, $limit = null) */
abstract class Base implements Shape {
    abstract protected function make(int &$count, $loose): Base;
    final public function done() { return 1; }
    private function hidden() {}
    /** @return string */
    public function documented() {}
}
final class Circle extends Base {
    public function area(): float { return 1.0; }
    protected function make(int &$count, $loose): Base { return $this; }
}
readonly class Point { public function __construct(public int $x = 0) {} }
const LIMIT = 10;
/** @param string $name */
function greet($name) {}
function recircle(): int { return 1; }
function encircled(): int { return 2; }
$anonymous = new class { public function inner() {} };
$é = !!!$token; $second = !!!$token;
`

// agentFixture is a project with a vendor package, a symlinked vendor entry
// and declarations of every kind.
func agentFixture(t *testing.T) (*Server, context.Context, string, string) {
	t.Helper()
	root, storage := t.TempDir(), t.TempDir()
	if err := os.Chmod(storage, 0o700); err != nil {
		t.Fatal(err)
	}
	for name, text := range map[string]string{
		"custos.json":          `{"php":"8.2","comparisonStyle":"yoda","rules":{"NestedNotOperators":{"enabled":true,"severity":"warning"}}}`,
		"src/shapes.php":       declarations,
		"src/deep/more.php":    "<?php\n$more = !!!$token;\n",
		"vendor/pkg/Tool.php":  "<?php\nnamespace Vendor;\n\nclass Tool { public function run(): int { return 1; } }\nfunction circle(): int { return 3; }\ndefine('App\\DYNAMIC', time());\n",
		"vendor/pkg/other.php": "<?php\nfunction vendor_other() {}\n",
	} {
		p := filepath.Join(root, name)
		if err := os.MkdirAll(filepath.Dir(p), 0o700); err != nil {
			t.Fatal(err)
		}
		writeFile(t, p, text)
	}
	if err := os.Symlink(t.TempDir(), filepath.Join(root, "vendor", "linked")); err != nil {
		t.Fatal(err)
	}
	s, err := New(Options{Default: "default", Storage: storage, Workspaces: []Workspace{{ID: "default", Root: root, Readers: []string{"local"}, Writers: []string{"local"}}}}, []analysis.Rule{nestednotoperators.New()})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(s.Close)
	ctx := context.WithValue(context.WithValue(context.Background(), identityKey{}, "local"), writeScopeKey{}, true)
	a := invoke(ctx, t, s, "analyze", Input{})
	if !strings.Contains(a["limitations"].(string), "vendor/linked (symlink)") {
		t.Fatal(a)
	}
	summary := a["summary"].(Output)
	if summary["bySeverity"].(map[string]int)["warning"] != 3 || summary["byPath"].([]Output)[0]["path"] != "src/shapes.php" || len(summary["byRule"].([]Output)) != 1 {
		t.Fatal(summary)
	}
	return s, ctx, root, a["checkpointId"].(string)
}

func symbolNames(out Output) string {
	var all []string
	for _, item := range out["items"].([]any) {
		all = append(all, item.(Output)["name"].(string))
	}
	return strings.Join(all, " ")
}

func TestSymbolSearchRanksAndRenders(t *testing.T) {
	s, ctx, root, cp := agentFixture(t)
	// Exact name, whole trailing segment, suffix, then substring; project
	// declarations precede vendor ones of the same rank.
	ranked := invoke(ctx, t, s, "symbols", Input{CheckpointID: cp, Query: `\circle`, Kind: "function"})
	if symbolNames(ranked) != `Vendor\circle App\recircle App\encircled` {
		t.Fatal(symbolNames(ranked))
	}
	ranked = invoke(ctx, t, s, "symbols", Input{CheckpointID: cp, Query: "circle"})
	if !strings.HasPrefix(symbolNames(ranked), `App\Circle Vendor\circle App\recircle App\Circle::area`) || ranked["scope"] != "all" {
		t.Fatal(symbolNames(ranked))
	}
	signatures := map[string]string{}
	all := invoke(ctx, t, s, "symbols", Input{CheckpointID: cp, Limit: 100})
	for _, raw := range all["items"].([]any) {
		item := raw.(Output)
		signatures[item["name"].(string)] = item["signature"].(string)
		if strings.Contains(item["name"].(string), "@anonymous") || item["provenance"] != "project" {
			t.Fatal(item)
		}
	}
	for name, want := range map[string]string{
		`App\Shape`:            `interface App\Shape`,
		`App\Named`:            `trait App\Named`,
		`App\Suit`:             `enum App\Suit implements UnitEnum, BackedEnum`,
		`App\Base`:             `abstract class App\Base implements App\Shape`,
		`App\Circle`:           `final class App\Circle extends App\Base`,
		`App\Point`:            `readonly class App\Point`,
		`App\Named::label`:     `protected static function label(string ...$parts)`,
		`App\Base::make`:       `abstract protected function make(int &$count, $loose): \App\Base`,
		`App\Base::done`:       `final public function done(): int`,
		`App\Base::hidden`:     `private function hidden()`,
		`App\Base::documented`: `public function documented(): string`,
		`App\Base::magic`:      `public function magic(array $criteria, $limit = null): int`,
		`App\LIMIT`:            `const App\LIMIT = 10`,
		`App\greet`:            `function App\greet(string $name): null`,
		`App\recircle`:         `function App\recircle(): int`,
	} {
		if signatures[name] != want {
			t.Errorf("%s: %q, want %q", name, signatures[name], want)
		}
	}
	magic := invoke(ctx, t, s, "symbols", Input{CheckpointID: cp, Query: "base::magic"})["items"].([]any)[0].(Output)
	if magic["declaredBy"] == nil || magic["line"] != nil || magic["returnTypeFrom"] != "phpdoc" {
		t.Fatal(magic)
	}
	vendor := invoke(ctx, t, s, "symbols", Input{CheckpointID: cp, Scope: "vendor"})
	first := vendor["items"].([]any)[0].(Output)
	if vendor["total"] != 5 || first["name"] != `App\DYNAMIC` || first["signature"] != `const App\DYNAMIC = time()` || constantSignature(&index.Constant{FQN: "X"}) != "const X" || first["line"] != 6 || first["provenance"] != "vendor" {
		t.Fatal(vendor)
	}
	if out := invoke(ctx, t, s, "symbols", Input{CheckpointID: cp, Query: "run", Path: "vendor/pkg/Tool.php"}); out["total"] != 1 {
		t.Fatal(out)
	}
	if out := invoke(ctx, t, s, "symbols", Input{CheckpointID: cp, Query: "tool", Scope: "project"}); out["total"] != 0 {
		t.Fatal(out)
	}
	// A vendor file that changed after analysis keeps its byte range but
	// loses the line and column derived from its source.
	snap := s.snapshots[cp]
	writeFile(t, filepath.Join(root, "vendor/pkg/Tool.php"), "<?php // replaced")
	stale, err := s.symbols(ctx, s.checkpoints[cp], snap, Input{Query: `vendor\tool`})
	item := stale["items"].([]any)[0].(Output)
	if err != nil || item["line"] != nil || item["start"] == nil {
		t.Fatal(stale, err)
	}
}

func TestTypeByLineAndColumn(t *testing.T) {
	s, ctx, _, cp := agentFixture(t)
	before, _, _ := strings.Cut(declarations, "$é")
	line := strings.Count(before, "\n") + 1
	column := len([]rune("$é = !!!")) + 1
	out := invoke(ctx, t, s, "type", Input{CheckpointID: cp, Path: "src/shapes.php", Line: line, Column: column})
	if out["expression"] != "$token" || out["line"] != line || out["column"] != column {
		t.Fatal(out)
	}
	// Without a column the line start is used; a column past the line end
	// stops at the line break.
	if out := invoke(ctx, t, s, "type", Input{CheckpointID: cp, Path: "src/shapes.php", Line: line}); out["expression"] != "$é" {
		t.Fatal(out)
	}
	if out := invoke(ctx, t, s, "type", Input{CheckpointID: cp, Path: "src/shapes.php", Line: line, Column: 500}); out["resolved"] != false {
		t.Fatal(out)
	}
	vendor := invoke(ctx, t, s, "type", Input{CheckpointID: cp, Path: "vendor/pkg/Tool.php", Line: 4, Column: strings.Index("class Tool { public function run(): int { return 1; } }", "1;") + 1})
	if vendor["native"] != "int" {
		t.Fatal(vendor)
	}
	for _, in := range []Input{
		{Path: "src/shapes.php", Line: 5000},
		{Path: "src/shapes.php", Offset: 1 << 20},
		{Path: "src/missing.php", Line: 1},
		{Path: "custos.json", Line: 1},
	} {
		in.CheckpointID = cp
		rejected(ctx, t, s, "type", in)
	}
}

func TestFindingFiltersAndLeanItems(t *testing.T) {
	s, ctx, _, cp := agentFixture(t)
	for path, total := range map[string]int{"": 3, ".": 3, "src": 3, "src/": 3, "src/deep": 1, "src/shapes.php": 2, "src/sha": 0} {
		out := invoke(ctx, t, s, "findings", Input{CheckpointID: cp, Path: path, Rule: "nestednotoperatorsinspection"})
		if out["total"] != total {
			t.Fatal(path, out)
		}
	}
	out := invoke(ctx, t, s, "findings", Input{CheckpointID: cp, Path: "src/shapes.php"})
	b, _ := json.Marshal(out)
	if strings.Contains(string(b), "fingerprint") || out["untrusted"] == nil || out["summary"].(Output)["byRule"].([]Output)[0]["count"] != 2 {
		t.Fatal(string(b))
	}
	if s.checkpoints[cp].Findings[0].Fingerprint == "" {
		t.Fatal("persisted evidence lost its fingerprint")
	}
	// Excerpts never split a multibyte character.
	src := []byte("aé" + strings.Repeat("x", 10) + "éb")
	if got := excerpt(src, 3, 13, 1, 0); got != strings.Repeat("x", 10) {
		t.Fatal(got)
	}
	if got := excerpt(src, 0, 100, 0, 2); got != "a" {
		t.Fatal(got)
	}
	if got := excerpt(src, 500, 600, 80, 0); got != "" {
		t.Fatal(got)
	}
}

func TestRuleSearchAndSettingsViews(t *testing.T) {
	s, ctx, root, _ := agentFixture(t)
	found := invoke(ctx, t, s, "rules", Input{Query: "nested  NOT operators"})
	item := found["items"].([]any)[0].(Output)
	if found["total"] != 1 || item["override"].(Output)["severity"] == nil || item["override"].(Output)["enabled"] != true {
		t.Fatal(found)
	}
	if out := invoke(ctx, t, s, "rules", Input{Rule: "nestednotoperatorsinspection"}); out["total"] != 1 {
		t.Fatal(out)
	}
	if out := invoke(ctx, t, s, "rules", Input{Rule: "UnnecessaryAssertion"}); out["items"].([]any)[0].(Output)["override"] != nil {
		t.Fatal(out)
	}
	settings := invoke(ctx, t, s, "workspace", Input{})["items"].([]any)[0].(Output)["settings"].(Output)
	if settings["php"] != "8.2" || settings["comparisonStyle"] != "yoda" || len(settings["rules"].(Output)) != 1 {
		t.Fatal(settings)
	}
	view := overrideView(analysis.RuleConfig{Options: map[string]any{"@derived": 1, "CHOSEN": true}})
	if len(view["options"].(Output)) != 1 || view["enabled"] != nil {
		t.Fatal(view)
	}
	writeFile(t, filepath.Join(root, "custos.json"), `{"php":"8.2"}`)
	settings = invoke(ctx, t, s, "workspace", Input{})["items"].([]any)[0].(Output)["settings"].(Output)
	if settings["comparisonStyle"] != "regular" || len(settings["rules"].(Output)) != 0 {
		t.Fatal(settings)
	}
}

func TestStaleCheckpointsNameChangedInputs(t *testing.T) {
	s, ctx, root, cp := agentFixture(t)
	for _, name := range []string{"src/shapes.php", "src/deep/more.php", "vendor/pkg/Tool.php", "vendor/pkg/other.php"} {
		writeFile(t, filepath.Join(root, name), "<?php // edited")
	}
	_, err := s.call(ctx, "findings", Input{CheckpointID: cp})
	if err == nil || !strings.Contains(err.Error(), "(src/deep/more.php, src/shapes.php, vendor/pkg/Tool.php, +1 more)") {
		t.Fatal(err)
	}
	if !s.checkpoints[cp].Stale || s.snapshots[cp] != nil {
		t.Fatal("changed checkpoint kept")
	}
	// Without a snapshot the manifests are compared: a vendor change is
	// reported as such, and added or removed inputs are named.
	next := invoke(ctx, t, s, "analyze", Input{})["checkpointId"].(string)
	delete(s.snapshots, next)
	writeFile(t, filepath.Join(root, "vendor/pkg/other.php"), "<?php // again")
	_, err = s.call(ctx, "findings", Input{CheckpointID: next})
	if err == nil || !strings.Contains(err.Error(), "(vendor files)") {
		t.Fatal(err)
	}
	next = invoke(ctx, t, s, "analyze", Input{})["checkpointId"].(string)
	delete(s.snapshots, next)
	if err := os.Remove(filepath.Join(root, "src/deep/more.php")); err != nil {
		t.Fatal(err)
	}
	_, err = s.call(ctx, "symbols", Input{CheckpointID: next})
	if err == nil || !strings.Contains(err.Error(), "(src/deep/more.php)") {
		t.Fatal(err)
	}
	// A stale checkpoint stays listed and says so when resumed.
	rejected(ctx, t, s, "findings", Input{CheckpointID: next})
	resumed := invoke(ctx, t, s, "checkpoints", Input{CheckpointID: next, Action: "resume"})
	if resumed["stale"] != true || !strings.Contains(resumed["limitation"].(string), "stale") {
		t.Fatal(resumed)
	}
	if listed := invoke(ctx, t, s, "checkpoints", Input{}); listed["total"] != 3 {
		t.Fatal(listed)
	}
	// A failing walk marks the checkpoint stale unless the caller cancelled.
	next = invoke(ctx, t, s, "analyze", Input{})["checkpointId"].(string)
	if err := os.Symlink(t.TempDir(), filepath.Join(root, "src", "link")); err != nil {
		t.Fatal(err)
	}
	rejected(ctx, t, s, "findings", Input{CheckpointID: next})
	if !s.checkpoints[next].Stale {
		t.Fatal("unwalkable inputs kept fresh")
	}
}

func TestArgumentsAreCheckedBeforeSnapshots(t *testing.T) {
	s, ctx, root, cp := agentFixture(t)
	// With the snapshot gone, any call reaching validation would rebuild it.
	delete(s.snapshots, cp)
	for _, in := range []Input{
		{},
		{PlanID: "missing"},
		{Choices: []Choice{{FindingID: "missing"}}},
		{Choices: []Choice{{FindingID: "0", Fix: 5}}},
		{Choices: []Choice{{FindingID: "0", Fix: -1}}},
		{Iterative: true, Rule: "unknown-rule"},
	} {
		in.CheckpointID = cp
		rejected(ctx, t, s, "prepare_fixes", in)
	}
	rejected(ctx, t, s, "findings", Input{CheckpointID: cp, Rule: "unknown-rule"})
	if s.snapshots[cp] != nil {
		t.Fatal("rejected arguments rebuilt a snapshot")
	}
	buffer := "<?php"
	rejected(ctx, t, s, "analyze", Input{Buffer: &buffer, Path: "not-php.txt"})
	if _, err := s.prepare(ctx, s.workspaces["default"], s.checkpoints[cp], nil, Input{}); err == nil {
		t.Fatal("unchecked preparation")
	}
	saved := ruleCatalogue
	ruleCatalogue = func() ([]meta.Rule, error) { return nil, errors.New("metadata unavailable") }
	rejected(ctx, t, s, "findings", Input{CheckpointID: cp, Rule: "NestedNotOperators"})
	ruleCatalogue = saved

	// Applying checks freshness without analyzing again, then supersedes
	// the checkpoint with a fresh analysis.
	plan := invoke(ctx, t, s, "prepare_fixes", Input{CheckpointID: cp, Iterative: true, Path: "src/deep/more.php"})
	edit := plan["items"].([]any)[1].(Output)
	if plan["total"] != 2 || edit["text"] != "$more = !$token;\n" || edit["start"] != uint32(6) {
		t.Fatal(plan)
	}
	delete(s.snapshots, cp)
	applied := invoke(ctx, t, s, "apply_fixes", Input{PlanID: plan["planId"].(string)})
	if applied["partialFailure"] != false || s.checkpoints[cp] != nil || applied["analysis"].(Output)["findings"] != 2 {
		t.Fatal(applied)
	}
	src, err := os.ReadFile(filepath.Join(root, "src/deep/more.php"))
	if err != nil || string(src) != "<?php\n$more = !$token;\n" {
		t.Fatal(string(src), err)
	}
}

func TestCallsQueueAndCursorsExplainThemselves(t *testing.T) {
	s, ctx, _, cp := agentFixture(t)
	var wg sync.WaitGroup
	failures := make(chan error, 8)
	for range 8 {
		wg.Go(func() {
			if _, err := s.call(ctx, "findings", Input{CheckpointID: cp}); err != nil {
				failures <- err
			}
		})
	}
	wg.Wait()
	close(failures)
	for err := range failures {
		t.Fatal("queued call failed", err)
	}
	for cursor, want := range map[string]string{
		"???": "invalid cursor",
		base64.RawURLEncoding.EncodeToString([]byte("no-separator")): "invalid cursor",
		base64.RawURLEncoding.EncodeToString([]byte("other:0")):      "another selection",
	} {
		if _, err := page(ctx, "scope", Input{Cursor: cursor}, []any{1}); err == nil || !strings.Contains(err.Error(), want) {
			t.Fatal(cursor, err)
		}
	}
}

func TestResultsAreOneTextBlock(t *testing.T) {
	s, ctx, _, _ := agentFixture(t)
	serverTransport, clientTransport := sdk.NewInMemoryTransports()
	ss, err := s.Protocol("test").Connect(ctx, serverTransport, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer ss.Close()
	session, err := sdk.NewClient(&sdk.Implementation{Name: "test", Version: "1"}, nil).Connect(ctx, clientTransport, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer session.Close()
	list, err := session.ListTools(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	for _, tool := range list.Tools {
		if tool.OutputSchema != nil {
			t.Fatal("output schema advertised without structured content", tool.Name)
		}
	}
	result, err := session.CallTool(ctx, &sdk.CallToolParams{Name: "custos_workspace", Arguments: map[string]any{}})
	if err != nil || result.IsError || len(result.Content) != 1 || result.StructuredContent != nil {
		t.Fatal(result, err)
	}
}

func TestCallTurnsACrashIntoAnError(t *testing.T) {
	s, ctx, _, _ := agentFixture(t)
	s.rules = []analysis.Rule{nil}
	_, err := s.call(ctx, "analyze", Input{})
	if err == nil || !strings.HasPrefix(err.Error(), "internal failure in analyze: ") {
		t.Fatal(err)
	}
	s.rules = []analysis.Rule{nestednotoperators.New()}
	if _, err := s.call(ctx, "analyze", Input{}); err != nil {
		t.Fatal("the server must keep serving after a crash:", err)
	}
}
