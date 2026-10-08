---
id: ShortListSyntaxCanBeUsed
group: Language level migration
kind: syntax
needs: []
php: { min: "7.1", max: "" }
---

# ShortListSyntaxCanBeUsed

## Summary

From PHP 7.1 destructuring can use the short `[...]` form instead of
`list(...)`, consistent with short array syntax.

## Detection

- **D1 Assignment form**: a destructuring assignment whose left side starts
  with the `list` keyword (`list(...) = expr`), wherever the assignment
  appears: as a statement on its own or as a sub-expression (a `while`
  condition, a `for` header, an argument, the right side of another
  assignment, …). Report its `list` keyword.
- **D2 foreach form**: a `foreach` statement that declares at least one loop
  variable (key, value, or a variable inside the destructuring pattern), whose
  header contains a `list` keyword as a direct part of the `foreach` construct
  (the value target, with or without a key: `as list(...)`,
  `as $k => list(...)`). Report the first such `list` keyword (scanning the
  header from left to right, stopping at the loop body).

## Exceptions (no report)

- **E1** Language level below 7.1.
- **E3** Already short syntax (`[$a, $b] = …`, `foreach ($x as [$a, $b])`).
- **E4** `foreach` with no variable at all (e.g. an all-empty pattern — that is
  handled by UnsupportedEmptyListAssignments).
- **E5** Nested `list(...)` inside an outer pattern is not reported on its own
  (only the outermost `list` keyword of a statement / foreach).

## Report

- Range: the `list` keyword token only (4 characters).
- Severity: info (weak warning).
- Message (D1): `Use short destructuring syntax '[...] = ...'.`
- Message (D2): `Use short destructuring syntax in foreach ('as [...]').`

## Fix

- **F1** Starting at the reported `list` keyword, find the `(` that opens its
  pattern and the matching `)` (balanced at the pattern's own level). Then:
  delete the whitespace run directly following `list` (if any), replace that
  `(` with `[`, replace the matching `)` with `]`, and delete the `list`
  keyword. Everything between the parentheses (items, keys `'k' => $v`, empty
  slots, comments, inner whitespace) is kept verbatim.
  - `list($a, $b) = $pair;` → `[$a, $b] = $pair;`
  - `list ( 'x' => $x ) = $p;` → `[ 'x' => $x ] = $p;`
  - `foreach ($rows as $i => list($u, $v))` → `foreach ($rows as $i => [$u, $v])`

## Options

None.

## PHP versions

Reported only at language level ≥ 7.1 (short destructuring exists since
7.1). The upstream fixture runs at 7.1.

## Examples

```php
<?php
<weak_warning descr="Use short destructuring syntax '[...] = ...'.">list</weak_warning>($host, $port) = explode(':', $addr);
<weak_warning descr="Use short destructuring syntax '[...] = ...'.">list</weak_warning> ('w' => $w, 'h' => $h) = $size;

foreach ($pairs as <weak_warning descr="Use short destructuring syntax in foreach ('as [...]').">list</weak_warning>($left, $right)) {
    swap($left, $right);
}
foreach ($grid as $row => <weak_warning descr="Use short destructuring syntax in foreach ('as [...]').">list</weak_warning> (, $cell)) {
    show($row, $cell);
}

while (<weak_warning descr="Use short destructuring syntax '[...] = ...'.">list</weak_warning>($key, $item) = each($legacy)) {
    use_it($key, $item);
}
[$p, $q] = $coords;
```

```php
<?php
[$host, $port] = explode(':', $addr);
['w' => $w, 'h' => $h] = $size;

foreach ($pairs as [$left, $right]) {
    swap($left, $right);
}
foreach ($grid as $row => [, $cell]) {
    show($row, $cell);
}

while ([$key, $item] = each($legacy)) {
    use_it($key, $item);
}
[$p, $q] = $coords;
```

## Divergences

- Upstream converts only the outermost `list(...)`; a nested
  `list($a, list($b, $c)) = $v;` becomes `[$a, list($b, $c)] = $v;`, which
  PHP rejects (mixing `[]` and `list()` is a compile error). Recommendation:
  convert nested `list(` keywords inside the same pattern as well (still one
  report). Not covered by upstream fixtures.
- **`list()` in sub-expressions (custos diverges from upstream).** Upstream
  only reports `list(...) = …` used as a whole statement, so
  `while (list($k, $v) = each($a))`, `for (list($i) = $start; …)` or
  `$copy = list($x, $y) = $p;` are left alone. The short form is valid and
  means the same in every one of those positions, so custos reports and
  converts them too (D1).
