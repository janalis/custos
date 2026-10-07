---
id: CascadeStringReplacement
group: Performance
kind: semantic
needs: [types, stubs, index]
php: { min: "", max: "" }
---

# CascadeStringReplacement

## Summary
`str_replace()` accepts arrays of searches and replacements and applies them
in order, so several consecutive replacements on the same subject — written
as back-to-back assignments or as nested calls — can be one call. Also, a
search array whose items are all the same string literal can be that string.

## Detection
Terminology: a *str_replace call* is a plain function call (not a method or
static call) that resolves to the global function (names compared
case-insensitively, as PHP does; `\` and a global `use function` import are
fine, but a same-named function declared in the current namespace, one
imported from another namespace, or a qualified non-global name such as
`Ns\str_replace` does not count): `str_replace`. Its arguments are `search`,
`replace`, `subject` (positions 0, 1, 2). *Array literal* = `array(...)` or
`[...]`.

### Entry points
- **D1** A `return` statement whose returned expression, after stripping any
  parentheses, is a str_replace call `F`; or an assignment expression whose
  right side, after stripping parentheses, is a str_replace call `F`
  (statement-level assignments `$v = str_replace(...);` are the relevant
  case; compound assignments: see Divergences).
- **D2** `F` has exactly 3 arguments, or 4 (the by-reference `$count`
  argument). With 3 arguments all checks below (D3–D9) run. With 4 only part
  C (D7/D8) runs: merging would make `$count` include the other call's
  replacements, whereas collapsing identical search items leaves the count
  unchanged (the repeated items never match again). The merged-in call `G`
  or `I` must have exactly 3 arguments (a call carrying `$count` is never
  eliminated).

### Mergeability (shared by D3 and D6)
- **D9** Two calls `A`, `B` *can be merged* unless one of the four arguments
  `A.search`, `A.replace`, `B.search`, `B.replace` that is **not** an array
  literal has an *array-only* inferred type — its known types (unknown
  ignored; `T[]` forms count as `array`) include `array` but not `string` —
  and the configured PHP level is below 7.4. At 7.4+ such arguments are
  allowed (they are spread with `...array_values(…)` by the fix, F3). Untyped/unknown
  variables are not array-only; the `str_replace` stub return type
  (`string|string[]`) is not array-only.

### A. Cascading assignments
- **D3** Find the *previous statement* `S`:
  - for a `return`: the nearest preceding sibling statement;
  - for an assignment: the nearest preceding sibling of the expression
    statement that contains it;
  skipping whitespace and comments, and then skipping any doc comments
  (`/** */`) backwards. `S` must be an expression statement whose expression
  is an assignment `$p = G` where `G` (parentheses stripped) is a
  str_replace call. No other statement may sit between them.
- **D4** The left side of `S` is a plain variable `$p`, `F.subject` is a plain
  variable with the same name, and
  - for an assignment entry: the left side of `F`'s assignment is
    equivalent to `$p` (a plain variable with the same name);
  - for a `return` entry: no further condition.
- **D5** `F` and `G` can be merged (D9) → report `F` (cascade report); fix:
  patch `F`, eliminate `G`.

### B. Nested call
- **D6** `F.subject` is itself (no parentheses) a str_replace call `I`, and
  `I` and `F` can be merged (D9) → report `I` (nested report); fix: patch
  `F`, eliminate `I`. Only the subject of an `F` reached via D1 is checked:
  in `$v = str_replace(a, b, str_replace(c, d, str_replace(e, f, $s)));`
  only the middle call is reported.

### C. Redundant search array
- **D7** `F.replace` is a string literal, `F.search` is an array literal,
  and every element's first child (the value, or the key for `k => v`
  elements) is a string literal; collect their exact source texts (quotes
  included, so `'a'` and `"a"` differ).
- **D8** Exactly one distinct text → report `F.search` (simplification
  report). An empty array or any non-string element → no report.

D3/D5, D6 and D7/D8 are independent; one `F` can yield several reports.

## Exceptions (no report)
- **E1** Previous assignment stores into a different variable than `F`'s
  subject/result (`$b = str_replace(.., $a)` after `$a = …`), or a property
  /array element is involved on either side.
- **E2** Any other statement between the two assignments (comments and doc
  comments do not count).
- **E3** `F` not directly returned/assigned (e.g. passed as an argument):
  none of A, B, C are checked for it.
- **E4** Below PHP 7.4, merging would need spreading an array-only variable
  or expression (D9).
- **E5** Search array with more than one distinct literal, or `replace` not a
  string literal (an array, a variable), for part C.
- **E6** `F` with fewer than 3 or more than 4 arguments; parts A and B when
  `F`, `G` or `I` has a 4th `$count` argument.

## Report
- A (cascade): range = the whole call `F` (name through `)`); severity
  warning. Message: `Fold this str_replace() into the preceding one on the
  same variable.`
- B (nested): range = the whole inner call `I`; severity warning. Message:
  `Fold this nested str_replace() into the enclosing call.`
- C: range = the array literal `F.search`; severity info (weak warning).
  Message: `All searched items are identical; pass the single string.`

## Fix
Notation: patch call `P` (kept), eliminated call `E` (merged in). For a
cascade `P = F`, `E = G`; for nesting `P = F`, `E = I`.

Fixes apply to the code as it stands when applied. When several fixes of a
file are applied top-down, a chain of cascading assignments folds into the
last one (each fix merges with the already-merged previous call) — custos
must reproduce this cumulative result (e.g. by re-running detection after
each fix pass).

- **F1 Replace merge.** Let `unbox(x)` = if `x` is an array literal with
  exactly one element that has no key and whose value is a string literal,
  that string literal; else `x`. Let `const(x)` = if `x` is a global
  constant reference (not `true`/`false`/`null`) resolving to exactly one
  `define()`d value, that value; else `x`.
  - Shortcut: if `const(unbox(P.replace))` and `const(unbox(E.replace))` are
    both string literals and `unbox(E.replace)` is equivalent to
    `unbox(P.replace)` (same node kind, same text — `'.'` vs `"."` differ),
    then `P.replace` becomes `unbox(P.replace)` (e.g. `['-']` → `'-'`) and
    replace lists are **not** merged.
  - Otherwise first *expand* each of `P` and `E`: if its `search` is an
    array literal with ≥ 2 elements and `const(replace)` is not an array
    literal, its `replace` becomes an array of `n` copies of the replace text
    (`n` = number of search elements): `'x'` with 2 searches → `['x', 'x']`.
    Then `P.replace := merge(P.replace, E.replace)` (F3).
- **F2 Search merge.** `P.search := merge(P.search, E.search)` (F3), using
  the original (unexpanded) search arguments.
- **F3 `merge(to, from)`** produces the elements of `from` followed by the
  elements of `to`, where an array literal contributes its elements
  (verbatim text, `k => v` kept) and any other expression contributes itself
  as one element — written `...array_values({expr})` when it has an
  array-only type (D9; `\array_values` when an unqualified call there
  would not reach the global function: a `use function` import under that
  name, or a same-named function declared in the current namespace).
  `str_replace()` ignores keys, so re-indexing does
  not change the result, while spreading the array directly would throw on
  string keys before PHP 8.1 and, from 8.1, let equal string keys of two
  spread arrays overwrite each other and drop entries.
  - If `to` is an array literal, it keeps its own text and the `from`
    elements are inserted right after its opening bracket/parenthesis
    (after any whitespace there), each followed by `, `. If `to` is an
    **empty** array literal nothing is inserted (upstream quirk, see
    Divergences).
  - Otherwise a new array literal is created: `array(` + elements joined
    with `, ` + `)` (converted to the configured syntax in F5).
  - **Explicit keys.** Look at the search and replace array literals of all
    calls taking part in the fix (the whole cascade chain, or `F` and `I`).
    If none has an element with an explicit key, elements are merged
    verbatim as above. Otherwise merged elements are written **without**
    their keys (`'amp' => '&'` contributes `'&'`; a by-reference element
    keeps its `&`), and a `to` array literal no longer keeps its own text:
    the result is a new array literal of `from`'s then `to`'s elements.
    `str_replace()` only observes element order, so dropping keys is exact
    as long as every keyed literal has distinct effective keys; merging
    with keys kept would let an equal key in the other list overwrite an
    entry (`['a' => 'x']` + `['a' => 'y']`, `[1 => 'p']` after two unkeyed
    elements). If a keyed literal has a non-literal key (`$k => 'a'`), a
    key that is not an int/string literal, a negative int key, a spread
    element, or two elements with the same effective key (string keys that
    are canonical decimal integers count as ints; unkeyed elements take the
    next int index), the call is still reported but **no fix** is offered.
- **F4 Subject.** `P.subject` is replaced by the verbatim text of
  `E.subject`. For a cascade, the whole previous statement `S` is then
  deleted together with the whitespace run directly following it; comments
  and doc comments around it stay. For nesting nothing else is deleted (the
  inner call disappears because it was `P.subject`).
- **F5 Rebuild.** `P` is re-emitted as
  `{qualifier}str_replace({a0}, {a1}, {a2})` (qualifier as written, e.g.
  `\`; arguments joined by `, `; original line breaks inside the call are
  lost). Each argument is its current text, except an array literal whose
  syntax differs from the option: it is rewritten as `[e1, e2, …]` (option
  `USE_SHORT_ARRAYS_SYNTAX` = true) or `array(e1, e2, …)` (false), elements'
  verbatim texts joined by `, `. Array literals already in the requested
  syntax keep their text.
- **F6 Simplification fix (part C).** Replace the search array literal with
  the shared string literal text (`array('x', 'x')` → `'x'`).

Worked results (short syntax on):
- `$s = str_replace('a', 'b', $s0); $s = str_replace('c', 'd', $s);` →
  `$s = str_replace(['a', 'c'], ['b', 'd'], $s0);`
- `str_replace('k', '!', str_replace(['j'], ['!'], $in))` →
  `str_replace(['j', 'k'], '!', $in)` (equal replaces, shortcut)
- `str_replace(['p', 'q'], 'z', str_replace(['r'], 'w', $in))` →
  `str_replace(['r', 'p', 'q'], ['w', 'z', 'z'], $in)` (expansion)

## Options
| Option | Type | Default | Effect |
|---|---|---|---|
| USE_SHORT_ARRAYS_SYNTAX | bool | false | Array syntax for arrays in the rebuilt call (F5): `[...]` when true, `array(...)` when false. Detection unaffected. Upstream fixtures run with `true`. |

## PHP versions
- Merging when a non-literal search/replace argument is array-only requires
  PHP ≥ 7.4 (array spread `...array_values($x)`); below that those cases are not
  reported (D9). Other cases: no gating.
- Main upstream fixture runs at the test default level (5.6–7.0); a second
  fixture runs at 7.4 for the spread cases.

## Examples
Short array syntax option on.

```php
<?php
function slug($title, $raw, $line) {
    $title = str_replace('&', 'and', $title);
    /** spaces */
    $title = <warning descr="Fold this str_replace() into the preceding one on the same variable.">str_replace(' ', '-', $title)</warning>;
    return <warning descr="Fold this str_replace() into the preceding one on the same variable.">str_replace(['.', ','], '', $title)</warning>;
}

function others($raw, $line) {
    $code = str_replace('_', '-', <warning descr="Fold this nested str_replace() into the enclosing call.">str_replace(' ', '_', $raw)</warning>);
    $mark = str_replace(['k'], ['!'], <warning descr="Fold this nested str_replace() into the enclosing call.">str_replace('j', '!', $raw)</warning>);
    $clean = str_replace(<weak_warning descr="All searched items are identical; pass the single string.">array("\t", "\t")</weak_warning>, ' ', $line);

    $x = str_replace('a', 'b', $raw);
    $y = str_replace('c', 'd', $x);
    $keep = str_replace(['a', 'b'], 'c', $line);
    log_it(str_replace('q', 'r', str_replace('s', 't', $line)));
    return $y;
}
```

```php
<?php
function slug($title, $raw, $line) {
    /** spaces */
    return str_replace(['&', ' ', '.', ','], ['and', '-', '', ''], $title);
}

function others($raw, $line) {
    $code = str_replace([' ', '_'], ['_', '-'], $raw);
    $mark = str_replace(['j', 'k'], '!', $raw);
    $clean = str_replace("\t", ' ', $line);

    $x = str_replace('a', 'b', $raw);
    $y = str_replace('c', 'd', $x);
    $keep = str_replace(['a', 'b'], 'c', $line);
    log_it(str_replace('q', 'r', str_replace('s', 't', $line)));
    return $y;
}
```

PHP 7.4 (array-typed arguments are spread through `array_values()`; below
7.4 nothing is reported):

```php
<?php
function swap(array $from, array $to, $text) {
    $text = str_replace('<', '&lt;', $text);
    $text = <warning descr="Fold this str_replace() into the preceding one on the same variable.">str_replace($from, $to, $text)</warning>;
    return $text;
}
```

```php
<?php
function swap(array $from, array $to, $text) {
    $text = str_replace(['<', ...array_values($from)], ['&lt;', ...array_values($to)], $text);
    return $text;
}
```

## Divergences
- **Arity of the merged-in call not checked upstream:** `G`/`I` are not
  required to have 3 arguments (upstream would fail on fewer). Recommendation:
  require exactly 3 arguments for `G` and `I` as well.
- **Empty array target:** `merge(to = [], from)` silently drops `from`'s
  elements, changing behaviour. Recommendation: produce the `from` elements
  (treat `[]` as an empty element list). No fixture covers it.
- **Spread of array-only arguments:** custos diverges from upstream (F3).
  Upstream spreads such arguments as plain `...$x` from PHP 7.4. Arrays
  with string keys cannot be spread before 8.1 (the merged call would
  throw), and from 8.1 two spread arrays sharing a string key keep only the
  later entry, so search and replace lists lose items or fall out of step.
  custos emits `...array_values($x)`, which is valid from 7.4 whatever the
  keys and gives the same `str_replace()` result. The affected EA 7.4 case
  is listed in `testdata/ea-divergences.json`.
- **Compound assignments** (`$s .= str_replace(...)`): whether upstream visits
  them as assignments is unverified; recommendation: only plain `=`.
- **Non-statement assignments** (`if ($s = str_replace(…, $s))`): upstream
  looks at the preceding sibling inside the enclosing expression, which is
  never an assignment statement; recommendation: only statement-level
  assignments enter part A (parts B and C still apply).
- **Fourth `$count` argument (custos diverges):** upstream skips every
  check on a call with a 4th argument, including part C. Collapsing
  `['x', 'x']` to `'x'` does not change `$count`, so custos still runs part C
  there (D2); parts A and B stay suppressed because merging would change or
  drop the count.
- **Keyed array literals (custos diverges):** upstream copies merged
  elements with their keys, so two literals sharing a key (`['amp' =>
  '&']` and `['amp' => '@']`, or an explicit int key equal to an implicit
  index of the other list) collapse into one entry and the merged call
  replaces less, with search and replace lists out of step. custos drops
  the keys when merging and offers no fix when the keys cannot be checked
  (F3, "Explicit keys").
- **Function-name matching — custos diverges from upstream.** Upstream matches
  `str_replace` (and `define()` for constant values) by the name as written
  (case-sensitive, any namespace qualifier, no resolution), so a differently
  cased call such as `Str_Replace(...)` is missed while a namespaced or
  imported user function of the same name is reported (and rewritten) as if it
  were the builtin. custos matches case-insensitively and only calls that
  reach the global function.
- **Builtin spelling (custos diverges).** The inserted `array_values(` (a
  custos addition, see above) is written `\array_values(` when a function of
  that name declared in or imported into the namespace would capture a bare
  call (F3).
