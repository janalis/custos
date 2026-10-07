---
id: SenselessTernaryOperator
group: Confusing constructs
kind: syntax
needs: []
php: { min: "", max: "" }
---

# SenselessTernaryOperator

## Summary
`$a === $b ? $a : $b` always evaluates to `$b`: when the two are identical it
does not matter which one is returned. The same holds for `!==` with swapped
branches. Such ternaries only add branching; replace them with the operand
that is always the result.

## Detection
- **D1** A full ternary `C ? T : F` (short ternaries `C ?: F` are never
  reported).
- **D2** `C` with all surrounding parentheses removed is a binary `===` or
  `!==` expression with left operand `L` and right operand `R`.
- **D3** Effective branches: for `===`, `Yes = T`, `No = F`; for `!==`,
  `Yes = F`, `No = T`.
- **D4** The operands map one-to-one onto the branches: either `L` is
  equivalent to `Yes` and `R` to `No`, or `L` is equivalent to `No` and `R`
  to `Yes`. Matching only one branch with both operands
  (`$x === $x ? $x : $y`) does not qualify.
- **D5** Equivalence: same node kind and either (a) for two variables, equal
  names; or (b) structurally identical ignoring whitespace and comments, or
  identical source text, where keywords and the names PHP resolves
  case-insensitively (function and method names, class names in calls,
  `new`, `instanceof` and `::` accesses) compare case-insensitively
  (`count($a)` matches `COUNT($a)`); variables, properties and constants stay
  case-sensitive. Parentheses are **not** stripped from `T`, `F`, `L`,
  `R` for this comparison (`($x)` is not equivalent to `$x`).
- **D6** Replacement text = the source text of `No` exactly as written
  (including its own parentheses, if any).

## Exceptions (no report)
- **E1** Loose comparisons (`==`, `!=`, `<>`), other operators, or a
  condition that is not a comparison (`$x ? $x : $y`).
- **E2** Short ternaries.
- **E3** Operands that differ (e.g. `$x === $y ? $x : $z`).
- **E4** Both operands matching the same branch while the other branch is
  different: `$x === $x ? $x : $y` (the result is `$x`, not `$y`).

## Report
- Range: the whole ternary expression (from the start of `C` — including
  parentheses around `C` if written — to the end of `F`). Parentheses around
  the ternary itself are not included.
- Severity: warning.
- Message: `This ternary always yields '{replacement}'; use it directly.`

## Fix
- **F1** Replace the ternary expression (the reported range) with the
  replacement text from D6. Surrounding code (parentheses around the ternary,
  the statement's `;`) is left untouched.
  - `$p === $q ? $p : $q` → `$q`
  - `$p !== $q ? $p : $q` → `$p`
  - `$n === 1 ? 1 : $n` → `$n`;  `$n === 1 ? $n : 1` → `1`
  - `$n !== 1 ? $n : 1` → `$n`;  `$n !== 1 ? 1 : $n` → `1`

## Options
None.

## PHP versions
No gating.

## Examples

```php
<?php
$a = <warning descr="This ternary always yields '$right'; use it directly.">$left === $right ? $left : $right</warning>;
$b = <warning descr="This ternary always yields '$left'; use it directly.">$left !== $right ? $left : $right</warning>;
$c = <warning descr="This ternary always yields '$size'; use it directly.">$size === -1 ? -1 : $size</warning>;
$d = <warning descr="This ternary always yields 'null'; use it directly.">($ref === null) ? $ref : null</warning>;
$e = <warning descr="This ternary always yields '$t'; use it directly.">$t !== 'n/a' ? $t : 'n/a'</warning>;
$f = f(<warning descr="This ternary always yields '1.5'; use it directly.">$ratio !== 1.5 ? 1.5 : $ratio</warning>);
$g = $left == $right ? $left : $right;
$h = $left === $right ?: $right;
$i = $left === $right ? ($left) : $right;
$j = $left === $right ? $left : $other;
```

```php
<?php
$a = $right;
$b = $left;
$c = $size;
$d = null;
$e = $t;
$f = f(1.5);
$g = $left == $right ? $left : $right;
$h = $left === $right ?: $right;
$i = $left === $right ? ($left) : $right;
$j = $left === $right ? $left : $other;
```

## Divergences
- **Case of names in the equivalence (custos diverges from upstream).**
  Upstream compares operands textually, so `count($a) === Count($b) ?
  count($a) : Count($b)` is not reported although both spellings call the
  same function. custos compares function, method and class names
  case-insensitively as PHP does (D5).
- **custos diverges from upstream** on degenerate conditions (D4/E4).
  Upstream checks each operand against the branches independently, so
  `$x === $x ? $x : $y` is reported and rewritten to `$y`, although the
  condition is always true and the expression always yields `$x`. custos
  requires the two operands to match the two different branches, which is
  exactly the case where the ternary always yields its "no" branch.
