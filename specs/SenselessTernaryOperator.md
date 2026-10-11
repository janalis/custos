---
id: SenselessTernaryOperator
group: Confusing constructs
kind: semantic
needs: [names, types]
php: { min: "", max: "" }
---

# SenselessTernaryOperator

## Summary

`$a === $b ? $a : $b` always evaluates to `$b`: when the two are identical it
does not matter which one is returned. The same holds for `!==` with swapped
branches. Such ternaries only add branching; replace them with the operand
that is always the result. This only holds when writing an operand twice
denotes the same value twice, so operands with side effects and float zero
literals are left alone.

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
- **E5** `L` or `R` may have a side effect: the ternary evaluates the
  repeated operand twice and the two evaluations need not agree
  (`next($a) === false ? false : next($a)` advances twice and returns the
  second result). An operand counts as impure when it contains, outside a
  closure or arrow function body: a method or static call, `new`, `clone`,
  an assignment, `++`/`--`, `include`/`require`, `eval`, `exit`, `print`,
  `throw`, `yield`, a backtick command, the pipe operator, or a function
  call that does not resolve to a native function from a fixed list of
  value-only builtins (`count`, `strlen`, `trim`, `abs`, `min`, `max`,
  `in_array`, `array_keys`, ...). Dynamic calls, user functions and a
  namespaced function shadowing a builtin are therefore impure.
- **E6** `L` or `R` contains a float literal whose value is zero (`0.0`,
  `.0`, `0e0`, also under a unary sign or nested in an array literal).
  `-0.0 === 0.0` holds although the two values behave differently
  (`fdiv(1, $x)`, printing), so `$x === 0.0 ? 0.0 : $x` is the usual way to
  drop the sign of a zero and is not senseless. Two non-literal operands
  that could hold zeros of opposite sign cannot be told apart and are still
  reported.

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
- **F2** A single space is added before the replacement when the character
  just before the ternary is a word character (letter, digit, `_`, non-ASCII
  byte) and the replacement starts with a word character or a backslash, and
  after it when the replacement ends with a word character and one follows
  directly. The ternary can start with `(` or end with a quote or bracket
  right against a keyword, where a bare name would fuse with it:
  `return($v === null) ? $v : null;` → `return null;`.

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
$k = next($list) === false ? false : next($list);
$l = $ratio === 0.0 ? 0.0 : $ratio;
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
$k = next($list) === false ? false : next($list);
$l = $ratio === 0.0 ? 0.0 : $ratio;
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
- **custos diverges from upstream** on operands that are not stable values
  (E5/E6). Upstream compares the operands as text only, so a repeated call
  with side effects or a negative-zero guard is reported and its fix changes
  the result. custos does not report either shape.
- **Fix spacing (F2).** custos separates the replacement from an adjacent
  keyword instead of producing a fused name such as `returnnull`.
