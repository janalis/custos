package infer_test

import (
	"strings"
	"testing"

	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
	"custos/internal/semantic/index"
	"custos/internal/semantic/infer"
	"custos/internal/semantic/names"
	"custos/internal/semantic/stubs"
)

// Inside a trait, `$this->p` takes the type the trait declares for p.
func TestTraitOwnProperties(t *testing.T) {
	checkWith(t, nil, `<?php
class Dto { public function name(): string { return ''; } }
trait Inner { protected ?Dto $inner = null; }
/** @property Dto $magic */
trait HasDto {
    use Inner;
    private Dto $dto;
    /** @var list<Dto> */
    private array $all = [];
    private ?self $parent = null;
    /** @var HasDto|null */
    private $named;
    private static Dto $shared;
    public $untyped;
    public function run() {
        t('dto', $this->dto);
        t('call', $this->dto->name());
        t('doc', $this->all);
        t('inherited', $this->inner);
        t('self', $this->parent);
        t('named', $this->named);
        t('magic', $this->magic);
        t('undeclared', $this->nothing);
        t('untyped', $this->untyped);
        t('dynamic', $this->{'dto'});
        $fn = function () { t('closure', $this->dto); };
    }
}
class User { use HasDto; public function f($o) { t('other', $o->dto); } }
`, map[string]string{
		"dto": `\Dto`, "call": "string", "doc": `\Dto[]`, "inherited": `\Dto|null`,
		"self": "?unknown", "named": "?unknown", "magic": "?unknown", "undeclared": "?unknown", "untyped": "?unknown",
		"dynamic": "?unknown", "closure": `\Dto`, "other": "?unknown",
	})
}

// Inside a trait, `$this->m()` takes the return type the trait declares for
// m, outside the native environment.
func TestTraitOwnMethods(t *testing.T) {
	src := `<?php
class Dto {}
trait Base { abstract protected function make(): Dto; }
/** @method Dto magic() */
trait Uses {
    use Base;
    protected function name(): string { return ''; }
    protected function me(): self { return $this; }
    /** @return Uses */
    protected function named() { return $this; }
    /** @return list<Dto> */
    protected function all(): array { return []; }
    protected function body() { return 1; }
    public function run() {
        t('declared', $this->name());
        t('doc', $this->all());
        t('abstract', $this->make());
        t('self', $this->me());
        t('namedTrait', $this->named());
        t('body', $this->body());
        t('magic', $this->magic());
        t('missing', $this->nothing());
    }
}
`
	checkWith(t, nil, src, map[string]string{
		"declared": "string", "doc": `\Dto[]`, "abstract": `\Dto`, "self": "?unknown", "namedTrait": "?unknown",
		"body": "?unknown", "magic": "?unknown", "missing": "?unknown",
	})
	checkNative(t, src, map[string]string{"declared": "?unknown", "abstract": "?unknown"})
}

// checkNative is checkWith (no other files) through the native environment.
func checkNative(t *testing.T, src string, want map[string]string) {
	t.Helper()
	f := syntax.Parse("t.php", []byte(src), syntax.Options{Version: phpversion.PHP84})
	ix := index.New(stubs.Index())
	ix.Add(index.Extract(f))
	env := infer.NewEnv(f, names.New(f), ix, phpversion.PHP84).Native()
	got := map[string]string{}
	syntax.InspectFile(f, func(n syntax.Node) bool {
		if c, ok := n.(*syntax.FuncCall); ok {
			if nm, ok := c.Name.(*syntax.Name); ok && nm.Value == "t" && len(c.Args.Args) == 2 {
				label := strings.Trim(string(f.Src[c.Args.Args[0].Span().Start:c.Args.Args[0].Span().End]), "'")
				got[label] = env.TypeOf(c.Args.Args[1].(*syntax.Arg).Value).ShapeString()
			}
		}
		return true
	})
	for k, w := range want {
		if got[k] != w {
			t.Errorf("native %s: got %s want %s", k, got[k], w)
		}
	}
}
