---
id: TernaryOperatorSimplify
group: Control flow
kind: syntax
needs: [types]
php: { min: "", max: "" }
---

# TernaryOperatorSimplify

## Summary
A ternary whose branches are the literals `true` and `false` merely restates
its condition. When the condition is a binary expression, the ternary can be
replaced by the condition itself, its negation, or a boolean cast.

## Detection
Visit every ternary expression `C ? T : F` (full form; a short ternary `C ?: F`
has no explicit true branch and is never reported).

- **D1** Strip any parentheses around `C`; the result `B` must be a binary
  expression (any binary operator: comparison, arithmetic, bitwise, `&&`,
  `||`, `and`, `or`, `xor`, `.`, `??`, `instanceof`, `<=>`, …). Function calls,
  variables, unary `!` etc. are not targets.
- **D2** Strip any parentheses around `T` and around `F`. Both must be one of
  the constants `true` / `false` (case-insensitive, optionally written with a
  leading `\`). Any combination qualifies, including `true : true` and
  `false : false` (see Divergences).
- **D3** `inverted` is true when the stripped `T` is `false`.
- **D4** The replacement `R` is built from `B` (operator = `B`'s operator token):
  - **Comparison operators** `==`, `===`, `!=` (also spelled `<>`), `!==`, `>`,
    `<`, `>=`, `<=`:
    - not inverted: `R` = source text of `B`, verbatim;
    - inverted: `R` = `left + " " + opposite + " " + right`, with left/right
      operand source texts verbatim and the opposite operator from:
      `==`→`!=`, `===`→`!==`, `!=`/`<>`→`==`, `!==`→`===`, `>`→`<=`,
      `<`→`>=`, `>=`→`<`, `<=`→`>`. Original spacing around the operator is not
      kept (always single spaces).
    - **D4a** inverted ordering operator (`>`, `<`, `>=`, `<=`) when either
      operand is not known to be `int`, `string`, `bool` or `null` (inferred
      type; unknown, `float`, arrays, objects and unions containing them all
      fail): `R` = `!(` + text of `B` + `)`. For floats (NAN) and
      incomparable arrays/objects, `a < b` and `a >= b` can both be false,
      so swapping the operator would change the result; negation does not.
  - **Symbolic logical operators** `&&` and `||`:
    - not inverted: `R` = `(` + text of `B` + `)`;
    - inverted: `R` = `!(` + text of `B` + `)`.
  - **Any other binary operator** (including keyword `and`/`or`/`xor`):
    - not inverted: `R` = `(bool)(` + text of `B` + `)`;
    - inverted: `R` = `!(` + text of `B` + `)`.

## Exceptions (no report)
- **E1** Condition (after paren stripping) is not a binary expression
  (`is_int($v) ? false : true`, `$flag ? true : false`, `!$a ? …`).
- **E2** Either branch is not a boolean constant (`null`, `0`, `1`, `'yes'`,
  a variable…).
- **E3** Short ternary `?:`.

## Report
- Range: the whole ternary expression, from the first character of `C` (including
  any parentheses that are part of `C`) to the end of `F`. Parentheses wrapping
  the ternary itself are not included.
- Severity: info (weak warning).
- Message: `Replace the ternary with '{R}'.`

## Fix
- **F1** Replace the ternary expression with `R` verbatim. No extra
  parentheses are added around `R` even if the surrounding context binds
  tighter (see Divergences).
  - `$ok = $n < 10 ? true : false;` → `$ok = $n < 10;`
  - `$ok = $n < 10 ? false : true;` with `int $n` → `$ok = $n >= 10;`
  - `$ok = $x < 10 ? false : true;` with untyped `$x` → `$ok = !($x < 10);`
  - `$ok = $m % 2 ? TRUE : FALSE;` → `$ok = (bool)($m % 2);`
  - `$ok = $a || $b ? false : true;` → `$ok = !($a || $b);`
  - `$ok = ($n < 10) ? (true) : false;` → `$ok = $n < 10;`

## Options
None.

## PHP versions
No gating.

## Examples

```php
<?php
function probe($n, $m, $a, $b, $obj, int $k) {
    $r = <weak_warning descr="Replace the ternary with '$n !== 7'.">$n !== 7 ? true : false</weak_warning>;
    $r = <weak_warning descr="Replace the ternary with '!($n <= 7)'.">$n <= 7 ? false : true</weak_warning>;
    $r = <weak_warning descr="Replace the ternary with '$k > 7'.">$k <= 7 ? false : true</weak_warning>;
    $r = <weak_warning descr="Replace the ternary with '(bool)($m % 2)'.">$m % 2 ? TRUE : FALSE</weak_warning>;
    $r = <weak_warning descr="Replace the ternary with '!($m | 4)'.">$m | 4 ? false : true</weak_warning>;
    $r = <weak_warning descr="Replace the ternary with '($a || $b)'.">$a || $b ? true : false</weak_warning>;
    $r = <weak_warning descr="Replace the ternary with '!($a || $b)'.">$a || $b ? false : true</weak_warning>;
    $r = <weak_warning descr="Replace the ternary with '(bool)($obj instanceof Countable)'.">$obj instanceof Countable ? true : false</weak_warning>;
    $r = <weak_warning descr="Replace the ternary with '$n == $m'.">($n == $m) ? (true) : (false)</weak_warning>;
    $r = $n == $m ? true : 0;
    $r = in_array($n, $b) ? false : true;
    $r = $n ?: false;
    return $r;
}
```

```php
<?php
function probe($n, $m, $a, $b, $obj, int $k) {
    $r = $n !== 7;
    $r = !($n <= 7);
    $r = $k > 7;
    $r = (bool)($m % 2);
    $r = !($m | 4);
    $r = ($a || $b);
    $r = !($a || $b);
    $r = (bool)($obj instanceof Countable);
    $r = $n == $m;
    $r = $n == $m ? true : 0;
    $r = in_array($n, $b) ? false : true;
    $r = $n ?: false;
    return $r;
}
```

## Divergences
- `C ? true : true` / `C ? false : false` are reported by upstream as if they
  were `true : false` / `false : true`, and the fix changes semantics.
  Recommendation: only report when the two branches differ.
- The fix inserts `R` without parentheses. Parentheses around the ternary are
  outside the replaced range and stay, so this is only a concern for an
  unparenthesised ternary used as an operand (rare, e.g. nested ternaries).
  Recommendation: follow upstream (verbatim insertion).
- **Inverted ordering comparisons (custos diverges):** upstream always swaps
  `<`/`>=` and `>`/`<=` when inverting. That is wrong for `NAN` (every
  ordering comparison is false) and for arrays or objects that cannot be
  ordered. custos swaps only when both operands are known non-float scalars
  and otherwise emits `!(B)` (D4a).
