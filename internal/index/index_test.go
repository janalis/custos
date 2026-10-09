package index

import (
	"testing"

	"custos/internal/phpver"
	"custos/internal/syntax"
)

func extract(t *testing.T, path, src string) *FileSymbols {
	t.Helper()
	return Extract(syntax.Parse(path, []byte(src), syntax.Options{Version: phpver.PHP84}))
}

func TestExtractAndLookup(t *testing.T) {
	ix := New(nil)
	ix.Add(extract(t, "a.php", `<?php
namespace App;
use Base\Model;
interface HasName { public function name(): string; }
trait Greets { public function greet(): string { return "hi"; } }
/**
 * @property-read int $id
 * @method static self make()
 */
final class User extends Model implements HasName {
    use Greets;
    public const ROLE = 'user';
    /** @var string[] */
    private array $tags = [];
    public function __construct(private readonly string $email) {}
    public function name(): string { return ''; }
    /** @return static */
    public static function create(?int $id = null, string ...$rest) { return new static(); }
}
function helper(int $x): bool { return true; }
const LIMIT = 10;
define('GLOBAL_FLAG', true);
`))
	ix.Add(extract(t, "b.php", `<?php
namespace Base;
abstract class Model { protected $table; public function save(): void {} }
`))
	u := ix.Class(`\App\User`, 0)
	if u == nil || !u.Final || u.Parent != `Base\Model` || len(u.Interfaces) != 1 || u.Traits[0] != `App\Greets` {
		t.Fatalf("class: %+v", u)
	}
	if p := u.Props["email"]; p == nil || !p.Promoted || !p.Readonly || p.Visibility != Private || p.Type != "string" {
		t.Fatalf("promoted: %+v", p)
	}
	if p := u.Props["tags"]; p == nil || p.DocType != "string[]" || !p.HasDefault {
		t.Fatalf("tags: %+v", p)
	}
	if p := u.Props["id"]; p == nil || !p.Magic || p.DocType != "int" {
		t.Fatalf("magic prop: %+v", p)
	}
	m := u.Methods["create"]
	if m == nil || !m.Static || m.DocReturn != "static" || len(m.Params) != 2 || m.Params[0].Type != "int|null" || !m.Params[1].Variadic {
		t.Fatalf("create: %+v", m)
	}
	if ix.FindMethod(`App\User`, "GREET", 0) == nil || ix.FindMethod(`App\User`, "save", 0) == nil {
		t.Fatal("inherited methods not found")
	}
	if ix.FindProperty(`App\User`, "table", 0) == nil || ix.FindConst(`App\User`, "ROLE", 0) == nil {
		t.Fatal("inherited members")
	}
	if !ix.IsSubtype(`App\User`, `\app\hasname`, 0) || ix.IsSubtype(`Base\Model`, `App\User`, 0) {
		t.Fatal("IsSubtype")
	}
	if f := ix.Function(`app\HELPER`, 0); f == nil || f.Return != "bool" {
		t.Fatalf("function: %+v", f)
	}
	if ix.Constant(`App\LIMIT`, 0) == nil || ix.Constant("GLOBAL_FLAG", 0) == nil {
		t.Fatal("constants")
	}
	if ch := ix.Children(`Base\Model`); len(ch) != 1 || ch[0] != `App\User` {
		t.Fatalf("children: %v", ch)
	}
	ix.Remove("b.php")
	if ix.Class(`Base\Model`, 0) != nil || ix.FindMethod(`App\User`, "save", 0) != nil {
		t.Fatal("remove")
	}
}

func TestCyclicHierarchy(t *testing.T) {
	ix := New(nil)
	ix.Add(extract(t, "c.php", `<?php class A extends B {} class B extends A {}`))
	if len(ix.Ancestors("A", 0)) != 2 || len(ix.ParentChain("A", 0)) != 1 {
		t.Fatal("cycle handling")
	}
}

func TestAvailability(t *testing.T) {
	ix := New(nil)
	ix.Add(extract(t, "stub.php", `<?php
#[PhpStormStubsElementAvailable(from: '5.3', to: '7.4')]
function f($a): string|false {}
#[PhpStormStubsElementAvailable(from: '8.0')]
function f($a): string {}
`))
	if f := ix.Function("f", phpver.PHP73); f == nil || f.Return != "false|string" {
		t.Fatalf("7.3: %+v", f)
	}
	if f := ix.Function("f", phpver.PHP84); f == nil || f.Return != "string" {
		t.Fatalf("8.4: %+v", f)
	}
}

func TestTemplates(t *testing.T) {
	ix := New(nil)
	ix.Add(extract(t, "g.php", `<?php
namespace Lib;
/** @template T */
final class Box {
    /**
     * @param T $v
     * @return T
     */
    public function put($v) { return $v; }
    /**
     * @template K
     * @param K $k
     * @return list<K>
     */
    public function wrap($k): array { return [$k]; }
}
`))
	m := ix.FindMethod(`Lib\Box`, "put", 0)
	if m.Params[0].DocType != "mixed" || m.DocReturn != "mixed" {
		t.Fatalf("put: %+v", m)
	}
	if w := ix.FindMethod(`Lib\Box`, "wrap", 0); w.DocReturn != "array<int,mixed>" {
		t.Fatalf("wrap: %+v", w)
	}
}

func TestPolyfillYieldsToBuiltin(t *testing.T) {
	stubs := New(nil)
	stubs.Add(extract(t, "stub.php", `<?php
#[PhpStormStubsElementAvailable(from: '8.3')]
function pad(string $s): string {}
`))
	project := New(stubs)
	project.Add(extract(t, "polyfill.php", `<?php
if (!function_exists('pad')) { function pad($s) { return $s; } }
`))
	file := New(project)
	if f := file.Function("pad", phpver.PHP84); f == nil || f.Return != "string" {
		t.Fatalf("8.4: the builtin must win: %+v", f)
	}
	if f := file.Function("pad", phpver.PHP80); f == nil || f.Return != "" {
		t.Fatalf("8.0: the polyfill must be used: %+v", f)
	}
}
