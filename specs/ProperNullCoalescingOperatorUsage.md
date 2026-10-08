---
id: ProperNullCoalescingOperatorUsage
group: Code style
kind: semantic
needs: [names, types, hierarchy, index]
php: { min: "7.0", max: "" }
---

# ProperNullCoalescingOperatorUsage

## Summary

Two misuses of `??`: (a) `call() ?? null` — a call result never triggers an
undefined-index/variable notice, so `?? null` does nothing; (b) a fallback
whose type has nothing in common with the left operand's type, which usually
signals a bug or a confusing API.

## Detection

Applies to a binary `??` expression `L ?? R`.

Pre-conditions (all cases):

- P1: the expression's direct parent is not another `??` binary expression
  (in either operand position). A `??` wrapped in parentheses inside another
  `??` is still inspected (its direct parent is the parentheses).
- P2: the expression is not the direct content of parentheses that are the
  operand of a type cast, i.e. not `(type) (L ?? R)` (any cast: `(string)`,
  `(int)`, `(bool)`, `(array)`, `(object)`, `(float)`, `(unset)`, …).

Case A — useless null fallback:

- D1: `R` is the `null` constant (case-insensitive). If additionally `L` is a
  function call, method call (instance, nullsafe, or static) or a call on a
  callable variable (`$fn()`), report. If `R` is `null` and `L` is anything
  else, nothing is reported and Case B is not evaluated.

Case B — non-complementary types (only when `ANALYZE_TYPES` is true and `R` is
not `null`):

- D2: the expression is inside a function, method, closure or arrow function
  body (not at file/class level).
- D3: resolve the type sets of `L` and `R`. A set is unusable (→ no report)
  when the inferred type is missing, contains an unknown/unresolved part, or
  contains `mixed` or `object`. From a usable set remove `null` and `static`
  (`$this` counts as `static`); if it is then empty → no report.
  Types are normalised: `integer`→`int`, `boolean`/`true`/`false`→`bool`,
  any `T[]`→`array`, `Closure`→`callable`, `iterable`→ both `array` and
  `\Traversable` (custos diverges, see Divergences); class types are
  fully-qualified names.
- D4: complementary check:
  - `ALLOW_OVERLAPPING_TYPES` true: complementary when at least one type of
    `R` is in `L`'s set.
  - `ALLOW_OVERLAPPING_TYPES` false: complementary when `R`'s set contains all
    of `L`'s types.
- D5: when not complementary, check relatedness of class types: take the class
  types (fully-qualified names) of each side, expand each class/interface to
  its full inheritance closure (itself, all parent classes, all implemented
  interfaces recursively, all used traits recursively). If the two closures
  share at least one element, the sides are related → no report.
- D5a: scalar fallback (custos diverges). When every type left in `L`'s set
  and every type left in `R`'s set is a scalar (`int`, `float`, `string`,
  `bool`), the sides are compatible → no report. A placeholder of another
  scalar type (`$id ?? 'new'`, `$count ?? '?'`) is a deliberate display or
  serialisation choice, not a bug. Applies under both values of
  `ALLOW_OVERLAPPING_TYPES`. A side holding a scalar *and* a non-scalar
  (`int|Engine`) is not covered by D5a.
- D5b: iterable family (custos diverges). A type belongs to the iterable
  family when it is `array`, `\Traversable`, or a class/interface whose
  inheritance closure (as in D5) contains `\Traversable` (`\Iterator`,
  `\IteratorAggregate`, `\Generator`, `\ArrayObject`, `\DOMNamedNodeMap`,
  `\DOMNodeList`, a Doctrine `Collection`, any user class implementing one
  of those). When `L`'s set and `R`'s set each contain at least one member
  of the family, the sides are related → no report. The typical case is an
  empty-array fallback for an iterable: `$node->attributes ?? []`,
  `$owner->items ?? []` with `Collection $items`.
- D6: not complementary, not related, and neither D5a nor D5b applies →
  report.

## Exceptions (no report)

- E1: nested `??` chains: inner `??` nodes are skipped (P1).
- E2: `??` that is directly cast, e.g. `(string) ($a ?? $b)` (P2).
- E3: `$var ?? null`, `$arr['k'] ?? null`, `$obj->prop ?? null` (left is not a
  call).
- E4: any type set unknown/partial, or containing `mixed`/`object`; or only
  `null` after resolution.
- E5: class types in an inheritance relationship (`Child ?? Base`, or both
  implementing a shared interface).
- E6: code outside any function body (for Case B).
- E7: both sides scalar-only after removing `null` (D5a), e.g.
  `sprintf('#%s', $row->id ?? 'new')` with `?int $id`.
- E8: both sides in the iterable family (D5b), e.g.
  `iterator_to_array($el->attributes ?? [])` with `?\DOMNamedNodeMap`, or
  `foreach ($order->lines ?? [] as $l)` where `$order` is nullable and
  `lines` is a `Collection`.

## Report

- Range: the whole `L ?? R` binary expression.
- Severity: info (weak warning).
- Message (Case A): "'{L}' alone is equivalent; drop the '?? null' fallback."
- Message (Case B): "Operand types of '??' do not match ({left types} vs {right types})."

## Fix

- F1 (Case A only): replace the whole `L ?? R` expression with the text of `L`.
- Case B has no fix.

## Options

| Option | Type | Default | Effect |
|---|---|---|---|
| ANALYZE_TYPES | bool | true | Enables Case B (type complementarity). |
| ALLOW_OVERLAPPING_TYPES | bool | true | When true, any shared type makes operands complementary; when false, the right operand's types must cover all of the left operand's types. |

Upstream fixtures run with `ANALYZE_TYPES` = true and
`ALLOW_OVERLAPPING_TYPES` = false.

## PHP versions

The `??` operator exists since PHP 7.0; nothing to report before.

## Examples

With `ALLOW_OVERLAPPING_TYPES` = false:

```php
<?php
interface Shape {}
class Square implements Shape {}
class Disc implements Shape {}
class Engine {}
class Turbo extends Engine {}

class Garage {
    /** @var Engine|null */
    private $engine;

    public function lookup(): ?Engine { return null; }

    public function demo(Engine $e, Turbo $t, Square $s, Disc $d, ?int $n = null, string $label = null) {
        return [
            <weak_warning descr="'$this->lookup()' alone is equivalent; drop the '?? null' fallback.">$this->lookup() ?? null</weak_warning>,
            <weak_warning descr="'strrev($label)' alone is equivalent; drop the '?? null' fallback.">strrev($label) ?? NULL</weak_warning>,
            <weak_warning descr="Operand types of '??' do not match ([\Engine] vs [string]).">$this->engine ?? 'none'</weak_warning>,
            $n ?? 'n/a',                    // scalar placeholder (D5a)
            $this->engine ?? null,
            $label ?? null,
            $t ?? $e,
            $e ?? $t,
            $s ?? $d,                       // related through Shape
            $n ?? 42,
            (int) ($label ?? $e),
        ];
    }
}
```

```php
<?php
interface Shape {}
class Square implements Shape {}
class Disc implements Shape {}
class Engine {}
class Turbo extends Engine {}

class Garage {
    /** @var Engine|null */
    private $engine;

    public function lookup(): ?Engine { return null; }

    public function demo(Engine $e, Turbo $t, Square $s, Disc $d, ?int $n = null, string $label = null) {
        return [
            $this->lookup(),
            strrev($label),
            $this->engine ?? 'none',
            $n ?? 'n/a',                    // scalar placeholder (D5a)
            $this->engine ?? null,
            $label ?? null,
            $t ?? $e,
            $e ?? $t,
            $s ?? $d,                       // related through Shape
            $n ?? 42,
            (int) ($label ?? $e),
        ];
    }
}
```

Iterable fallbacks (D5b) and the shapes produced by
`NullCoalescingOperatorCanBeUsed`'s fix are not reported:

```php
<?php
class Shelf implements \IteratorAggregate {
    public function getIterator(): \Iterator { return new \ArrayIterator([]); }
}
class Room {
    public Shelf $shelf;
}
class Inventory {
    private ?int $cap = null;

    public function demo(?Room $room, ?\DOMElement $el, ?\Iterator $it, ?float $ratio) {
        foreach ($room?->shelf ?? [] as $item) {}
        $attrs = iterator_to_array($el?->attributes ?? []);
        $rows = $it ?? [];
        $shown = $ratio ?? '-';
        return [$this->cap ?? 'none', $attrs, $rows, $shown];
    }
}
```

## Divergences

- **Scalar fallbacks (custos diverges, D5a).** Upstream reports any `??`
  whose scalar operand types differ, e.g. a nullable `int` with a string
  placeholder. That fallback is ordinary code (labels, log lines, array rows
  for `implode()`), and custos's own `NullCoalescingOperatorCanBeUsed` fix
  produces it from `null === $x ? 'none' : $x`; custos stays silent when
  both sides are scalar-only.
- **Iterable family (custos diverges, D5b).** Upstream compares an
  object type with `array` as unrelated, so `$node->attributes ?? []` on a
  `?DOMNamedNodeMap`, or `$owner->items ?? []` on a `Collection`, is
  reported although both operands can be iterated the same way (and the
  latter is what `NullCoalescingOperatorCanBeUsed` writes for
  `$owner ? $owner->items : []`). custos treats `array`, `\Traversable` and
  every `\Traversable` implementor as related. Class-vs-scalar pairs
  (`$engine ?? 'none'`) remain reported: the ternary that a
  `NullCoalescingOperatorCanBeUsed` fix would turn into one mixes the same
  types, so the finding is about the original code, not the rewrite.

- **`iterable` (custos diverges):** upstream compares `iterable` as an
  opaque type name, so `$this->rows() ?? []` on an `iterable`-returning
  method is reported although an array is an iterable. custos expands
  `iterable` to `array` and `\Traversable` before the complementary and
  relatedness checks (D3).
- Upstream's type inference (IDE engine) decides which Case B reports appear;
  our inference must give up (no report) whenever any part of a type is
  unknown, so we may report less often than upstream, never with a guess.
  Expressions such as `time() + SOME_CONSTANT` must resolve to `int` for the
  upstream fixture's false-positive cases to stay silent; if constant types
  cannot be inferred, the set is unknown and no report is made either way.
- **Unresolved classes (D3, implementation fix).** A class type that does
  not resolve in the project index (`$h ?? new Vendor\Missing\Handler()`
  with the vendor tree absent) makes the set unusable, as D3 requires;
  custos used to compare it as an unrelated class and report.
- **Invokable classes (custos diverges).** Upstream treats `callable` and a
  class as unrelated, so `?callable $h ... $h ?? new Dispatcher()` is
  reported although `Dispatcher` declares `__invoke()` and is a valid
  callable. custos treats a class declaring (or inheriting) `__invoke()` as
  related to `callable` (D5).
