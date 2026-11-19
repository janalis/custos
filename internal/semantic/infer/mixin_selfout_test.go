package infer_test

import (
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
