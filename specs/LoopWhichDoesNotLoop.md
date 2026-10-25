---
id: LoopWhichDoesNotLoop
group: Control flow
kind: semantic
needs: [types]
php: { min: "", max: "" }
---

# LoopWhichDoesNotLoop

## Summary

A loop whose body always leaves on the first pass (its last statement is
`break`, `return` or `throw` and nothing continues it), or whose body is
empty, never actually iterates. It is either a bug or a disguised `if` /
"take first element" idiom.

## Detection

Skipped entirely when the file name ends with `.blade.php`.

Applies to `foreach`, `for`, `while` and `do … while` loops `L`.

- **D1** `L` has a braced body `{ … }`. Loops without a group body
  (`while (x) stmt;`, `for (;;);`) are never reported.
- **D2** Let `last` be the last statement of the body, ignoring comments and
  docblocks. Either the body has no statement at all (`{}` or only comments;
  custos: `foreach` only, see Divergences),
  or `last` is a `break` (any level), a `return` (with or without value) or a
  `throw` statement.
- **D3** No `continue` inside the body continues `L`. For every `continue`
  anywhere in the body, walk up its ancestors (stopping at a function /
  method / closure / arrow-function boundary or the file — a `continue` inside
  a closure never counts); count the loops met (`foreach`, `for`, `while`,
  `do … while`; `switch` is **not** counted), starting at 1 for the innermost
  enclosing loop. When `L` is reached at count `k`, the `continue` continues
  `L` iff its level equals `k`. The level is 1 without argument, otherwise the
  argument's source text parsed as a decimal integer; an argument that does
  not parse (e.g. `continue (2);`, `continue $n;`) counts as 1.
- **D4** `foreach` only: if `last` exists and is not a `throw` (i.e. it is
  `break` or `return`), resolve the type of the iterated expression; if any of
  its known types (unknown parts dropped; union members considered) is
  `iterable`, `\Generator`, `\Traversable`, `\Iterator`, or a class or
  interface that extends/implements `\Traversable` in the index (e.g. a
  class implementing `\IteratorAggregate` or `\Iterator`, an interface
  extending `\Iterator`, `\ArrayIterator`, `\PDOStatement`) → no report
  (taking the first element of a lazy or stateful traversable is a
  legitimate idiom). Array-of types (`\Generator[]`) do not count.
  Empty-body `foreach` loops are not exempt.
- **D5** `foreach` only, `last` is `break` or `return`: the iterated
  expression (parentheses ignored) is a method call `recv->name()` with a
  literal name, and for one of the receiver's known classes the method found
  for `name` (own, inherited or imported from a trait) is a generator: its
  body holds a `yield` / `yield from` of its own (nested closures, functions
  and classes do not count). A method declared in another file is recognized
  through the body-derived return type kept by the project index. Such a
  call returns a `\Generator` even though inference leaves the type of an
  overridable method without return type unknown → no report.

When D1–D3 hold (and neither D4 nor D5 exempts), report.

## Exceptions (no report)

- **E1** Last statement is anything other than `break`/`return`/`throw`
  (including `exit`, a call, an `if` that contains a `break`).
- **E2** A matching `continue` (D3), e.g. `continue;` in an `if` directly in
  the loop, or `continue 2;` in an inner loop of `L` or in a `switch` directly
  in `L`. A plain `continue;` inside a `switch` only leaves the switch and does
  not exempt `L`.
- **E3** `foreach` over a value typed `iterable` or any `\Traversable`
  type (`\Generator`, `\Iterator`, `\IteratorAggregate` implementations, …)
  ending with `break`/`return` (D4).
  Likewise over a call to a method whose body yields, with or without a
  declared return type (D5): `foreach ($this->rows() as $row) { return $row; }`.
- **E4** Brace-less loops; `.blade.php` files.

## Report

- Range: the loop's first keyword token (`foreach`, `for`, `while`, or `do`
  for do-while).
- Severity: warning.
- Message: `Loop body exits on the first iteration; the loop never repeats.`

## Fix

None.

## Options

None.

## PHP versions

No gating.

## Examples

```php
<?php
<warning descr="Loop body exits on the first iteration; the loop never repeats.">foreach</warning> ($queue as $job) {
    process($job);
    break;
}
<warning descr="Loop body exits on the first iteration; the loop never repeats.">for</warning> ($t = 0; $t < 5; $t++) {
    return $t;
}
<warning descr="Loop body exits on the first iteration; the loop never repeats.">while</warning> (poll()) {
    throw new \LogicException('stop');
}
<warning descr="Loop body exits on the first iteration; the loop never repeats.">do</warning> {
    break 1;
} while ($again);
<warning descr="Loop body exits on the first iteration; the loop never repeats.">while</warning> (wait()) {
    // nothing yet
}
<warning descr="Loop body exits on the first iteration; the loop never repeats.">foreach</warning> ($grid as $line) {
    foreach ($line as $cell) {
        while (true) {
            continue 2;        // continues the middle loop, not the outer one
        }
    }
    break;
}

foreach ($grid as $line) {
    foreach ($line as $cell) {
        continue 2;            // continues the outer loop
    }
    break;
}
foreach ($grid as $line) {
    if (!$line) { continue; }
    break;
}
while ($ok) {
    step();
}
/** @var \Iterator $cursor */
foreach ($cursor as $first) {
    break;
}
class Ledger {
    public function entries() { yield 1; yield 2; }
    public function firstEntry() {
        foreach ($this->entries() as $entry) {   // generator method, no return type
            return $entry;
        }
        return null;
    }
}
while ($ok) break;
```

## Divergences

- `continue` inside `switch` (custos diverges from upstream). Upstream does
  not count `switch` as a level, so a bare `continue;` in a `switch` inside
  the loop is taken as continuing the loop and the loop is not reported. PHP
  counts `switch` as a loop structure: that `continue;` acts like `break` of
  the switch, so the loop still exits on its first pass. custos counts
  `switch` (D3): it reports such loops, and `continue 2;` in the switch
  exempts them.
- Empty braced bodies are reported (D2); this is upstream behavior, kept.
- Traversable classes (custos diverges from upstream). Upstream exempts only
  the four exact type names `\Generator`, `\Traversable`, `\Iterator` and
  `iterable`, so taking the first element of a user collection implementing
  `\IteratorAggregate` (or of `\ArrayIterator`, `\PDOStatement`, …) is
  reported although it is the same legitimate idiom. custos exempts every
  type that is a subtype of `\Traversable` (D4).
- Alternative syntax (`foreach (…): … endforeach;`): upstream behavior
  depends on whether the parser exposes a group body. custos treats the
  statement list as the body.
- **Empty `while`/`do`/`for` bodies (custos diverges).** Upstream reports
  any empty loop body, but `while (@ob_end_flush()) {}` or
  `for (; next($a) !== false;) {}` repeat as long as the condition holds —
  the condition does the work. custos reports an empty body only for
  `foreach`.
- **Properties of unknown type (custos diverges).** A `foreach` over a
  property whose type is unknown, with a body ending in `return` or
  `break`, is not reported: such a property is usually a collection its
  class fills elsewhere (an untyped Doctrine relation, `@ORM\ManyToMany`),
  and returning its first element is the point. A property known to hold
  an array is still reported.
- **Empty bodies over objects (custos diverges).** An empty `foreach` body
  is still reported over known arrays, but not when the subject's type is
  unknown, `object`, `iterable`, `mixed` or a class: iterating may be done
  for the iterator's side effects (Doctrine tests initialize lazy
  collections with `foreach ($user->groups as $g) {}`).
- **Empty bodies keeping the last pair (custos diverges).** An empty
  `foreach` body whose key or value goes to a property, an array element
  or a `list()` (`foreach ($pairs as $this->name => $this->value) {}`), or
  to a variable read after the loop, keeps the last key or value on
  purpose: it is not reported.
- **`$this` in a trait (custos diverges, D4).** `foreach ($this as $el) {
  return false; }` inside a trait method tests the using class (a
  Traversable collection, CakePHP's `CollectionTrait::isEmpty()`) for
  emptiness; `$this` has no class type in a trait, so the Traversable
  exemption cannot apply: such loops are not reported.
- **Generator methods without return type (custos diverges, D5).** Upstream
  only looks at the resolved type of the iterated expression, which is
  unknown for an untyped method that may be overridden. custos also looks at
  the called method itself: when its body yields, the loop takes the first
  element of a generator and is not reported. Limits: trait and private
  generator methods are only recognized when declared in the analyzed file.
