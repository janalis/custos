---
id: NestedAssignmentsUsage
group: Code style
kind: syntax
needs: []
php: { min: "", max: "" }
---

# NestedAssignmentsUsage

## Summary

Chained assignments such as `$x = $y = 5` pack several writes into one
expression; they are easy to misread and a typo (`=` instead of `+`/`==`)
silently becomes another assignment. Separate statements are clearer.

## Detection

- D1: a plain assignment `L = R` (operator `=`, including by-reference `= &`)
  whose right-hand side `R` is itself an assignment expression (directly, not
  through parentheses). `R` may be a plain, by-reference or compound assignment
  (`+=`, `.=`, `??=`, …).
- D2: only the outermost assignment of a chain is reported: an assignment whose
  direct parent is itself an assignment expression (plain or compound) is not
  reported (its chain is reported via the parent, or not at all if the parent is
  compound — see E2).
- The context does not matter for detection: statement level, conditions,
  `return`, call arguments, array elements, etc. are all reported.

## Exceptions (no report)

- E1: the right-hand side is a parenthesised assignment, e.g. `$p = ($q = 1)`.
- E2: the outermost operator is compound (`$t .= $u = 'x'`): compound
  assignments are never the reported node, and the inner `$u = 'x'` is skipped
  by D2 because its parent is an assignment.
- E3: list/array destructuring assignments (`[$a, $b] = …`) are never the
  reported node.

## Report

- Range: the whole outermost assignment expression, from the start of the first
  left-hand side to the end of the innermost value (excluding a trailing `;`).
- Severity: info (weak warning).
- Message: "Split this chained assignment into separate assignments."

## Fix

Available only when the reported assignment is the entire expression of an
expression statement (its parent is a statement `…;`). In any other context
(condition, `return`, argument, …) the issue is reported without a fix.

Let the chain be `V1 = V2 = … = Vn = E` (n ≥ 2 targets, `E` the innermost value).

- F1: decide whether `E` is "simple": a plain variable (not `$$name` or
  `${expr}`), a constant reference (`FOO`, `true`, `null`, …), a class constant
  reference on a class name (`A::B`, `A::class`; not `make()::B` or `$o::B`),
  a number literal (optionally negated: `-3`), or a string literal without
  interpolation (single-quoted, or double-quoted/heredoc without embedded
  variables/expressions).
- F2: if `E` is simple, the statement is replaced by n statements, innermost
  target first:
  `Vn = E;` `Vn-1 = E;` … `V1 = E;`
- F3: otherwise (calls, arrays, expressions, interpolated strings, …) the first
  statement assigns `E` to the innermost target and every following statement
  copies the innermost target:
  `Vn = E;` `Vn-1 = Vn;` … `V1 = Vn;`
- Each new statement is `<target text> = <value text>;` (single spaces around
  `=`) and is placed on its own line at the original statement's position and
  indentation; the original statement is removed. Target and value texts are
  copied verbatim.
- When the statement is closed by `?>` instead of `;`, only the expression is
  replaced: the closing tag and what follows it stay as they are.
- F4: no fix when any link of the chain is compound or by-reference, or the
  chain contains a destructuring assignment (see Divergences).
- F5: no fix unless the split is provably equivalent (see Divergences):
  - the outermost target `V1` is a plain variable, or a chain of
    property/element/static-property accesses on one (or on a class name) with
    identifier member names and literal, constant or plain-variable keys,
    optionally ending in an append `[]`;
  - every target between the outermost and the innermost is a plain variable;
  - the innermost target `Vn` is, for F2, a plain variable or a direct element
    of a superglobal (`$_SESSION['k']`) — any target when `E` is `null`; for
    F3, a target that can be read back (same shape as `V1`, without `[]`).

## Options

None.

## PHP versions

No gating.

## Examples

```php
<?php
function setup() {
    <weak_warning descr="Split this chained assignment into separate assignments.">$width = $height = 64</weak_warning>;
    <weak_warning descr="Split this chained assignment into separate assignments.">$first = $last = "none"</weak_warning>;
    <weak_warning descr="Split this chained assignment into separate assignments.">$p = $q = $r = make_point(1, 2)</weak_warning>;
    <weak_warning descr="Split this chained assignment into separate assignments.">$left = $right = $origin</weak_warning>;
    <weak_warning descr="Split this chained assignment into separate assignments.">$lo = $hi = -1</weak_warning>;

    if (<weak_warning descr="Split this chained assignment into separate assignments.">$m = $n = fetch()</weak_warning>) {}   // reported, no fix

    $solo = 3;            // not chained
    $k = ($j = 2);        // parenthesised: not reported
    $acc .= $tmp = 'x';   // outer compound: not reported
}
```

```php
<?php
function setup() {
    $height = 64;
    $width = 64;
    $last = "none";
    $first = "none";
    $r = make_point(1, 2);
    $q = $r;
    $p = $r;
    $right = $origin;
    $left = $origin;
    $hi = -1;
    $lo = -1;

    if ($m = $n = fetch()) {}   // reported, no fix

    $solo = 3;            // not chained
    $k = ($j = 2);        // parenthesised: not reported
    $acc .= $tmp = 'x';   // outer compound: not reported
}
```

## Divergences

- F4: upstream offers the fix also for chains containing a compound or
  by-reference inner link, producing plain `=` statements that change
  semantics (`$a = $b += 1` would become `$b = 1; $a = 1;`, and `= &` loses the
  reference). We report but offer no fix in those cases. Not covered by
  upstream fixtures.
- **Re-readable innermost target (custos diverges).** For a non-simple
  value the fix makes the outer targets read the innermost one back. That
  is only equivalent when the read has no side effect and returns the
  stored value: a plain variable, or property/element accesses on one with
  identifier names and literal, constant or plain-variable keys. Otherwise
  (`$t = $list[] = f()`, `$a = $m[$i++] = f()`, `$m[g()]`, dynamic names,
  method-call receivers) the report has no fix (Composer: the fix produced
  `$target = $this->filesToCleanup[];`, a fatal error).
- **Converting inner targets (custos diverges).** An assignment expression
  yields what its target ended up holding, which is not always the value
  written: a typed property converts it (`$shown = $this->ratio = 0` gives
  `0.0` for a `float` property) and a string offset keeps one character
  (`$t = $s[0] = 'xyz'` gives `'x'`). Copying a simple value to every target
  would hand the outer targets the unconverted value, so with a simple value
  the fix needs a plain-variable innermost target (or a superglobal element,
  or the value `null`, which is never converted), and in every chain the
  targets between the outermost and the innermost must be plain variables.
  Other chains are reported without a fix.
- **Evaluation order of the outermost target (custos diverges).** PHP
  evaluates a non-trivial key or receiver of an outer target before the inner
  assignments, while the split statements run it after them
  (`$x[$i++] = $y[$i++] = 0`, `$a[$o->p] = $o->p = 5`). The fix is offered
  only when the outermost target is built from plain variables, identifier
  member names and literal, constant or plain-variable keys, which PHP reads
  after the inner assignments anyway.
- **Values evaluated once (custos diverges).** `${name()}` and `make()::N`
  are not simple values: repeating them per target would run the call again.
  They use the copy-from-innermost form (F3).
- **Close-tag terminated statements (custos diverges).** `<?php $a = $b = 1 ?>`
  keeps its `?>`; replacing the whole statement would drop it and turn the
  following markup into PHP code.
- **Brace-less bodies (custos).** When the chained assignment is the whole
  body of a brace-less `if`/`elseif`/`else`/loop/`declare`
  (`if ($c) $n = $_SESSION['k'] = 10; else …`, adodb), the split
  statements are wrapped in braces; otherwise the second statement would
  leave the body and detach the `else` (parse error).
