package infer_test

import (
	"fmt"
	"strings"
	"testing"

	"custos/internal/php/syntax"
)

func TestMixinMembers(t *testing.T) {
	checkAnywhere(t, `<?php
class Query {
    public int $limit;
    public function count(): int {}
    /** @return static */
    public function where() {}
    /** @return $this */
    public function orWhere() {}
}
class Deep { public function deep(): string {} }
/** @mixin Deep */
class Middle {}
/**
 * @mixin Query
 * @mixin Middle
 */
class Builder {
    public function own(): float {}
    public function __call($n, $a) {}
}
class Child extends Builder {}
/** @mixin LoopB */
class LoopA {}
/** @mixin LoopA */
class LoopB {}
function f(Builder $b, Child $c, LoopA $l) {
    t('method', $b->count());
    t('property', $b->limit);
    t('own', $b->own());
    t('fluent', $b->where());
    t('fluentThis', $b->orWhere());
    t('inherited', $c->count());
    t('chained', $b->deep());
    t('missing', $b->nothing());
    t('missingProp', $b->nothing);
    t('cycle', $l->nothing());
}
`, map[string]string{
		"method": "int", "property": "int", "own": "float", "fluent": `\Builder`, "fluentThis": `\Builder`,
		"inherited": "int", "chained": "string", "missing": "?unknown", "missingProp": "?unknown", "cycle": "?unknown",
	})
}

func TestTemplateMixins(t *testing.T) {
	checkAnywhere(t, `<?php
class Query { public int $limit; public function count(): int {} }
class Other { public function name(): string {} }
/**
 * @template T
 * @mixin T
 */
class Proxy {}
/**
 * @template T of Query
 * @mixin T
 */
class Bounded {}
/** @extends Proxy<Query> */
class QueryProxy extends Proxy {}
/**
 * @param Proxy<Query> $bound
 * @param Proxy<Query|Other> $union
 * @param Proxy<int> $scalar
 * @param Bounded<Query> $viaArg
 */
function f($bound, $union, $scalar, Proxy $raw, Bounded $viaBound, $viaArg, QueryProxy $inherited) {
    t('method', $bound->count());
    t('property', $bound->limit);
    t('unionFirst', $union->count());
    t('unionSecond', $union->name());
    t('scalar', $scalar->count());
    t('raw', $raw->count());
    t('viaBound', $viaBound->count());
    t('viaArg', $viaArg->count());
    t('inherited', $inherited->count());
}
`, map[string]string{
		"method": "int", "property": "int", "unionFirst": "int", "unionSecond": "string", "scalar": "?unknown",
		"raw": "?unknown", "viaBound": "int", "viaArg": "int", "inherited": "int",
	})
}

// A chain of mixins longer than the lookup budget ends the search.
func TestMixinLookupIsBounded(t *testing.T) {
	var src strings.Builder
	src.WriteString("<?php\n")
	const links = 40
	for i := range links {
		fmt.Fprintf(&src, "/** @mixin C%d */\nclass C%d {}\n", i+1, i)
	}
	fmt.Fprintf(&src, "class C%d { public function far(): int {} }\n", links)
	src.WriteString("function f(C38 $near, C0 $start) {\n    t('near', $near->far());\n    t('far', $start->far());\n}\n")
	checkAnywhere(t, src.String(), map[string]string{"near": "int", "far": "?unknown"})
}

// PHPDoc alone documents a mixin or a receiver's new type.
func TestMixinAndSelfOutStayOutOfNativeTypes(t *testing.T) {
	src := `<?php
class Query { public int $limit; public function count(): int {} }
/** @mixin Query */
class Builder {}
class Open {}
class Draft {
    /** @phpstan-self-out Open */
    public function open(): void {}
}
function f(Builder $b, Draft $d) {
    $n = $b->count();
    $p = $b->limit;
    $d->open();
    return $d;
}
`
	env, file := envFor(t, src, nil)
	native := env.Native()
	syntax.InspectFile(file, func(n syntax.Node) bool {
		switch x := n.(type) {
		case *syntax.Assign:
			if got := native.TypeOf(x.Value); !got.IsUnknown() {
				t.Errorf("native type of a mixin member: %s", got)
			}
		case *syntax.Return:
			if got := native.TypeOf(x.Expr).String(); got != `\Draft` {
				t.Errorf("native receiver after the call: %s", got)
			}
		}
		return true
	})
}

func TestSelfOutRebindsAProperty(t *testing.T) {
	checkAnywhere(t, `<?php
class Open { public function id(): int {} }
class Draft {
    /** @phpstan-self-out Open */
    public function open(): void {}
    /** @phpstan-self-out mixed */
    public function vague(): void {}
    public function plain(): void {}
}
class Holder {
    public Draft $state;
    public ?Draft $maybe;
    public Draft $other;
    public Draft $third;
    public Draft $fourth;
    public function run(Holder $h, string $name) {
        $this->state->open();
        t('after', $this->state);
        t('chain', $this->state->id());
        $this->other->vague();
        t('vague', $this->other);
        $this->third->plain();
        t('plain', $this->third);
        $this->maybe?->open();
        t('nullsafe', $this->maybe);
        $h->fourth->open();
        t('otherObject', $h->fourth);
        $this->fourth->{$name}();
        t('dynamic', $this->fourth);
    }
}
`, map[string]string{
		"after": `\Open`, "chain": "int", "vague": `\Draft`, "plain": `\Draft`, "nullsafe": `\Draft|null`,
		"otherObject": `\Open`, "dynamic": `\Draft`,
	})
}

func TestSelfOutRebindsTheReceiver(t *testing.T) {
	checkAnywhere(t, `<?php
class Open { public function id(): int {} }
class Draft {
    /** @phpstan-self-out Open */
    public function open(): void {}
    /** @phpstan-self-out mixed */
    public function vague(): void {}
    /** @phpstan-self-out static */
    public function same(): void {}
    public function plain(): void {}
}
class Bare {}
class Other {
    /** @phpstan-self-out Open */
    public function open(): void {}
    public function vague(): void {}
}
function f(Draft $d, Draft $e, Draft $g, Draft $h, Draft $i, ?Draft $n, Draft|Other $u, Draft|Other $w, $x, bool $c, Draft|Bare $m, Draft $a) {
    t('before', $d);
    $d->open();
    t('after', $d);
    t('chain', $d->id());
    $e->vague();
    t('vague', $e);
    $g->same();
    t('same', $g);
    $h->plain();
    t('plain', $h);
    if ($c) { $i->open(); }
    t('branch', $i);
    $n?->open();
    t('nullsafe', $n);
    $u->open();
    t('union', $u);
    $w->vague();
    t('partly', $w);
    $x->open();
    t('unknown', $x);
    $m->open();
    t('noMethod', $m);
    $run = fn() => [$a->open(), t('inArrow', $a)];
    t('outsideArrow', $a);
}
`, map[string]string{
		"before": `\Draft`, "after": `\Open`, "chain": "int", "vague": `\Draft`, "same": `\Draft`, "plain": `\Draft`,
		"branch": `\Draft|\Open`, "nullsafe": `\Draft|\Open|null`, "union": `\Open`, "partly": `\Draft|\Other`, "unknown": "?unknown",
		"noMethod": `\Bare|\Draft`, "inArrow": `\Open`, "outsideArrow": `\Draft`,
	})
}
