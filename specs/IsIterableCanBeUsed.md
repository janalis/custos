---
id: IsIterableCanBeUsed
group: Language level migration
kind: syntax
needs: []
php: { min: "7.1", max: "" }
---

# IsIterableCanBeUsed

## Summary

`is_array($v) || $v instanceof Traversable` is exactly what `is_iterable($v)`
checks. On a language level where `is_iterable()` is available, the one call
is shorter and makes the intent obvious.

## Detection

Visit every function call `F`.

- **D1** `F` is a call to the global `is_array()`: name compared
  case-insensitively (`Is_Array` matches), written unqualified or
  `\`-qualified, and resolving to the built-in (`Foo\is_array(...)`, a
  `use function` import from another namespace, or an unqualified call in a
  namespace declaring its own `is_array` do not match).
- **D2** `F` has exactly one argument `A`.
- **D3** `F`'s **direct** parent is a binary `||` expression (the symbolic
  `||` only; the keyword `or` does not count). A parenthesised call
  (`(is_array($v)) || …`) has a parenthesised-expression parent and does not
  qualify.
- **D4** Find the top of the `||` chain: starting from that parent, repeatedly
  move to the enclosing expression, skipping any number of parentheses, as long
  as the enclosing expression is itself a `||` binary. Call the top `C`.
- **D5** Flatten `C` into fragments: for each operand of a `||` node, strip
  surrounding parentheses; if the result is a `||` binary, recurse into it,
  otherwise it is a fragment (a binary with another operator — `&&`, `and`,
  comparison, `instanceof`, … — is a single fragment, not entered).
- **D6** Among the fragments (in source order), look for one that is an
  `instanceof` binary whose right operand is a class reference whose short name
  resolving (namespace and `use` imports applied, compared case-insensitively)
  to the global interface `\Traversable` — `\Traversable`, `\TRAVERSABLE`, an
  import alias of it, or `Traversable` in the global namespace match;
  `Some\Traversable`, or a bare `Traversable` inside a namespace without an
  import, do not — and whose left operand is structurally
  equivalent to `A` (same node kind and same token sequence ignoring
  whitespace/comments, or identical source text; for plain variables: same
  name).
- **D7** When such a fragment exists, report `F` once (stop at the first match).

The instanceof fragment may appear before or after the call and may sit inside
nested parentheses of the same `||` chain.

- **Name case.** Wherever this rule compares two expressions for
  equivalence, the names PHP resolves case-insensitively — function and
  method names, class names in calls, `new`, `instanceof` and `::`
  accesses, and keywords — compare case-insensitively (`Cache::$map['k']`
  matches `cache::$map['k']`, `$o->Name()` matches `$o->name()`); variable,
  property and constant names stay case-sensitive.

## Exceptions (no report)

- **E1** Language level below 7.1.
- **E2** `is_array` with zero or more than one argument.
- **E3** The call is not a direct `||` operand (`!is_array($v) || …`,
  `(is_array($v)) || …`, `is_array($v) && …`, `is_array($v) or …`).
- **E4** The instanceof subject differs from the call argument, the class is
  not named `Traversable` (e.g. `Countable`, `Iterator`), or the instanceof sits in a
  fragment that is not directly part of the `||` chain (e.g. inside an `&&`
  sub-expression).

## Report

- Range: the whole `is_array(...)` call `F` (name through closing parenthesis).
  The instanceof part is not highlighted.
- Severity: info (weak warning).
- Message: `Use 'is_iterable({arg})' instead of the is_array()/instanceof Traversable pair.`
  where `{arg}` is the instanceof subject's source text.

## Fix

None.

## Options

None.

## PHP versions

Reported only when the configured language level is ≥ 7.1 (`is_iterable()`
exists since 7.1). The upstream fixture runs at 7.1.

## Examples

```php
<?php
function walkAll($bag, $other) {
    $a = <weak_warning descr="Use 'is_iterable($bag)' instead of the is_array()/instanceof Traversable pair.">is_array($bag)</weak_warning> || $bag instanceof \Traversable;
    $b = ($bag instanceof Traversable) || <weak_warning descr="Use 'is_iterable($bag->items)' instead of the is_array()/instanceof Traversable pair.">is_array($bag->items)</weak_warning> || $bag->items instanceof Traversable;
    $c = $other === null || (<weak_warning descr="Use 'is_iterable($other)' instead of the is_array()/instanceof Traversable pair.">is_array($other)</weak_warning> || ($other instanceof \Traversable));

    $d = is_array($bag) || $other instanceof Traversable;           // different subject
    $e = (is_array($bag)) || $bag instanceof Traversable;           // call is parenthesised
    $f = is_array($bag) || $bag instanceof Iterator;              // other interface
    $g = is_array($bag) || ($bag instanceof Traversable && $other); // inside &&
    $h = is_array($bag) or $bag instanceof Traversable;             // keyword or
    $i = is_array($bag) && $x || $bag instanceof Traversable;       // direct parent is &&
}
```

## Divergences

- Only `\Traversable` itself is accepted; sub-interfaces
  (`Iterator`, `IteratorAggregate`) are not treated as equivalent, matching
  upstream.
- **Name resolution (custos diverges):** upstream compares the written last
  segments case-sensitively: it misses `Is_Array($v) || $v instanceof
  \TRAVERSABLE`, and reports user functions named `is_array` and unrelated
  classes such as `Some\Traversable` (or a bare `Traversable` inside a
  namespace). custos requires the call to resolve to the global `is_array()`
  (unqualified or `\`-qualified, not shadowed by a namespaced or imported
  function) and the class to resolve to `\Traversable`, both names compared
  case-insensitively (D1, D6).
- **Name case (custos diverges).** Upstream compares the expressions
  textually, so operands that differ only in the case of a function, method
  or class name (`Stats::$n` vs `stats::$n`), which PHP treats as the same,
  are not recognised as equivalent. custos folds the case of those names
  (Detection, "Name case").
