---
id: ArrayUniqueCanBeUsed
group: Control flow
kind: syntax
needs: []
php: { min: "7.2", max: "" }
---

# ArrayUniqueCanBeUsed

## Summary

Counting occurrences with `array_count_values()` just to take the keys or the
number of keys is an indirect way of deduplicating. Since PHP 7.2
`array_unique()` is fast and says what is meant: `array_values(array_unique($a))`
or `count(array_unique($a))`.

## Detection

D1. Only active when the configured PHP language level is **7.2 or higher**.

D2. Node: a plain function call (not a method/static call) whose name part is
    `array_count_values`, compared case-insensitively, that resolves to the
    global built-in (unqualified with no same-named function declared or
    imported in the current namespace, or `\`-qualified; `Ns\…` and
    shadowing user functions do not count), with **exactly one** argument. Let *A* be the verbatim source text of that
    argument.

D3. The call is directly an argument of an enclosing call *O* (the call's
    parent is an argument list, whose parent is *O*), and *O* is a plain
    function call (not a method/static call, not `new`). *O*'s name part
    (case-insensitive; *O* must resolve to the global function the same way)
    is:
    - D3a. `array_keys` → replacement `array_values(array_unique(A))`;
    - D3b. `count` → replacement `count(array_unique(A))`.
    Upstream checks neither the position of the inner call inside *O*'s
    argument list nor *O*'s argument count (see Divergences).

## Exceptions (no report)

E1. PHP level below 7.2 (this includes the default level used by fixtures
    without an explicit version, which is below 7.1).
E2. `array_count_values()` with zero or two-plus arguments.
E3. Enclosing call is anything other than `array_keys`/`count` (e.g.
    `sizeof`, `array_flip`, `max`), or a method call (`$c->count(...)`).
E4. The inner call is not a direct argument: wrapped in parentheses
    (`count((array_count_values($a)))`), part of an expression
    (`count(array_count_values($a) + $b)`), assigned to a variable first.
E5. (removed: any casing matches, `Count(ARRAY_COUNT_VALUES($a))` is
    reported.)

## Report

- Range: the enclosing call *O*, from the start of its name (including any
  namespace qualifier) to its closing `)`.
- Severity: info (fixture markup `weak_warning`).
- Message: `Use '{replacement}' instead (array_unique() is fast since PHP 7.2).`
  `{replacement}` is the unqualified form from D3 (no leading `\`).

## Fix

F1. Replace the whole enclosing call *O* with the replacement text from D3
    (exactly `array_values(array_unique(A))` or `count(array_unique(A))`, no
    spaces added, *A* verbatim, function names in lower case). A leading
    global qualifier is preserved: when *O* is written `\count`/`\array_keys`,
    the outer emitted function (`count`/`array_values`) gets `\`; when the
    inner call is written `\array_count_values`, `array_unique` gets `\`.
    `\count(\array_count_values($a))` → `\count(\array_unique($a))`. Other
    (namespaced) qualifiers are not carried over. Each emitted name also
    gets `\` when an unqualified call to it at that position would not
    reach the global function (a `use function` import under that name, or a same-named function declared in the current namespace):
    in `namespace S; function array_unique() {…}`,
    `array_keys(array_count_values($v))` → `array_values(\array_unique($v))`.

## Options

None.

## PHP versions

Requires language level ≥ 7.2. EA's fixture runs at 7.2.

## Examples

(PHP level 7.2)

```php
<?php
function tally(array $votes, $batch) {
    $names = <weak_warning descr="Use 'array_values(array_unique($votes['names']))' instead (array_unique() is fast since PHP 7.2).">array_keys(array_count_values($votes['names']))</weak_warning>;
    $total = <weak_warning descr="Use 'count(array_unique(load($batch)))' instead (array_unique() is fast since PHP 7.2).">\count(\array_count_values(load($batch)))</weak_warning>;

    $flip  = array_flip(array_count_values($votes['names']));
    $pair  = count(array_count_values($votes['names'], $batch));
    $wrap  = count((array_count_values($votes['names'])));
    $obj   = $batch->count(array_count_values($votes['names']));
    return [$names, $total, $flip, $pair, $wrap, $obj];
}
```

```php
<?php
function tally(array $votes, $batch) {
    $names = array_values(array_unique($votes['names']));
    $total = \count(\array_unique(load($batch)));

    $flip  = array_flip(array_count_values($votes['names']));
    $pair  = count(array_count_values($votes['names'], $batch));
    $wrap  = count((array_count_values($votes['names'])));
    $obj   = $batch->count(array_count_values($votes['names']));
    return [$names, $total, $flip, $pair, $wrap, $obj];
}
```

## Divergences

- Upstream does not check that the inner call is the **first and only**
  argument of `array_keys`/`count`: `array_keys(array_count_values($a), 2)`
  (keys with value 2 — different semantics) and
  `count(array_count_values($a), COUNT_RECURSIVE)` are reported and the fix
  silently drops the extra argument; `array_keys($x, array_count_values($a))`
  (inner call as the search value) is reported too. Recommendation: only
  report when *O* has exactly one argument (the inner call). No upstream
  fixture covers these shapes.
- custos diverges: upstream matches both function names case-sensitively,
  missing `Count(ARRAY_COUNT_VALUES($a))`. PHP function names are
  case-insensitive, so custos matches any casing (D2, D3).
- custos diverges: upstream ignores namespaces, so a user function named
  `count`, `array_keys` or `array_count_values` declared in the current
  namespace (or called as `Other\array_count_values`) was reported and
  rewritten to the global built-ins. custos requires both calls to resolve
  to the global functions (D2, D3).
- custos diverges: upstream's fix drops a leading `\` from the calls
  (`\count(...)` → `count(...)`). In namespaced code that turns a global call
  into a namespace-relative one (and loses the fully-qualified style); custos
  keeps the global qualifier on the corresponding emitted call (F1).
- **Builtin spelling (custos diverges).** Upstream inserts `array_values`
  and `array_unique` with only the original qualifiers, so a namespaced or
  imported function of that name captures the rewritten call. custos adds
  `\` in that case (F1).
