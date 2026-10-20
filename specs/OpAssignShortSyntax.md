---
id: OpAssignShortSyntax
group: Code style
kind: semantic
needs: [types]
php: { min: "", max: "" }
---

# OpAssignShortSyntax

## Summary

`$total = $total + $step` repeats the target; the compound assignment
`$total += $step` says the same thing more directly.

## Detection

Applies to plain assignments `V = R` (operator `=`; compound assignments and
destructuring are never inspected).

- D1: strip any parentheses around `R`; the result must be a binary expression
  whose operator `op` is one of: `+`, `-`, `*`, `/`, `%`, `.`, `&`, `|`, `^`,
  `<<`, `>>`. (`**`, `??`, logical and comparison operators are not
  eligible.)
- D2: walk down the left spine of that binary expression: collect its right
  operand as a fragment; then, while the left operand is a binary expression
  with the same operator `op`, collect its right operand and continue with its
  left operand. The walk stops at the first left operand that is not a binary
  expression with operator `op`; call it `B`.
- D3: `B` must be structurally equivalent to `V` (same token sequence ignoring
  whitespace and comments; for simple variables, the same name). If the walk
  stopped on a binary expression with a different operator, there is no match.
- D4: if more than one fragment was collected, `op` must be one of `+`, `.`,
  `*`. With exactly one fragment any eligible operator is accepted. (Of the
  three, only a `.` chain is rewritten automatically, see F2.)
- D5: no collected fragment may itself be a binary expression (fragments are
  inspected as written: a parenthesised fragment such as `($a - 1)` is not a
  binary expression and is accepted).
- Report when D1–D5 hold and E1 does not apply.

Fragments, in source order (left to right), form the suggested right-hand
side: `F1 op F2 op …`. Suggested replacement: `B op= F1 op F2 …` (with single
spaces, e.g. `$s .= 'a' . $t`).

- **Name case.** Wherever this rule compares two expressions for
  equivalence, the names PHP resolves case-insensitively — function and
  method names, class names in calls, `new`, `instanceof` and `::`
  accesses, and keywords — compare case-insensitively (`Cache::$map['k']`
  matches `cache::$map['k']`, `$o->Name()` matches `$o->name()`); variable,
  property and constant names stay case-sensitive.

## Exceptions (no report)

- E1: string offset writes: `V` is an array-access expression (`$s[i]`) whose
  base expression has a fully resolved type (no unknown component) that
  includes `string`. (Compound assignment on string offsets is a fatal error.)
  When the base type is unknown or partially unknown, the rule reports.
- E2: operand order reversed: `$n = 1 + $n` (the target is not the leftmost
  operand) is never reported.
- E3: mixed operators on the spine (`$n = $n . 'a' + 1`), non-chainable operator
  with several fragments (`$n = $n - $a - $b`, `/`, `%`, `&`, `|`, `^`, `<<`,
  `>>`), or a fragment that is an unparenthesised binary expression
  (`$n = $n + $a * 2`).
- E4: compound assignments (`$n += $n + 1`) are not inspected.

## Report

- Range: the whole assignment expression `V = R` (without trailing `;`).
- Severity: info (weak warning).
- Message: "Use the compound form '{replacement}'."

## Fix

- F1: replace the assignment expression with the suggested replacement:
  `<B text> <op>= <F1 text> <op> <F2 text> …`, fragments copied verbatim (keeping
  their own parentheses), one space on each side of `op=` and of each joining
  `op`. Parentheses that wrapped `R` disappear.
- F2: the report carries no fix when the compound form is not provably the
  same computation. "May have an effect" below means the fragment contains,
  outside closure and arrow-function bodies, a call of any kind, `new`,
  `clone`, an assignment, `++`/`--`, `include`/`require`, `eval`, `exit`,
  `print`, `throw`, `yield`, a pipe or a backtick string.
  - Several fragments joined by `+` or `*`: `$x = $x + 1 + 1` computes
    `($x + 1) + 1`, the compound form `$x + (1 + 1)`. The two differ for
    floats near their precision limit and when an integer intermediate
    overflows. A `.` chain regroups freely and keeps its fix.
  - `V` is not a plain `$name` variable (property, static property, element,
    variable variable) and a fragment may have an effect: the long form reads
    `V` before the fragment runs, the compound form after it, so a fragment
    that writes `V` (`$this->out = $this->out . $this->child()`) gives another
    result.
  - `V` is a plain variable, there are several fragments and one after the
    first may have an effect: the long form has already combined `V` with the
    first fragment when the later ones run (`$s = $s . 'a' . ($s = 'b')`). A
    plain variable is otherwise read after the fragment in both forms, so a
    single fragment, or the first of a chain, may contain calls.

## Options

None.

## PHP versions

No gating.

## Examples

```php
<?php
function tally(int $count, string $label, array $grid, $buf) {
    <weak_warning descr="Use the compound form '$count += 5'.">$count = $count + 5</weak_warning>;
    <weak_warning descr="Use the compound form '$count -= $step'.">$count = $count - $step</weak_warning>;
    <weak_warning descr="Use the compound form '$count %= 7'.">$count = ($count % 7)</weak_warning>;
    <weak_warning descr="Use the compound form '$count <<= 1'.">$count = $count << 1</weak_warning>;
    <weak_warning descr="Use the compound form '$label .= '-' . $count . '!''.">$label = $label . '-' . $count . '!'</weak_warning>;
    <weak_warning descr="Use the compound form '$count *= 3 * $k'.">$count = $count * 3 * $k</weak_warning>;
    <weak_warning descr="Use the compound form '$count += ($k * 2)'.">$count = $count + ($k * 2)</weak_warning>;
    <weak_warning descr="Use the compound form '$grid[1] |= 4'.">$grid[1] = $grid[1] | 4</weak_warning>;
    <weak_warning descr="Use the compound form '$buf[0] .= 'z''.">$buf[0] = $buf[0] . 'z'</weak_warning>;
    <weak_warning descr="Use the compound form '$grid[2] += next_id()'.">$grid[2] = $grid[2] + next_id()</weak_warning>; // no fix

    $count = 5 + $count;          // reversed operands
    $count = $count - $a - $b;    // '-' cannot be regrouped
    $count = $count + $k * 2;     // fragment is a binary expression
    $count = $count ** 2;         // '**' not eligible
    $label[0] = $label[0] . 'x';  // string offset (typed string)
}
```

```php
<?php
function tally(int $count, string $label, array $grid, $buf) {
    $count += 5;
    $count -= $step;
    $count %= 7;
    $count <<= 1;
    $label .= '-' . $count . '!';
    $count = $count * 3 * $k;
    $count += ($k * 2);
    $grid[1] |= 4;
    $buf[0] .= 'z';
    $grid[2] = $grid[2] + next_id(); // no fix

    $count = 5 + $count;          // reversed operands
    $count = $count - $a - $b;    // '-' cannot be regrouped
    $count = $count + $k * 2;     // fragment is a binary expression
    $count = $count ** 2;         // '**' not eligible
    $label[0] = $label[0] . 'x';  // string offset (typed string)
}
```

## Divergences

Note: with PHP 8 precedence (`+`/`-` bind tighter than `.`),
`$n = $n . 'a' + 1` parses as `$n . ('a' + 1)` whose single fragment is a binary
expression → not reported, the same outcome as with the older precedence.

- **Name case (custos diverges).** Upstream compares the expressions
  textually, so operands that differ only in the case of a function, method
  or class name (`Stats::$n` vs `stats::$n`), which PHP treats as the same,
  are not recognized as equivalent. custos folds the case of those names
  (Detection, "Name case").
- **Targets that write (custos diverges).** `$m[$i++] = $m[$i++] + 1`
  evaluates the target twice (two increments, two different elements);
  `$m[$i++] += 1` evaluates it once. custos does not report a target
  containing an assignment or `++`/`--` (outside closures). Calls in the
  target are still reported: there they are almost always getters.
- **Chains of `+` and `*` (custos diverges).** Upstream rewrites
  `$x = $x + 1 + 1` to `$x += 1 + 1`, which regroups the operation and can
  change a float result or turn an integer into a float on overflow. custos
  still reports the chain but offers no fix (F2).
- **Fragments that may write the target (custos diverges).** Upstream
  rewrites whatever the fragments contain. custos withholds the fix when the
  moment the target is read differs between the two forms and a fragment may
  have an effect (F2); the report stays.

Recorded in `internal/testing/conformance/testdata/ea-divergences.json`.
