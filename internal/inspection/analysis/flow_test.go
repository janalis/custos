package analysis

import (
	"testing"

	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
	"custos/internal/semantic/flow"
	"custos/internal/semantic/index"
)

type flowProbeRule struct{}

func (flowProbeRule) ID() string                  { return "UntrustedShellCommand" }
func (flowProbeRule) Kinds() []syntax.NodeKind    { return nil }
func (flowProbeRule) Check(*Context, syntax.Node) {}
func (flowProbeRule) Flow()                       {}

func TestFlowEngineCopiesAndLocalBufferShadow(t *testing.T) {
	opt := syntax.Options{Version: phpversion.Default}
	indexed := syntax.Parse("source.php", []byte("<?php function source() { return $_GET['value']; }"), opt)
	ix := index.New(nil)
	ix.Add(index.Extract(indexed))
	snapshot := flow.NewSnapshot(flow.Extract(indexed, ix, opt.Version, nil))
	original, err := NewEngine([]Rule{flowProbeRule{}}, Config{EnableAll: true, PHP: opt.Version})
	if err != nil {
		t.Fatal(err)
	}
	if !original.NeedsFlow() || !original.NeedsIndex() {
		t.Fatal("flow inspections must request both project summaries and symbols")
	}
	e := original.WithIndex(ix).WithFlow(snapshot)
	if original.flow != nil || e.flow != snapshot || e == original {
		t.Fatal("flow configuration mutated the published engine")
	}
	check := func(path, src string, want bool) {
		t.Helper()
		f := syntax.Parse(path, []byte(src), opt)
		ctx := NewTestContext(e, f)
		env := ctx.Flow()
		if env != ctx.Flow() {
			t.Fatal("file-local flow environment was rebuilt")
		}
		var arg syntax.Expr
		syntax.InspectFile(f, func(n syntax.Node) bool {
			if c, ok := n.(*syntax.FuncCall); ok {
				if name, ok := c.Name.(*syntax.Name); ok && name.Value == "probe" {
					arg = c.Args.Args[0].(*syntax.Arg).Value
				}
			}
			return true
		})
		if arg == nil || env.Tainted(arg, flow.Shell) != want {
			t.Fatalf("tainted buffer=%v, want %v", env.Value(arg), want)
		}
	}
	check("main.php", "<?php probe(source());", true)
	check("source.php", "<?php function source() { return 'fixed'; } probe(source());", false)
	check("main.php", "<?php probe(source());", true)
}
