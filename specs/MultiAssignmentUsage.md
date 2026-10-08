---
id: MultiAssignmentUsage
group: Control flow
kind: syntax
needs: []
php: { min: "", max: "" }
---

# MultiAssignmentUsage

## Summary
Two patterns where destructuring would be clearer:
1. unpacking the `foreach` value variable with `list(...) = $value` as a
   separate statement, when `foreach (... as list(...))` can do it directly
   (PHP 5.5+);
2. consecutive statements pulling numbered elements out of the same array
   one at a time (`$a = $row[0]; $b = $row[1];`), which a single
   `list(...) = $row` replaces.

## Detection

### Part A — destructuring a foreach variable
- **D1** Project language level is PHP 5.5 or higher.
- **D2** A destructuring assignment whose first token is `list` or `[`
  (`list($a, $b) = …` or `[$a, $b] = …`), and which is itself a whole
  expression statement (its parent is the statement `…;`; not nested in a
  condition, argument, etc.).
- **D3** Its right-hand side (after unwrapping a bare expression wrapper) is
  a plain variable `$v` (not `$v[0]`, not a call, not a property).
- **D4** (custos refinement, see Divergences) The statement is a **direct**
  statement of a `foreach` body (depth 1): its parent is the `foreach`'s
  braced block, its alternative-syntax statement list
  (`foreach (…): … endforeach;`), or it is the brace-less single body
  statement. That `foreach` declares a variable named like `$v` (key, value,
  or a variable inside a destructuring value). Statements nested in an `if`,
  `switch`, `try`, inner loop, plain `{ … }` block, etc. inside the loop are
  not reported.

- **D4c** (custos refinement, see Divergences) The loop body does not
  mention `$v` before the destructuring (e.g. `$v[2] ??= null;`): the
  header form would destructure the unmodified value.

### Part B — consecutive numbered element reads
- **D5** An assignment (`=`) whose left side is a plain variable and which is
  itself a whole expression statement.
- **D6** The previous statement in the same statement list — skipping
  comments and docblocks (any number of them) — is also an expression
  statement whose expression is a plain `=` assignment (left side may be
  anything: variable, array element, property).
- **D7** Both right-hand sides have the shape `X[N]`, where `N` is a number
  literal (integer or float, optionally preceded by a unary minus, e.g.
  `-1`) and the index is present, and the two bases `X` are structurally equal
  (same text ignoring whitespace; for variables, same name).
- **D7a** The two indexes denote **different** array keys. The key of `N` is
  its integer value (decimal, hex, octal, binary; `_` separators allowed),
  negated when preceded by `-`; a float literal is truncated toward zero, as
  PHP does for float keys (`$r[1.0]` and `$r[1]` are the same key).
  `$p = $r[0]; $q = $r[0];` reads one element twice — there is nothing to
  destructure — and is not reported. The indexes need not be consecutive:
  `list()` can skip positions and keyed destructuring takes any keys.
- **D7b** The base must denote the same array in both statements and be
  free of side effects (custos diverges, see Divergences):
  - `X` contains no call or other effectful expression (function, method or
    static call, `new`, assignment, `++`/`--`, `include`/`require`, `eval`,
    `yield`, `exit`, `print`): `load()[0]; load()[1]` calls `load()` twice,
    a destructuring would call it once;
  - the previous assignment's target, with trailing `[...]` element writes
    stripped (`$x[1]` → `$x`, `$o->p[2]` → `$o->p`), does not occur inside
    `X` (same node kind, same text ignoring whitespace):
    `$row = $row[0]; $b = $row[1];` reads `$row[1]` from the new value, and
    `$o = $o->list[0]; $p = $o->list[1];` likewise. A target that does not
    occur in `X` is fine (`$this->a = $this->items[0]; $b = $this->items[1];`).
- **D8** Report the later assignment (D5). In a run of three or more such
  statements every statement after the first is reported.

## Exceptions (no report)
- **E1** Part A below PHP 5.5.
- **E2** Destructuring of something other than a plain variable, or of a
  variable that is not declared by an enclosing `foreach` within the same
  function (the boundary is a function/closure, so a closure inside the loop
  stops the search).
- **E2b** Destructuring nested deeper than the loop body's own statement
  list (`if (is_array($row)) { [$a, $b] = $row; }`): moving it into the
  header would drop the guard.
- **E3** Destructuring that is not a statement of its own
  (`while (list($a) = $v)`, `f(list($a) = $v)`).
- **E4** String/variable/constant indexes (`$r['id']`, `$r[$k]`, `$r[ONE]`) or
  empty index `$r[]`.
- **E5** Current statement assigning into a non-variable
  (`$out[1] = $src[0];`), even if the previous one matches.
- **E6** Previous statement not an assignment, or a different base array.
- **E6b** The base contains a call or other side effect
  (`$a = make()[0]; $b = make()[1];`), or the previous statement writes the
  base (or something it is built from) before the second read
  (`$row = $row[0]; $b = $row[1];`, `$cfg[1] = $cfg[0]; $x = $cfg[2];`).
- **E7** Both statements read the same key (`$a = $r[1]; $b = $r[1];`,
  `$a = $r[1]; $b = $r[1.0];`).

## Report
- Range:
  - Part A: the whole destructuring assignment expression, from `list` / `[`
    to the end of the right-hand side, **excluding** the `;`.
  - Part B: the whole later assignment expression from the left variable to
    the closing `]` of the right side, excluding the `;`.
- Severity: info (weak warning).
- Messages: Part A `Destructure directly in the foreach header.`;
  Part B `Use one destructuring assignment from '{base}' instead.` where
  `{base}` is the source text of `X`.

## Fix
None.

## Options
None.

## PHP versions
- Part A requires PHP ≥ 5.5 (upstream test runs at the IDE default level,
  which is above 5.5).
- Part B: no gating (short `[…] =` syntax would need 7.1, but `list()` works
  everywhere).

## Examples

```php
<?php
foreach ($pairs as $idx => $pair) {
    <weak_warning descr="Destructure directly in the foreach header.">[$left, $right] = $pair</weak_warning>;
    if (is_array($pair)) {
        [$top, $bottom] = $pair;          // E2b: nested under a guard
    }
    list($m, $n) = $other;                // E2: not a loop variable
    $fn = function () use ($pair) { list($u) = $pair; };   // E2: closure boundary

    $width = $pair[0];
    // a plain comment
    /** @var int $height */
    <weak_warning descr="Use one destructuring assignment from '$pair' instead.">$height = $pair[1]</weak_warning>;
    <weak_warning descr="Use one destructuring assignment from '$pair' instead.">$depth = $pair[-1]</weak_warning>;

    $name = $pair['name'];
    $kind = $pair['kind'];                // E4
    $copy = [];
    $copy[0] = $pair[1];
    $copy[1] = $pair[0];                  // E5
}
```

```php
<?php
foreach ($records as $key => $record):
    <weak_warning descr="Destructure directly in the foreach header.">list($id, $label) = $record</weak_warning>;
    try {
        [$first] = $record;               // E2b: inside try
    } catch (\Throwable $t) {}
    foreach ($record as $cell) {
        [$x, $y] = $record;               // E2b: direct statement of the inner loop, which does not declare $record
    }
endforeach;
foreach ($records as $record) <weak_warning descr="Destructure directly in the foreach header.">[$id, $label] = $record</weak_warning>;
```

## Divergences
- **Depth-1 only (D4/E2b) — custos refinement, not upstream.** Upstream
  reports a destructuring of the loop variable anywhere inside the loop
  (under `if`, inner loops, …), so code like
  `if (is_array($t)) { [$a, $b] = $t; }` is told to move the destructuring
  into the header, which would drop the type guard (found on real code).
  custos only reports statements directly in the loop body. The upstream
  fixture's report is a direct body statement, so conformance is unaffected.
  Recorded in `docs/internals/decisions.md` ("Spec-level false positives").
- Upstream visits only plain assignments for Part B; whether compound
  (`.=`) or by-reference (`=&`) current statements are included depends on the
  host parser's node kinds. Recommendation: require plain `=` (by-reference
  `= &$r[1]` treated as plain too). No fixture coverage.
- Repeated index (custos diverges from upstream). Upstream does not compare
  the two indexes, so `$p = $r[0]; $q = $r[0];` — the same element copied
  into two variables — is told to use one destructuring assignment, which
  cannot express it. custos requires distinct keys (D7a). Non-consecutive
  distinct indexes are still reported (`list()` can skip positions, keyed
  destructuring takes any keys).
- **Base changed or re-evaluated (custos diverges).** Upstream only compares
  the two base texts. When the first statement overwrites the base
  (`$row = $row[0]; $b = $row[1];`) the second read sees a different array,
  and when the base is a call (`load()[0]; load()[1]`) each statement
  evaluates it anew; a single destructuring assignment would change the
  result or the number of calls in both cases. custos skips these (D7b).
- **Other uses of the value variable (D4c) — custos refinement, not
  upstream.** Upstream suggests destructuring in the header even when the
  body fills in a default first (`$m[2] ??= null; [$a, $b, $c] = $m;`, found
  on Nextcloud); moving the destructuring into the header would lose the
  default. custos only reports when the destructuring is the first mention
  of the variable in the body.
- **By-reference assignments (custos diverges from D5).** `$q =& $m[6];
  $t =& $m[7];` binds references (missing keys are created silently, writes
  go through to the array); destructuring would copy the values and warn on
  missing keys, so by-reference assignments are not paired.
