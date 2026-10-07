---
id: NestedNotOperators
group: Code style
kind: syntax
needs: []
php: { min: "", max: "" }
---

# NestedNotOperators

## Summary
Stacking logical-not operators (`!!$x`, `!!!$x`) is a cryptic way to write a
boolean cast or a single negation. An even number of `!` is a `(bool)` cast, an
odd number is one `!`.

## Detection
Consider a maximal chain of logical-not operators `!` where each `!` applies to
the next one, possibly through any number of parentheses, e.g. `!!$v`,
`!(!$v)`, `!(!((!$v)))`.

- D1: the innermost `!` is the one whose operand — after stripping any
  parentheses — is not itself a `!` expression. Call that stripped operand `X`.
- D2: walking outwards from the innermost `!`, count the `!` operators
  (`N`, starting at 1), passing through parentheses. Report when `N ≥ 2`.
- D3: the reported node `T` is the outermost `!` of the chain.
- D4: the subject text `S` is `X`'s source text; if `X` is a binary expression
  (any binary operator: arithmetic, comparison, logical, `instanceof`, `??`, …)
  `S` is `(` + text + `)`. Parentheses that originally surrounded `X` are
  dropped (they were stripped in D1).
- D5: if `N` is even the suggested replacement is `(bool)S`; if odd it is `!S`.

Each chain is reported once (only from its innermost `!`).

## Exceptions (no report)
- E1: a single `!` (N = 1).
- E2: chains broken by other operators are handled per D-walk rules; see
  Divergences for unary operators other than `!` sitting between nots.

## Report
- Range: the outermost `!` expression `T` — from its `!` token to the end of
  its operand (including closing parentheses that belong to the operand). Any
  parentheses wrapping `T` itself are not included.
- Severity: info (weak warning).
- Message: "Simplify the stacked negations to '{replacement}'." where
  `{replacement}` is the D5 text.

## Fix
- F1: replace `T` with `(bool)S` (even N) or `!S` (odd N), with no space after
  `(bool)` or `!`. Examples: `!!$k` → `(bool)$k`; `!!!$k` → `!$k`;
  `!! ($a && $b)` → `(bool)($a && $b)`; `!(!(($k)))` → `(bool)$k`;
  `!(!((!$k)))` → `!$k`.

## Options
None.

## PHP versions
No gating.

## Examples

```php
<?php
$ok   = <weak_warning descr="Simplify the stacked negations to '(bool)$token'.">!!$token</weak_warning>;
$off  = <weak_warning descr="Simplify the stacked negations to '!$token'.">!!!$token</weak_warning>;
$six  = <weak_warning descr="Simplify the stacked negations to '(bool)$token'.">!!!!!!$token</weak_warning>;
$both = <weak_warning descr="Simplify the stacked negations to '(bool)($m > $n)'.">!!($m > $n)</weak_warning>;
$neg  = <weak_warning descr="Simplify the stacked negations to '!($p instanceof Item)'.">!!! ($p instanceof Item)</weak_warning>;
$wrap = <weak_warning descr="Simplify the stacked negations to '(bool)count($list)'.">!( !(count($list)) )</weak_warning>;
$deep = <weak_warning descr="Simplify the stacked negations to '!$flag'.">!(!(!(($flag))))</weak_warning>;
$one  = !$token;
```

```php
<?php
$ok   = (bool)$token;
$off  = !$token;
$six  = (bool)$token;
$both = (bool)($m > $n);
$neg  = !($p instanceof Item);
$wrap = (bool)count($list);
$deep = !$flag;
$one  = !$token;
```

## Divergences
- Upstream walks outwards through *any* unary operator (unary minus/plus,
  bitwise not, casts, error suppression `@`, …) while counting only the `!`
  ones, so `!(int)!$v` or `!-!$v` is reported as `(bool)$v` and the fix drops
  the intermediate operator (semantics change). Recommendation: stop the
  outward walk at the first unary operator that is not `!` (only `!` and
  parentheses extend a chain). Not covered by upstream fixtures; results for
  `-!!$v` are identical either way (outermost `!` reported, `-` kept).
- Upstream wraps only binary subjects in parentheses; a ternary or assignment
  subject (`!!($a ? $b : $c)`, `!!($x = f())`) would produce `(bool)$a ? $b : $c`
  (wrong precedence). Recommendation: also parenthesise ternary, assignment
  and other low-precedence subjects (`print`, `yield`, `include`, `throw`). Not
  covered by upstream fixtures.
