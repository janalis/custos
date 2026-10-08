---
id: UnsupportedEmptyListAssignments
group: Language level migration
kind: syntax
needs: []
php: { min: "7.0", max: "" }
---

# UnsupportedEmptyListAssignments

## Summary

Since PHP 7.0 destructuring into a pattern that has no target at all
(`list()`, `list(, )`, `[]`, `[, ]`) is a compile-time fatal error. In a
`foreach` value position this is easy to miss.

## Detection

- **D1** A `foreach` statement that declares **no loop variable at all**: no
  key variable, no value variable, and no variable anywhere inside a
  destructuring pattern.
- **D2** Scanning the `foreach` header after the `as` keyword, the first
  `list` keyword or opening `[` that is a direct part of the header (the start
  of a destructuring pattern in the value position) is found.
- **D3** Report that token.

Plain destructuring assignments (`list(, ) = $x;`, `[, ] = $x;`) are **not**
reported by this rule.

## Exceptions (no report)

- **E1** Language level below 7.0.
- **E2** The pattern contains at least one variable (`list($a, )`,
  `[, $b]`).
- **E3** A key variable is present (`foreach ($x as $k => list())`) — the
  foreach then declares a variable, so nothing is reported even though PHP
  would reject the empty pattern.
- **E4** Destructuring assignments outside `foreach`.

## Report

- Range: the `list` keyword (4 characters) or the single `[` character that
  opens the pattern.
- Severity: error.
- Message: `Empty destructuring pattern: PHP 7+ rejects this with a fatal error.`

## Fix

None.

## Options

None.

## PHP versions

Reported only at language level ≥ 7.0 (earlier versions accepted it). The
upstream fixture runs at 7.0.

## Examples

```php
<?php
foreach ($queue as <error descr="Empty destructuring pattern: PHP 7+ rejects this with a fatal error.">list</error>(, , )) {
    tick();
}
foreach ($queue as <error descr="Empty destructuring pattern: PHP 7+ rejects this with a fatal error.">[</error>]) {
    tick();
}
foreach ($queue as list(, $second)) {
    use_it($second);
}
foreach ($queue as [$first]) {
    use_it($first);
}
list(, $only) = $queue;
```

## Divergences

- Upstream decides "no variables" from the loop's variable list; a pattern
  whose only targets are property or array-element writes
  (`foreach ($rows as list($this->a))`) may be seen as variable-less and
  reported although valid. Recommendation: report only when the pattern
  (recursively) contains no assignable target at all. Not covered by upstream
  fixtures.
- Empty patterns behind a key (`$k => list()`) and in plain assignments are
  real fatal errors that this rule leaves to the parser. Recommendation: keep
  upstream behaviour (a parse diagnostic covers them).
