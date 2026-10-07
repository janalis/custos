---
id: NestedTernaryOperator
group: Confusing constructs
kind: syntax
needs: []
php: { min: "", max: "" }
---

# NestedTernaryOperator

## Summary
A ternary inside another ternary is hard to read and, without parentheses,
its associativity is a classic PHP trap. Chains of short ternaries
(`$a ?: $b ?: $c`) read naturally and are accepted.

## Detection
For every ternary expression `O` (full `c ? t : f` or short `c ?: f`), look at
its operands; "unwrapped" means with all surrounding parentheses removed (any
depth):
- **D1** If the unwrapped condition is a ternary → report that inner ternary.
- **D2** If `O` is a full ternary and its unwrapped "then" operand is a
  ternary → report that inner ternary.
- **D3** If the unwrapped "else" operand is a ternary → report that inner
  ternary.
Each inner ternary can be reported at most once per parent; an expression with
several nested ternaries yields several reports (one per nested node, at every
depth — the inner ternary is itself checked as an `O`).

### Elvis chains (allowed)
- **E1** An unparenthesised chain of short ternaries is not reported:
  `$a ?: $b ?: $c` (any length). In terms of the syntax tree:
  - if the parser nests the chain to the **left** (`($a ?: $b) ?: $c`, PHP's
    actual associativity — custos' parser does this), D1 must not fire when
    both `O` and its condition are short ternaries **and** the condition is
    not wrapped in parentheses in the source;
  - (equivalently, for a right-nesting tree: D3 must not fire when both are
    short and the else operand is not parenthesised.)
  As soon as either link of the chain is written with explicit parentheses,
  `($a ?: $b) ?: $c` or `$a ?: ($b ?: $c)`, the inner ternary **is**
  reported.

## Exceptions (no report)
- **E1** Unparenthesised short-ternary chains (above).
- **E2** Ternaries whose operands contain ternaries only deeper inside other
  expressions (e.g. `$a ? f($b ? 1 : 2) : 3`, `$a ? [$b ?: 0] : 1`): only a
  ternary that is directly an operand (through parentheses only) counts.

## Report
- Range: the inner ternary expression, without the parentheses that wrap it
  (from the first character of its condition to the last character of its
  else operand).
- Severity: warning.
- Message: `Avoid nesting ternary operators; use if/else or extract a
  variable.`

## Fix
None.

## Options
None.

## PHP versions
No gating. Unparenthesised nested full ternaries (`$a ? 1 : $b ? 2 : 3`) are
deprecated in 7.4 and an error in 8.0, but are parsed at every level (see
Divergences for the range).

## Examples

```php
<?php
$size  = $big ? 'L' : (<warning descr="Avoid nesting ternary operators; use if/else or extract a variable.">$mid ? 'M' : 'S'</warning>);
$mode  = $ro ? (<warning descr="Avoid nesting ternary operators; use if/else or extract a variable.">$admin ? 'view-all' : 'view'</warning>) : 'edit';
$state = ((<warning descr="Avoid nesting ternary operators; use if/else or extract a variable.">$on ? 1 : 0</warning>)) ? 'up' : 'down';
$name  = (<warning descr="Avoid nesting ternary operators; use if/else or extract a variable.">$nick ?: $first</warning>) ?: 'anon';
$name  = $nick ?: (<warning descr="Avoid nesting ternary operators; use if/else or extract a variable.">$first ?: 'anon'</warning>);
$name  = $nick ?: $first ?: $last ?: 'anon';
$call  = $ok ? strtoupper($v ? 'y' : 'n') : '';
```

## Divergences
- Unparenthesised mixed chains such as `$a ? 1 : $b ? 2 : 3` or
  `$a ? $b : $c ?: $d`: upstream's tree shape for these is not observable in
  the fixtures, so the exact node reported is unknown. Recommendation: with
  custos' left-nesting tree, report the inner ternary that is the condition
  (`$a ? 1 : $b`), since E1 only exempts short-in-short chains.
