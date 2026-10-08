---
id: IssetConstructsCanBeMerged
group: Control flow
kind: syntax
needs: []
php: { min: "", max: "" }
---

# IssetConstructsCanBeMerged

## Summary

`isset()` accepts several arguments and is true only when all of them are set.
A chain such as `isset($p) && isset($q)` is therefore the same as
`isset($p, $q)`, and `!isset($p) || !isset($q)` the same as `!isset($p, $q)`.
Merging keeps conditions short.

## Detection

Only the operators `&&` and `||` are concerned (the word operators `and` /
`or` are **not**).

- **D1 — chain root.** Visit every binary expression whose operator is `&&`
  or `||`. Let `P` be its parent; if `P` is a parenthesized expression, let
  `C` be `P`'s parent, otherwise `C = P` (only **one** level of parentheses
  is looked through). If `C` is a binary expression with the **same**
  operator, skip this node (it is an inner part of a larger chain).
- **D2 — fragments.** Flatten the chain into an ordered list of fragments:
  for the left then the right operand, strip any number of enclosing
  parentheses; if the stripped operand is a binary expression with the same
  operator, recurse into it, otherwise append the stripped operand as one
  fragment. A binary expression with a different operator (e.g. `$m || $n`
  inside an `&&` chain) is one fragment. Proceed only when there are at least
  two fragments.
- **D3 — `&&` chains.** Walk the fragments left to right; the first fragment
  that is an `isset(...)` construct is the *first hit*; the next fragment that
  is an `isset(...)` construct is the *second hit*. When a second hit exists,
  report it (once) and stop scanning this chain. A negated `!isset(...)` is
  not an `isset` fragment in this mode.
- **D4 — `||` chains.** Same walk, but a fragment qualifies when it is a
  unary expression whose direct operand (no parentheses stripped) is an
  `isset(...)` construct, i.e. `!isset(...)`. First and second qualifying
  fragments are the hits; report the second (once) and stop. Upstream does
  not check which unary operator it is — see Divergences.
- **D5** Isset constructs already holding several arguments qualify the same
  way (`isset($a, $b) && isset($c)` is reported).
- **D6** The hits need not be adjacent: other fragments may sit between them
  (`isset($a) && $ok && isset($b)` is reported).
- **D7** A fragment between the two hits must not have side effects (a
  function/method/static call, `|>`, `new`, `clone`, an assignment, `++`/`--`,
  `include`/`require`, `eval`, `exit`, `print`, `throw`, `yield`, a backtick
  shell command; closure bodies are not inspected). When the walk meets such
  a fragment after a first hit, the first hit is forgotten and the walk
  continues, so a later qualifying construct becomes the new first hit.
  Merging across a side effect would move the second check before it, so the
  side effect could be skipped, or could change what the second check sees
  (`isset($a) && ($b = load()) && isset($b)`).

At most one report is produced per chain root, even when it contains three or
more qualifying constructs; the remaining ones are only seen after the fix is
applied and the file re-analysed.

## Exceptions (no report)

- **E1** Mixed polarity: `isset($a) && !isset($b)`, `isset($a) || !isset($b)`,
  `!isset($a) && !isset($b)`, `isset($a) || isset($b)`.
- **E2** A single qualifying construct in the chain.
- **E3** Chains built with `and` / `or`.
- **E4** `!(isset($a)) || ...`: the operand of `!` is a parenthesized
  expression, not an `isset`, so it does not qualify.
- **E5** D7: the only two qualifying constructs are separated by a fragment
  with side effects (`isset($a) && notify() && isset($b)`).

## Report

- Range:
  - `&&` chain: the whole second `isset(...)` construct, from the `isset`
    keyword to its closing `)`.
  - `||` chain: the second `isset(...)` construct **without** the leading `!`
    (from `isset` to `)`).
- Severity: info (weak warning).
- Messages (our wording):
  - `&&`: `Merge this check into the preceding isset() call.`
  - `||`: `Merge this check into the preceding !isset() call.`

## Fix

- **F1** Replace the **chain root** binary expression (D1 node, without any
  parentheses wrapping it) with new text built as follows:
  1. *Merged construct*: the arguments of the first hit followed by the
     arguments of the second hit, each argument's source text verbatim,
     joined with `", "`, wrapped as `isset(...)` (`&&`) or `!isset(...)`
     (`||`).
  2. *Fragments*: the D2 fragment list in original order with the second hit
     removed (for `||` the removed fragment is the `!isset(...)` unary
     expression) and the first hit replaced by the merged construct. Each
     other fragment contributes its source text, except that when the
     fragment's **immediate** parent is a parenthesized expression, the text
     of that parenthesized expression is used instead (exactly one level of
     parentheses restored).
  3. Join the fragments with `" && "` (`&&`) or `" || "` (`||`).
  No extra parentheses are added around the result.
- Consequences (must be reproduced):
  - `isset($a) && isset($b)` → `isset($a, $b)`.
  - `isset($a) && (isset($b) && isset($c))` → `isset($a, $b) && isset($c)`.
  - `isset($a) && isset($b) && ($m || $n)` → `isset($a, $b) && ($m || $n)`.
  - The merged construct stays where the first hit was, so the evaluation
    order of the other fragments is unchanged:
    `$ok && isset($a) && isset($b)` → `$ok && isset($a, $b)`.
  - Parentheses that grouped same-operator sub-chains are dropped; original
    spacing between fragments is normalised to single spaces around the
    operator; comments located between fragments are lost.

## Options

None.

## PHP versions

No gating.

## Examples

```php
<?php
function probe(array $cfg, $row, $extra) {
    $r1 = isset($cfg['host']) && <weak_warning descr="Merge this check into the preceding isset() call.">isset($cfg['port'])</weak_warning>;
    $r2 = isset($row->id, $row->name) && <weak_warning descr="Merge this check into the preceding isset() call.">isset($extra)</weak_warning>;
    $r3 = $extra > 2 && isset($cfg['a']) && (<weak_warning descr="Merge this check into the preceding isset() call.">isset($cfg['b'])</weak_warning> && isset($cfg['c']));
    $r4 = isset($cfg['x']) && <weak_warning descr="Merge this check into the preceding isset() call.">isset($cfg['y'])</weak_warning> && ($extra || $row);
    $r5 = !isset($row->id) || !<weak_warning descr="Merge this check into the preceding !isset() call.">isset($row->name)</weak_warning> || $extra;

    $n1 = isset($cfg['host']) && !isset($cfg['port']);
    $n2 = isset($cfg['host']) || isset($cfg['port']);
    $n3 = !isset($cfg['host']) && !isset($cfg['port']);
    $n4 = isset($cfg['host']) and isset($cfg['port']);
    return [$r1, $r2, $r3, $r4, $r5, $n1, $n2, $n3, $n4];
}
```

```php
<?php
function probe(array $cfg, $row, $extra) {
    $r1 = isset($cfg['host'], $cfg['port']);
    $r2 = isset($row->id, $row->name, $extra);
    $r3 = $extra > 2 && isset($cfg['a'], $cfg['b']) && isset($cfg['c']);
    $r4 = isset($cfg['x'], $cfg['y']) && ($extra || $row);
    $r5 = !isset($row->id, $row->name) || $extra;

    $n1 = isset($cfg['host']) && !isset($cfg['port']);
    $n2 = isset($cfg['host']) || isset($cfg['port']);
    $n3 = !isset($cfg['host']) && !isset($cfg['port']);
    $n4 = isset($cfg['host']) and isset($cfg['port']);
    return [$r1, $r2, $r3, $r4, $r5, $n1, $n2, $n3, $n4];
}
```

## Divergences

- **Unary operator not checked (upstream bug).** In `||` chains upstream
  accepts any unary expression wrapping `isset` (e.g. a cast
  `(bool) isset($a)` or `@`-like wrappers the parser models as unary) and the
  fix emits `!isset(...)`, changing semantics. Recommendation: only accept
  the logical-not operator `!`.
- **Double parentheses around a sub-chain.** D1 looks through one level of
  parentheses only, while D2 strips any number; so in
  `isset($a) && ((isset($b) && isset($c)))` the inner chain is also treated as
  a chain root and may produce a second, overlapping report/fix.
  Recommendation: look through all parentheses in D1 too. No fixture covers
  it.
- **Reordering (custos diverges from upstream).** Upstream moves the merged
  construct to the front of the chain and accepts side-effecting fragments
  between the hits. Both can change behaviour: fragments before the first hit
  (`load($o) && isset($a) && isset($b)`) would no longer run first, and a
  call or assignment between the hits could be skipped or could change what
  the moved check sees. custos keeps the merged construct at the first hit's
  position (F1) and never merges across a side-effecting fragment (D7).
