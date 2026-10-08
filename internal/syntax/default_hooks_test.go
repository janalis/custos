package syntax

import (
	"testing"

	"custos/internal/phpver"
)

// A `{` after a property or promoted-parameter default opens the hook list,
// also in permissive mode, where legacy `$a{0}` offsets are accepted.
func TestDefaultBeforeHooksPermissive(t *testing.T) {
	src := `<?php class A {
    public ?string $a = null { set(?string $v) { $this->a = $v; } }
    public function __construct(public int $b = FOO { get => 1; }) {}
}
$x = $s{0};`
	f := Parse("a.php", []byte(src), Options{Version: phpver.Max, Permissive: true})
	if len(f.Errors) > 0 {
		t.Fatalf("errors: %v", f.Errors)
	}
	cl := f.Stmts[0].(*ClassLike)
	if p := cl.Members[0].(*Property); len(p.Hooks) != 1 {
		t.Fatalf("property hooks: %d", len(p.Hooks))
	}
	if m := cl.Members[1].(*Method); len(m.Params[0].Hooks) != 1 {
		t.Fatalf("parameter hooks: %d", len(m.Params[0].Hooks))
	}
	as := f.Stmts[1].(*ExprStmt).Expr.(*Assign)
	if d, ok := as.Value.(*ArrayDimFetch); !ok || !d.Curly {
		t.Fatalf("legacy offset: %T", as.Value)
	}
}
