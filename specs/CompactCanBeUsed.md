---
id: CompactCanBeUsed
group: Control flow
kind: syntax
needs: []
php: { min: "", max: "" }
---

# CompactCanBeUsed

## Summary

An array literal whose every entry maps a string key to the variable of the
same name (`['id' => $id, 'tag' => $tag]`) can be written as
`compact('id', 'tag')`, which avoids repeating each name twice.

## Detection

- **D1** An array literal, either `[...]` or `array(...)`.
- **D2** It is **not** the target of an assignment: if its direct parent is an
  assignment expression (plain `=` or compound) and the array is not that
  assignment's value (i.e. it is the left-hand side, destructuring), skip it.
- **D3** Every element of the array (in source order) is a key/value pair
  where:
  - the key is a string literal (single- or double-quoted), and
  - the value is directly a simple variable (no parentheses, no cast, no
    reference/spread wrapper, no property/array access), and
  - the key's string value (text between the quotes with escape sequences
    decoded per quote style, case-sensitive) is exactly equal to the
    variable's name without `$`: `"t\x61g" => $tag` matches, while
    `'t\x61g' => $tag` does not (single quotes keep the backslash). The fix
    still copies the key literals verbatim.
  If any element fails any of these (value-only element, numeric/constant
  key, mismatching name, non-variable value, spread `...$x`), the whole array
  is not reported.
- **D4** At least two elements satisfy D3 (a single pair is not reported; an
  empty array is not reported).

## Exceptions (no report)

- **E1** Destructuring assignment targets: `['a' => $a, 'b' => $b] = $src;`.
- **E2** One-element arrays.
- **E3** Any key/name mismatch (`'a' => $b`), or any value that is an
  expression rather than a bare variable (`'a' => (int) $a`, `'a' => $a[0]`).
- **E4** Non-string keys (`0 => $a`, `KEY => $a`).

## Report

- Range: the first token of the array literal only — the `[` of a short
  array, or the `array` keyword of a long array (not the whole literal).
- Severity: info (weak warning). Rule disabled by default.
- Message (our wording): `Replace with '{replacement}'.` where
  `{replacement}` is the F1 text.

## Fix

- **F1** Replace the whole array literal (from `[`/`array` to the closing
  `]`/`)`) with `compact(K1, K2, ...)`, where each `Ki` is the **source text
  of the key literal including its quotes** (so `"a"` stays double-quoted),
  in element order, joined with `", "`. Examples:
  `['id' => $id, 'tag' => $tag]` → `compact('id', 'tag')`;
  `array("n" => $n, 'm' => $m)` → `compact("n", 'm')`.
  Comments inside the literal are dropped. `compact` is written `\compact`
  when an unqualified call at that position would not reach the global
  function (a `use function` import under that name, or a same-named function declared in the current namespace).

## Options

None.

## PHP versions

No gating (`compact()` and both array syntaxes exist in all supported
versions; short syntax requires 5.4 only for parsing).

## Examples

```php
<?php
function payload($id, $tag, $note) {
    $a = <weak_warning descr="Replace with 'compact('id', 'tag')'.">[</weak_warning>'id' => $id, 'tag' => $tag];
    $b = <weak_warning descr="Replace with 'compact(&quot;id&quot;, 'note', 'tag')'.">array</weak_warning>("id" => $id, 'note' => $note, 'tag' => $tag,);
    send(<weak_warning descr="Replace with 'compact('note', 'id')'.">[</weak_warning>'note' => $note, 'id' => $id]);

    $c = ['id' => $id];
    $d = ['id' => $tag, 'tag' => $id];
    $e = ['id' => $id, 'tag' => trim($tag)];
    $f = ['id' => $id, 'tag' => $tag, $note];
    $g = ['ID' => $id, 'tag' => $tag];
    ['id' => $id, 'tag' => $tag] = load();
    return [$a, $b, $c, $d, $e, $f, $g];
}
```

```php
<?php
function payload($id, $tag, $note) {
    $a = compact('id', 'tag');
    $b = compact("id", 'note', 'tag');
    send(compact('note', 'id'));

    $c = ['id' => $id];
    $d = ['id' => $tag, 'tag' => $id];
    $e = ['id' => $id, 'tag' => trim($tag)];
    $f = ['id' => $id, 'tag' => $tag, $note];
    $g = ['ID' => $id, 'tag' => $tag];
    ['id' => $id, 'tag' => $tag] = load();
    return [$a, $b, $c, $d, $e, $f, $g];
}
```

## Divergences

- **Other destructuring contexts (upstream false positive).** Upstream only
  excludes arrays that are the direct left side of an assignment. Keyed
  destructuring in `foreach ($rows as ['id' => $id, 'tag' => $tag])`, nested
  destructuring (`[['a' => $a, 'b' => $b]] = $src;` — inner array) and
  `list('a' => $a, ...)`-style targets nested in other arrays would be
  reported and "fixed" into invalid code. Recommendation: skip every array
  literal in a write/destructuring position. No upstream fixture covers
  these, so conformance is unaffected.
- **By-reference values.** `['a' => &$a, 'b' => &$b]` — upstream behaviour
  depends on how the parser exposes the `&`; `compact()` copies values, so
  the rewrite would change semantics. Recommendation: do not report when any
  value is by-reference.
- **`$this`.** `['this' => $this, 'x' => $x]` matches D3 upstream.
  Recommendation: keep (compact('this') works), no fixture covers it.
- **Escapes in keys (custos diverges).** Upstream compares the key text as
  written, so `"t\x61g" => $tag` is missed although the key is `tag`. custos
  decodes the literal before comparing (D3); `compact()` receives the same
  verbatim literal, so the rewrite stays equivalent.
- **Builtin spelling (custos diverges).** Upstream always inserts a bare
  `compact(`, which a namespaced or imported function of that name captures
  (it would not even see the local variables). custos writes `\compact(`
  in that case (F1).
- **Arrow functions (custos diverges).** An arrow function captures only
  the outer variables its body names; `fn () => compact('a')` names none,
  so `$a` is undefined there. An array inside an arrow function is
  reported only when every variable is a parameter of that arrow function.
