package infer_test

import (
	"strings"
	"testing"

	"custos/internal/index"
	"custos/internal/infer"
	"custos/internal/names"
	"custos/internal/phpver"
	"custos/internal/syntax"
)

func TestAnonymousClassMemberInference(t *testing.T) {
	check(t, `<?php
 trait Members { public function traitValue(): int { return 1; } }
 class Base { public function inherited(): string { return ''; } }
 $x = new class(2) extends Base {
 use Members;
 public function __construct(public int $value) {}
 public function myself(): self { return $this; }
 public function late(): static { return $this; }
 public function ancestor(): parent { return new Base; }
 final public function inferred() { return $this->value; }
 public function inside() { t('this-member', $this->value); t('self-member', (new self(3))->value); t('parent-member', (new parent)->inherited()); }
 };
 t('property', $x->value); t('trait', $x->traitValue()); t('inherited', $x->inherited());
 t('self', $x->myself()->value); t('static', $x->late()->value); t('parent', $x->ancestor()->inherited()); t('body', $x->inferred());
 `, map[string]string{"property": "int", "trait": "int", "inherited": "string", "self": "int", "static": "int", "parent": "string", "body": "int", "this-member": "int", "self-member": "int", "parent-member": "string"})
}

func TestAnonymousClassCrossFileReturn(t *testing.T) {
	factory := syntax.Parse("factory.php", []byte(`<?php function factory() { return new class { public string $value = ''; public function text(): string { return ''; } }; }`), syntax.Options{Version: phpver.PHP84})
	consumer := syntax.Parse("consumer.php", []byte(`<?php factory()->text(); factory()->value;`), syntax.Options{Version: phpver.PHP84})
	fs := index.Extract(factory)
	infer.AnnotateReturns(factory, fs, nil, phpver.PHP84)
	ix := index.New(nil)
	ix.Add(fs)
	ix.Add(index.Extract(consumer))
	env := infer.NewEnv(consumer, names.New(consumer), ix, phpver.PHP84)
	for _, st := range consumer.Stmts {
		es, ok := st.(*syntax.ExprStmt)
		if !ok {
			continue
		}
		if got := env.TypeOf(es.Expr).String(); got != "string" {
			t.Errorf("cross-file member: %s", got)
		}
	}
	for _, fn := range fs.Functions {
		if !strings.Contains(fn.Inferred, "@anonymous:") {
			t.Errorf("missing indexed identity: %s", fn.Inferred)
		}
	}
}

func TestAnonymousClassConstructorReference(t *testing.T) {
	f := syntax.Parse("constructor.php", []byte(`<?php $x = new class(1) { public function __construct(public int $value) {} };`), syntax.Options{Version: phpver.PHP84})
	nm := names.New(f)
	ix := index.New(nil)
	ix.Add(index.Extract(f))
	env := infer.NewEnv(f, nm, ix, phpver.PHP84)
	syntax.InspectFile(f, func(n syntax.Node) bool {
		if cl, ok := n.(*syntax.ClassLike); ok {
			ref := env.ClassRef(cl)
			if ref != nm.SymbolFQN(cl) || ix.FindMethod(ref, "__construct", phpver.PHP84) == nil {
				t.Errorf("anonymous constructor reference: %s", ref)
			}
		}
		return true
	})
}
