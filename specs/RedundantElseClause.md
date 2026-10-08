---
id: RedundantElseClause
group: Control flow
kind: syntax
needs: []
php: { min: "", max: "" }
---

# RedundantElseClause

## Summary

When an `if` body always leaves the current flow (return, throw, break,
continue, exit/die), the following `else`/`elseif` keyword is superfluous: its
code can simply follow the `if`. Removing it flattens nesting.

## Detection

Visit every `if` statement `I`.

- **D1** `I` is not the nested `if` of an `else if` (i.e. its direct parent is
  not an `else` clause). Only the head of a chain is examined.
- **D2** `I`'s own body is a braced block: it starts with `{`. Single-statement
  bodies (`if ($c) return;`) and alternative syntax (`if ($c): … endif;`) are
  ignored.
- **D3** `I` has at least one alternative branch (`elseif` or `else`). The
  *first alternative* `A` is the first `elseif` if there is any, otherwise the
  `else`.
- **D4** Shape of `A`:
  - `elseif`: its body must be a block (braced); `elseif ($c) foo();` → no
    report;
  - `else`: its body must be a block, or be an `if` statement (`else if …`,
    whose own body may have any form); `else foo();` / `else ;` → no report.
- **D5** The last statement of `I`'s body (ignoring trailing comments and
  docblocks) is one of:
  - `return …;`
  - `throw …;`
  - `break …;`
  - `continue …;`
  - an expression statement whose expression *starts with* `exit`/`die`
    (`exit;`, `die;`, `exit(2);`, `die('x');`).
  An empty body or a body ending with any other statement (including a nested
  block or `if` that itself returns) is not reported.

## Exceptions (no report)

- **E1** The `if` inside `else if` (D1); its predecessor chain is judged only
  from the head `if`.
- **E2** Non-braced or alternative-syntax `if` bodies.
- **E3** First alternative with a non-block body (other than `else if`).
- **E4** `if` body not ending in a flow-leaving statement.
- **E5** Only the first alternative is ever considered: later `elseif`/`else`
  branches are not reported separately.

## Report

- Range: the keyword token of `A`: `else` (also the `else` of `else if`) or
  `elseif`. Exactly that keyword, nothing else.
- Severity: warning.
- Messages (our wording):
  - `else`: `Drop the 'else' and move its code after the 'if'.`
  - `elseif`: `Turn this 'elseif' into a separate 'if'.`

## Fix

Let `I` be `if (COND1) BODY1 A …`.

- **F1** `A` is `else { … }` (block body): remove `A` entirely together with
  the whitespace between `BODY1`'s closing `}` and the `else` keyword. Then
  insert the block's inner content (text between its braces, with leading and
  trailing whitespace trimmed) right after `}`, separated by a newline.
  Exception required for conformance: when the inner content begins with an
  empty statement `;`, no separator is inserted (upstream's formatter glues a
  `;` to the preceding `}`): `if ($c) { die; }  else { ; }` →
  `if ($c) { die; };`. An empty block (`else {}`) inserts nothing.
  Example: `if ($ok) { return 1; } else { log(); return 2; }` →
  `if ($ok) { return 1; }` + newline + `log(); return 2;`.
- **F2** `A` is `else if (…) …` (else followed by an `if`): remove the `else`
  clause and insert, after `I` and a newline, the nested `if` statement
  verbatim, including its own `elseif`/`else` branches.
  `if ($c) { return; } else if ($d) { a(); } else { b(); }` →
  `if ($c) { return; }` + newline + `if ($d) { a(); } else { b(); }`.
- **F3** `A` is `elseif (COND2) BODY2` followed by optional more branches
  `REST`: the statement becomes two statements separated by a newline:
  1. `if (COND1) BODY1` — rebuilt as the text `if (` + COND1 + `) ` + BODY1
     (BODY1 verbatim, condition text verbatim, exactly one space before `(`
     and before the body);
  2. the original `if` with its condition replaced by COND2's text, its body
     replaced by BODY2's text verbatim, and the `elseif` clause removed, so any
     `REST` branches stay attached: `if (COND2) BODY2 REST`. Header spacing of
     this second `if` is the original `I`'s spacing (`if`, whitespace, `(`,
     `)`, whitespace before the body).
  Example: `if ($a) { return 0; } elseif ($b)  { x(); }  elseif($c) { y(); }` →
  `if ($a) { return 0; }` newline `if ($b) { x(); } elseif($c) { y(); }`.
- Fix output is compared whitespace-collapsed; the newline separators only need
  to be some whitespace, except for F1's `};` rule.

## Options

None.

## PHP versions

No gating.

## Examples

```php
<?php
function route($req, $list) {
    if ($req === null) { exit(3); } <warning descr="Drop the 'else' and move its code after the 'if'.">else</warning> { ; }
    if (!$list) { throw new LogicException('none'); }
    <warning descr="Drop the 'else' and move its code after the 'if'.">else</warning> { ; }
    foreach ($list as $item) {
        if ($item < 0) { continue; } <warning descr="Drop the 'else' and move its code after the 'if'.">else</warning> if ($item > 9) { break; }
    }
    if ($req === 'a') { return 1; }
    <warning descr="Turn this 'elseif' into a separate 'if'.">elseif</warning> ($req === 'b') { $list = []; } else { $list = [1]; }
    if ($req === 'c') { $list[] = 2; } else { ; }
    if ($req === 'd') return 4; else ;
    if ($req === 'e') { return 5; } else $list = [];
    if ($req === 'f') { ; } else if ($req === 'g') { return 6; } else { ; }
    if ($req === 'h'):
        return 7;
    else:
        $list = null;
    endif;
    return $list;
}
```

```php
<?php
function route($req, $list) {
    if ($req === null) { exit(3); };
    if (!$list) { throw new LogicException('none'); };
    foreach ($list as $item) {
        if ($item < 0) { continue; }
        if ($item > 9) { break; }
    }
    if ($req === 'a') { return 1; }
    if ($req === 'b') { $list = []; } else { $list = [1]; }
    if ($req === 'c') { $list[] = 2; } else { ; }
    if ($req === 'd') return 4; else ;
    if ($req === 'e') { return 5; } else $list = [];
    if ($req === 'f') { ; } else if ($req === 'g') { return 6; } else { ; }
    if ($req === 'h'):
        return 7;
    else:
        $list = null;
    endif;
    return $list;
}
```

## Divergences

- F1 separator: upstream inserts a newline and then lets the IDE formatter
  adjust it, which removes the space before a leading `;`. Recommendation:
  newline separator, none when the moved content starts with `;` (needed to
  match the upstream fixture). Re-indentation of moved statements is not
  required (comparison is whitespace-collapsed).
- F3 normalises the first `if`'s header to `if (COND1) ` (upstream rebuilds it
  from a template). Recommendation: follow upstream so `if($a){…}` becomes
  `if ($a) {…}`.
- Comments between `}` and `else`/`elseif`: upstream behaviour unverified
  (no fixture). Recommendation: preserve them (keep them before the moved code).
- **Code that cannot move (custos diverges).** No quick-fix when the `if`
  is the unbraced body of another `if` or a loop (the moved code would run
  unconditionally, or attach to the outer statement), or when the `else`
  block declares a function or class directly (PHP hoists an
  unconditional top-level declaration: `if (function_exists('f')) {
  return; } else { function f() {} }` would become "Cannot redeclare").
  The finding is still reported.
