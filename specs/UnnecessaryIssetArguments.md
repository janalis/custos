---
id: UnnecessaryIssetArguments
group: Unused
kind: syntax
needs: []
php: { min: "", max: "" }
---

# UnnecessaryIssetArguments

## Summary

`isset($m['a']['b'])` already implies that `$m` and `$m['a']` are set, so
listing those containers as extra `isset()` arguments next to the deeper
access is redundant. Drop them.

## Detection

Visit every `isset(...)` construct.

- **D1** It has two or more arguments.
- **D2** For each argument `C` (in source order) that is an array access
  expression (`x[k]`, also the legacy `x{k}` form) and has not itself been
  reported yet in this `isset`:
  - build the list of *bases* of `C` by repeatedly stepping to the accessed
    container while the current node is an array access: for `$m['a']['b']`
    the bases are `$m['a']`, then `$m`; for `$o->list[3]` the only base is
    `$o->list` (the walk stops at the first non-array-access node and does
    not descend into property fetches, calls, etc.).
  - for each base `B` (innermost-first order as above) and each other
    argument `M` of the same `isset` (`M` is not `C` and not yet reported):
    if `M` is *equivalent* to `B`, report `M` and mark it reported.
- **D3** *Equivalence*: same node kind and
  - for two simple variables: same name (`$m` ≡ `$m`; case-sensitive);
  - otherwise structurally identical ignoring whitespace/comments, or
    identical source text.
  Parentheses are not stripped (`($m)` is not equivalent to `$m`).

- **D4** *Pure bases only*: a base whose evaluation may do more than read a
  value — it contains a function, method or static call, `new`, `clone`, an
  assignment, `++`/`--`, or a similar effectful construct anywhere inside
  it — is skipped. Two occurrences of such an expression are evaluated
  separately and may designate different elements, so neither covers the
  other. The walk still continues to the outer, pure bases
  (`isset($m[next()]['k'], $m)` reports `$m`).

- **D5** *Evaluation order*: `isset` stops at its first unset argument. A
  candidate `M` written **before** the covering access `C` is skipped when
  any argument after `M` up to and including `C` is effectful in the sense
  of D4: without `M`, those arguments are evaluated even when `M` is unset.
  A candidate written after `C`, or covered by another access for which the
  condition holds, is still reported.

Each argument is reported at most once per `isset`, independent of how many
deeper accesses imply it.

- **Name case.** Wherever this rule compares two expressions for
  equivalence, the names PHP resolves case-insensitively — function and
  method names, class names in calls, `new`, `instanceof` and `::`
  accesses, and keywords — compare case-insensitively (`Cache::$map['k']`
  matches `cache::$map['k']`, `$o->Name()` matches `$o->name()`); variable,
  property and constant names stay case-sensitive.

## Exceptions (no report)

- **E1** Single-argument `isset`.
- **E2** Arguments that are not a base of another argument's array-access
  chain: unrelated variables, deeper accesses (`$m['a']['b']` is never
  reported because of `$m['a']`), the object of a property fetch
  (`isset($o, $o->list[3])` does not report `$o`).
- **E3** Duplicate identical arguments (`isset($m['a'], $m['a'])`): an
  argument is never its own base, so nothing is reported.

- **E4** Arguments matching an effectful base (D4):
  `isset($m[$i++], $m[$i++]['k'])` and `isset($m[key()], $m[key()]['k'])`
  report nothing, since removing the shorter argument would drop one
  evaluation of `$i++` / `key()` and may change the result.

- **E5** A pure argument followed by an effectful covering access (D5):
  `isset($m, $m[next()]['k'])` and `isset($m['a'], $m['a'][$i++])` report
  nothing, since `next()` / `$i++` do not run today when `$m` / `$m['a']`
  is unset. `isset($m['a'][$i++], $m['a'])` reports `$m['a']`.

## Report

- Range: the redundant argument expression exactly (e.g. `$m`, `$m['a']`).
- Severity: info (rendered as an "unused" highlight; fixture markup
  `weak_warning`).
- Message: `Redundant isset() argument: a deeper array access already covers it.`

## Fix

All fixes are applied one after another on the same `isset`.

- **F1** Argument is **not** the last argument: delete from the start of the
  argument through the next `,` that follows it; if the character run right
  after that comma is whitespace, delete that whitespace run too.
  `isset($m, $m['a'])` → `isset($m['a'])`.
- **F2** Argument **is** the last argument: delete from the nearest
  preceding `,` through the end of the argument (any whitespace between the
  comma and the argument goes too; whitespace before the comma stays).
  `isset($m['a']['b'], $m['a'])` → `isset($m['a']['b'])`.
- Multi-line layouts collapse naturally:

  ```text
  isset(
      $m,
      $m['a']['b'],
      $m['a']
  )
  ```

  becomes

  ```text
  isset(
      $m['a']['b']
  )
  ```

## Options

None.

## PHP versions

None.

## Examples

```php
<?php
function has_city(array $order, $customer)
{
    $a = isset(<weak_warning descr="Redundant isset() argument: a deeper array access already covers it.">$order</weak_warning>, $order['ship']['city']);
    $b = isset($order['ship']['city'], <weak_warning descr="Redundant isset() argument: a deeper array access already covers it.">$order['ship']</weak_warning>);
    $c = isset(
        <weak_warning descr="Redundant isset() argument: a deeper array access already covers it.">$customer->tags</weak_warning>,
        $customer->tags[0],
        $customer
    );
    $d = isset($order['ship'], $order['ship']);
    $e = isset($order['bill']['zip'], $order['ship']);
    $f = isset($order);
    return [$a, $b, $c, $d, $e, $f];
}
```

```php
<?php
function has_city(array $order, $customer)
{
    $a = isset($order['ship']['city']);
    $b = isset($order['ship']['city']);
    $c = isset(
        $customer->tags[0],
        $customer
    );
    $d = isset($order['ship'], $order['ship']);
    $e = isset($order['bill']['zip'], $order['ship']);
    $f = isset($order);
    return [$a, $b, $c, $d, $e, $f];
}
```

## Divergences

- With a trailing comma after the last argument (PHP 7.3+), F2 leaves the
  trailing comma in place (`isset($m['a'], $m,)` → `isset($m['a'],)`).
  custos behaves the same; no fixture.
- **Name case (custos diverges).** Upstream compares the expressions
  textually, so operands that differ only in the case of a function, method
  or class name (`Stats::$n` vs `stats::$n`), which PHP treats as the same,
  are not recognized as equivalent. custos folds the case of those names
  (Detection, "Name case").
- **Effectful bases (custos diverges).** Upstream treats two textually equal
  accesses as the same value even when their index or container has side
  effects, and its fix then removes one of the evaluations. custos does not
  report such arguments (D4, E4), nor a pure argument whose removal would let
  a later effectful argument run when it did not before (D5, E5).
