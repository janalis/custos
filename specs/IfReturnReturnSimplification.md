---
id: IfReturnReturnSimplification
group: Control flow
kind: syntax
needs: []
php: { min: "", max: "" }
---

# IfReturnReturnSimplification

## Summary

An `if` whose only job is to return `true` in one branch and `false` in the
other is a long-winded way of returning the condition itself (or its
negation). Return the expression directly.

## Detection

Visit every `if` statement `S` (the keyword form; `else if` is an `else` whose
body is a nested `if` and that nested `if` is visited on its own).

- **D1** Let `C` be `S`'s condition with all wrapping parentheses removed
  (`if ((($a > 1)))` → `$a > 1`). `C` must be a binary expression whose
  result is always a `bool`: a comparison (`===`, `!==`, `==`, `!=`/`<>`,
  `<`, `<=`, `>`, `>=`), a logical operator (`&&`, `||`, `and`, `or`, `xor`)
  or `instanceof`. Other binary operators (arithmetic, bitwise, `.`, `??`,
  `<=>`, …) do not match, nor do assignments and ternaries.
- **D2** `S` has no `elseif` branches.
- **D3** `S`'s `if` body is a braced block `{ … }` (or the statement list of
  the alternative syntax) containing exactly one statement; comments (line,
  block, doc) do not count. That statement is a `return` — call it `R1`.
  Bodies without braces (`if (c) return true;`) never match.
- **D4** Find the second return `R2`:
  - **D4a** if `S` has an `else` branch: its body must be a braced block with
    exactly one statement (comments ignored) and that statement is a `return`.
    `else if (…)` (else whose body is an `if`) and unbraced `else return …;`
    do not match.
  - **D4b** if `S` has no `else`: the next statement after `S` in the same
    statement list (skipping whitespace and comments) must be a `return`.
- **D5** Return values: the argument of a `return` counts as `true`/`false`
  only if it is literally the constant `true`/`false` (case-insensitive, a
  leading `\` is allowed). `return (true);`, `return 1;`, `return;` do not
  count.
  - *direct*: `R1` returns `true` and `R2` returns `false`;
  - *reverse*: `R1` returns `false` and `R2` returns `true`.
  One of the two must hold.
- **D6** Only for D4b (no `else`): skip when the statement immediately
  preceding `S` (skipping whitespace and comments) is an `if` that has no
  `else`/`elseif` branch and whose braced body's last statement (comments
  ignored, any number of statements) is a `return`. This keeps chains of
  "guard-if return" blocks readable. The guard does not apply when `S` has an
  `else`.

## Exceptions (no report)

- **E1** Condition (after unwrapping parentheses) is not binary: a variable,
  call, `!`-expression, `isset(...)`, assignment, ternary; or it is a binary
  expression whose value is not a `bool` (`$a + $b`, `$a . $b`, `$a ?? $b`,
  `$a & $b`, `$a <=> $b`).
- **E2** `S` has an `elseif`.
- **E3** Unbraced if/else bodies, bodies with more than one statement.
- **E4** Both returns give the same constant (`true`/`true`, `false`/`false`),
  or one of them returns something other than a bare `true`/`false`.
- **E5** D6: preceded by a no-else `if` ending with `return`, and `S` has no
  `else`.

## Report

- Range: the `if` keyword token of `S` (2 characters).
- Severity: warning.
- Message: `Return the condition directly: '{replacement}'.` where
  `{replacement}` is the final F-text below without the trailing `;`.

## Fix

Let `T` be the source text of `C` (D1, parentheses already stripped, inner
text kept verbatim).

- **F1** direct case: the new statement is `return T;`.
- **F2** reverse case: if `C`'s operator is one of the comparison operators
  below, the new statement is `return L OP' R;` where `L`/`R` are the verbatim
  source texts of `C`'s left and right operands, `OP'` the inverted operator,
  joined by single spaces:

  | in | out |
  |----|-----|
  | `===` | `!==` |
  | `!==` | `===` |
  | `==` | `!=` |
  | `!=` (and `<>`) | `==` |
  | `>` | `<=` |
  | `>=` | `<` |
  | `<` | `>=` |
  | `<=` | `>` |

  Otherwise (any other binary operator, e.g. `&&`, `instanceof`, `.`) the new
  statement is `return !(T);`.
- **F3** Placement:
  - with an `else` (D4a): the whole `if … else …` statement is replaced by the
    new statement;
  - without `else` (D4b): the new statement is inserted where `S` starts and
    everything from the start of `S` through the end of `R2` (including any
    whitespace and comments between them) is removed.
  The indentation before `S` is kept; text after `R2` is untouched.

## Options

None.

## PHP versions

No gating.

## Examples

```php
<?php
function isAdult(int $age): bool {
    <warning descr="Return the condition directly: 'return $age >= 18'.">if</warning> ($age >= 18) {
        return true;
    }
    return false;
}

function isBlank(string $s): bool {
    <warning descr="Return the condition directly: 'return $s !== '''.">if</warning> (($s === '')) { return FALSE; }
    // fallthrough
    return TRUE;
}

function hasBoth($m, $n) {
    <warning descr="Return the condition directly: 'return !($m && $n)'.">if</warning> ($m && $n) { return false; }
    else { return true; }
}

function notWidget($w) {
    <warning descr="Return the condition directly: 'return !($w instanceof Widget)'.">if</warning> ($w instanceof Widget) {
        return false;
    } else {
        return true;
    }
}

function guarded($q) {
    if ($q === null) { log_it(); return true; }
    if ($q > 9) { return true; }
    return false;
}

function guardedElse($q) {
    if ($q === null) { return true; }
    <warning descr="Return the condition directly: 'return $q < 3'.">if</warning> ($q < 3) { return true; }
    else { return false; }
}

function notBinary($q) {
    if (is_int($q)) { return true; }
    return false;
}

function unbraced($q) {
    if ($q > 1) return true;
    return false;
}

function sameValue($q) {
    if ($q > 1) { return false; }
    return false;
}

function hasElseif($q) {
    if ($q > 1) { return true; }
    elseif ($q < 0) { return true; }
    else { return false; }
}
```

```php
<?php
function isAdult(int $age): bool {
    return $age >= 18;
}

function isBlank(string $s): bool {
    return $s !== '';
}

function hasBoth($m, $n) {
    return !($m && $n);
}

function notWidget($w) {
    return !($w instanceof Widget);
}

function guarded($q) {
    if ($q === null) { log_it(); return true; }
    if ($q > 9) { return true; }
    return false;
}

function guardedElse($q) {
    if ($q === null) { return true; }
    return $q < 3;
}

function notBinary($q) {
    if (is_int($q)) { return true; }
    return false;
}

function unbraced($q) {
    if ($q > 1) return true;
    return false;
}

function sameValue($q) {
    if ($q > 1) { return false; }
    return false;
}

function hasElseif($q) {
    if ($q > 1) { return true; }
    elseif ($q < 0) { return true; }
    else { return false; }
}
```

## Divergences

- custos diverges from upstream on the accepted operators (D1/E1). Upstream
  takes any binary operator, so `if ($a + $b) { return true; } return false;`
  would become `return $a + $b;`, which returns an `int` instead of a `bool`
  and can break a `bool` return type or strict comparisons by callers (the
  same holds for `.`, `??`, bitwise operators and `<=>`). custos only reports
  conditions that are already boolean-valued.
- `return !(T);` for a reverse case keeps `T` verbatim; when `C` is itself an
  `and`/`or` expression the result is still correct thanks to the
  parentheses. No divergence needed.
- Alternative syntax (`if (…): return true; endif; return false;`) is not
  covered by upstream fixtures. Recommendation: treat its statement list like
  a braced block.
