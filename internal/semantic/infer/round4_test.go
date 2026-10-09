package infer_test

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
	"custos/internal/semantic/names"
	"custos/internal/testing/testbudget"
)

// A native `iterable` is refined by its doc type, as `array` is.
func TestIterableRefinedByDoc(t *testing.T) {
	checkAnywhere(t, `<?php
class Foo {}
/** @template T */
class Coll implements \IteratorAggregate { public function getIterator(): \Iterator {} }
class H {
    /** @var array<string, Foo> */
    private iterable $p = [];
    /**
     * @param Foo[] $b
     * @param iterable<string, Foo> $c
     */
    public function f(iterable $b, ?iterable $c) {
        foreach ($b as $x) { t('b', $x); }
        foreach ($c as $k => $x) { t('c', $x); t('ck', $k); }
        t('prop', $this->p);
        t('ret', $this->g());
    }
    /** @return Coll<Foo> */
    public function g(): iterable {}
}
`, map[string]string{"b": `\Foo`, "c": `\Foo`, "ck": "string", "prop": `\Foo[]`, "ret": `\Coll<\Foo>`})
}

// Fewer generic arguments than templates bind the templates without bound
// or default (Shopware's `@template TElement` / `@template TKey of
// array-key = array-key`); unbound ones take their default.
func TestPartialGenericArguments(t *testing.T) {
	checkAnywhere(t, `<?php
class Item {}
/**
 * @template TElement
 * @template TKey of array-key = array-key
 * @implements \IteratorAggregate<TKey, TElement>
 */
class Collection implements \IteratorAggregate {
    /** @return \Traversable<TKey, TElement> */
    public function getIterator(): \Traversable {}
    /** @return TKey */
    public function key() {}
}
/** @extends Collection<Item> */
class ItemCollection extends Collection {}
/**
 * @template TKey of array-key
 * @template TValue
 */
class Map { /** @return TValue */ public function get() {} /** @return TKey */ public function k() {} }
/**
 * @template K of array-key
 * @template V of object
 */
class Bounded { /** @return V */ public function v() {} }
/**
 * @template A
 * @template B = int
 */
class Pair { /** @return B */ public function second() {} }
/**
 * @param Map<Item> $m
 * @param Pair<string> $p
 * @param \Generator<Item> $g
 * @param Bounded<Item> $bd
 */
function f(ItemCollection $c, Map $m, Pair $p, \Generator $g, Bounded $bd) {
    t('bounded', $bd->v());
    foreach ($c as $key => $v) { t('key', $key); t('val', $v); }
    t('mapGet', $m->get());
    t('mapKey', $m->k());
    t('pairDefault', $p->second());
    foreach ($g as $x) { t('gen', $x); }
}
`, map[string]string{
		"key": "int|string", "val": `\Item`, "mapGet": `\Item`, "mapKey": "mixed", "pairDefault": "int", "gen": `\Item`, "bounded": `\Item`,
	})
}

// `$isObject = is_object($r); if ($isObject) { … }` narrows $r, unless
// either variable changed in between.
func TestBooleanAliasNarrowing(t *testing.T) {
	checkAnywhere(t, `<?php
function f(object|string|null $r, ?int $n, $c) {
    if (($isObject = is_object($r)) && $c) {}
    if ($isObject) { t('true', $r); } else { t('false', $r); }
    $has = $n !== null;
    if (!$has) { return; }
    t('guard', $n);
    $ok = is_int($n);
    $n = $c ? null : 1;
    if ($ok) { t('changed', $n); }
    $flag = is_string($r);
    if ($c) { $flag = true; }
    if ($flag) { t('twoDefs', $r); }
    $copy = $has;
    if ($copy) { t('aliasOfAlias', $n); }
    $call = strlen('x');
    if ($call) { t('notBool', $r); }
    if ($undefined) { t('undefinedAlias', $r); }
    t('replaceUnknown', str_replace('a', 'b', $c));
}
`, map[string]string{
		"true": "object", "false": "null|string", "guard": "int", "changed": "int|null",
		"twoDefs": "null|object|string", "aliasOfAlias": "int|null", "notBool": "null|object|string",
		"undefinedAlias": "null|object|string", "replaceUnknown": "?unknown",
	})
}

// Properties of plain variables narrow like `$this->p` (conditions,
// continue/return guards), and lose the fact when the variable changes.
func TestVariablePropertyNarrowing(t *testing.T) {
	checkAnywhere(t, `<?php
class Ref {
    public string|array $value = '';
    public ?Ref $next = null;
    /** @var array<string, int|string> */
    public array $list = [];
}
class H {
    private ?Ref $p = null;
    public function f(array $m, Ref $r, Ref $s) {
        foreach ($m as $ref) {
            /** @var Ref $ref */
            if (!is_string($ref->value)) continue;
            t('continue', str_replace('_', '-', $ref->value));
        }
        foreach ($m as $x) {
            if ($this->p === null) { continue; }
            t('thisProp', $this->p);
        }
        if ($r->next !== null) {
            t('cond', $r->next);
            $r = new Ref();
            t('reassigned', $r->next);
        }
        if ($s->next === null) { return; }
        t('guard', $s->next);
        $s = new Ref();
        t('afterReset', $s->next);
        if (isset($r->list['k'])) { t('dim', $r->list['k']); $r = new Ref(); t('dimReset', $r->list); }
        if (is_int($s->list['n'])) { $s = $r; $v = $s->list['n']; t('dimAfter', $v); }
        if ($r->next instanceof Ref && is_string($r->next->value)) {
            t('chain', $r->next->value);
            $r->next = new Ref();
            t('chainReset', $r->next->value);
        }
        if (is_string($this->p->value)) { t('thisChain', $this->p->value); }
    }
}
`, map[string]string{
		"continue": "string", "thisProp": `\Ref`, "cond": `\Ref`, "reassigned": `\Ref|null`, "guard": `\Ref`,
		"afterReset": `\Ref|null`, "dim": "int|string", "dimReset": "int[]|string[]", "dimAfter": "int|string",
		"chain": "string", "chainReset": "array|string", "thisChain": "string",
	})
}

// A method without return type keeps the return type of the method it
// overrides; declared `: mixed` is never replaced by the body.
func TestInheritedReturnSignature(t *testing.T) {
	checkAnywhere(t, `<?php
interface Source { public function value(): mixed; }
abstract class Base { /** @return int */ abstract public function count(); }
final class Impl extends Base implements Source {
    public function value() { return 1.5; }
    public function count() { return 'x'; }
    public function own() { return 2; }
    public function m(): mixed { return [1]; }
}
final class J implements \JsonSerializable { public function jsonSerialize() { return [1]; } }
function f(Impl $i, J $j) {
    t('value', $i->value());
    t('count', $i->count());
    t('own', $i->own());
    t('mixed', $i->m());
    t('json', $j->jsonSerialize());
}
`, map[string]string{"value": "mixed", "count": "int", "own": "int", "mixed": "mixed", "json": "mixed"})
}

// Keys absent from a literal get no type from the other elements, also
// after writes to other keys or nested writes.
func TestAbsentKeysAfterWrites(t *testing.T) {
	checkAnywhere(t, `<?php
function f(int $x, string $k) {
    $rows = ['total' => $x];
    $rows['meta']['s'] = 1;
    t('nested', $rows['meta']);
    $w = ['a' => 1];
    $w['b'] = 'x';
    t('otherKey', $w['c']);
    t('written', $w['b']);
    $l = [1, 2];
    $l[] = 3;
    t('append', $l[7]);
    $l2 = ['a' => 1];
    $l2[] = 3;
    t('appendStr', $l2['z']);
    $c = ['a' => 1];
    $c[$k] = 'v';
    t('computed', $c['q']);
    $d = ['a' => 1];
    [$d['b']] = [2];
    t('destructured', $d['b']);
}
`, map[string]string{
		"nested": "?unknown", "otherKey": "?unknown", "written": "string", "append": "?unknown",
		"appendStr": "?unknown", "computed": "string", "destructured": "?unknown",
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

// Promoted properties take their doc type from the constructor's @param.
func TestPromotedPropertyDoc(t *testing.T) {
	checkAnywhere(t, `<?php
class Foo {}
class H {
    /** @param Foo[] $items */
    public function __construct(private array $items, private int $n = 0) {}
    public function f() { t('items', $this->items); t('n', $this->n); }
}
`, map[string]string{"items": `\Foo[]`, "n": "int"})
}

// Many variable-property guards and alias reads stay linear.
func TestRound4Bounded(t *testing.T) {
	var b strings.Builder
	b.WriteString("<?php\nclass R { public ?R $n = null; }\nfunction f(R $r, ?int $x) {\n")
	for i := 0; i < 20000; i++ {
		fmt.Fprintf(&b, " if ($r->n === null) { return; } $y = $r->n; $ok%d = $x !== null; if ($ok%d) { $z = $x; }\n", i%50, i%50)
	}
	b.WriteString("}\n")
	if d := typeAll(t, b.String()); d > testbudget.Of(2*time.Second) && !raceEnabled {
		t.Errorf("took %v", d)
	}
}
