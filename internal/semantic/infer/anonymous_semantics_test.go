package infer_test

import (
	"strings"
	"testing"

	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
	"custos/internal/semantic/index"
	"custos/internal/semantic/infer"
	"custos/internal/semantic/names"
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
	factory := syntax.Parse("factory.php", []byte(`<?php function factory() { return new class { public string $value = ''; public function text(): string { return ''; } }; }`), syntax.Options{Version: phpversion.PHP84})
	consumer := syntax.Parse("consumer.php", []byte(`<?php factory()->text(); factory()->value;`), syntax.Options{Version: phpversion.PHP84})
	fs := index.Extract(factory)
	infer.AnnotateReturns(factory, fs, nil, phpversion.PHP84)
	ix := index.New(nil)
	ix.Add(fs)
	ix.Add(index.Extract(consumer))
	env := infer.NewEnv(consumer, names.New(consumer), ix, phpversion.PHP84)
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
	f := syntax.Parse("constructor.php", []byte(`<?php $x = new class(1) { public function __construct(public int $value) {} };`), syntax.Options{Version: phpversion.PHP84})
	nm := names.New(f)
	ix := index.New(nil)
	ix.Add(index.Extract(f))
	env := infer.NewEnv(f, nm, ix, phpversion.PHP84)
	syntax.InspectFile(f, func(n syntax.Node) bool {
		if cl, ok := n.(*syntax.ClassLike); ok {
			ref := env.ClassRef(cl)
			if ref != nm.SymbolFQN(cl) || ix.FindMethod(ref, "__construct", phpversion.PHP84) == nil {
				t.Errorf("anonymous constructor reference: %s", ref)
			}
		}
		return true
	})
}

// Anonymous classes are typed as the intersection of their parent and
// interfaces: members of either side are found; their own extra methods
// are not (no name could be written for the class).
func TestAnonymousClassMembers(t *testing.T) {
	src := `<?php
interface I { public function i(): string; }
abstract class P { public function p(): float { return 1.0; } }
function f() {
    $a = new class extends P implements I {
        public function i(): string { return ''; }
        public function own(): int { return 1; }
    };
    t('a', $a);
    t('i', $a->i());
    t('p', $a->p());
    t('own', $a->own());
    t('parentOnly', new class extends P {});
    t('ifaceOnly', new class implements I { public function i(): string { return ''; } });
    t('plain', new class {});
}
`
	f := syntax.Parse("t.php", []byte(src), syntax.Options{Version: phpversion.PHP84})
	r := names.New(f)
	var identities []string
	syntax.InspectFile(f, func(n syntax.Node) bool {
		if c, ok := n.(*syntax.ClassLike); ok && c.Name == nil {
			identities = append(identities, `\`+r.SymbolFQN(c))
		}
		return true
	})
	checkAnywhere(t, src, map[string]string{
		"a": identities[0], "i": "string", "p": "float", "own": "int",
		"parentOnly": identities[1], "ifaceOnly": identities[2], "plain": identities[3],
	})
}
