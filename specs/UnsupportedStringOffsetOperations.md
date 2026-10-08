---
id: UnsupportedStringOffsetOperations
group: Language level migration
kind: semantic
needs: [types]
php: { min: "7.1", max: "" }
---

# UnsupportedStringOffsetOperations

## Summary

Writing through a string offset as if it were a nested array
(`$str[0][1] = …`) or appending to a string with `[]` (`$str[] = …`) is a
fatal error since PHP 7.1. When the written-to value is known to be a string,
such writes are bugs.

## Detection

Visit every array-access expression `E` (`base[index]` or `base[]`); let `C`
be its base.

- **D1** `C` is a variable, a property access, or an array access.
- **D2** Superglobal guard: descend from `C` through nested array-access
  bases until reaching a non-array-access expression; if that is a variable
  named `_GET`, `_POST`, `_SESSION`, `_REQUEST`, `_FILES`, `_COOKIE`, `_ENV`,
  `_SERVER`, `GLOBALS` or `HTTP_RAW_POST_DATA`, stop (no report for `E`).
- **D3 Nested offset write** — when `E`'s direct parent is itself an array
  access:
  - target `T` = the outermost array access of the chain containing `E`
    (climb while the parent is an array access);
  - context = `T`'s parent; if that is an element of a destructuring pattern
    (`list(...)` / `[...]` item, keyed or not), use the pattern's parent
    instead;
  - the context must be an assignment (plain `=`, compound `.=`, `+=` …, or a
    destructuring assignment `list(...) = …` / `[...] = …`) in which `T` is
    **not** the assigned value (i.e. `T` is on the writing side).
- **D4 Append to string** — otherwise, when `E` has an empty index (`[]`) and
  `E`'s direct parent is an assignment in which `E` is not the assigned value:
  target `T` = `E`.
- **D5** `T` must be inside a function, method or closure body (code at file
  top level is not examined).
- **D6** The inferred type of `C` must consist of exactly one type (unknowns
  count; `string|null` is two), and that type is `string`.
- **D7** Report `T` with the D3 or D4 meaning.

Type inference needed for D6: declared parameter types and `@param` types,
property types, `@var` annotations, assignments. Element type of an
array access whose base is typed `X[]` is `X` (so `$names[0]` with
`string[] $names` is a string); an array access on a `string`-typed base is
**not** inferred as `string` (its type is unknown here). Thus for
`$s[0][] = 'x'` with `string $s` only the D3 report on the whole chain
appears, not an additional D4 report.

## Exceptions (no report)

- **E1** Language level below 7.1.
- **E2** Reads: `$v = $s[0][0];`, `$v = [$s[0][0]];`, `foo($s[0][0])`.
- **E3** Single-level writes `$s[0] = 'x';` (valid on strings).
- **E4** Superglobal roots (`$_POST['a'][] = 1`, `$GLOBALS['x']['y'] = 2`).
- **E5** Top-level (global scope) code.
- **E6** Base type unknown, mixed, union, or not `string`.
- **E7** `list($s[]) = …`: the append form requires the assignment to be the
  direct parent, so an append inside a destructuring pattern is not reported.

## Report

- Range: `T` — the whole outermost array-access chain for D3 (e.g.
  `$s[0]['k']`, `$s[0][]`), or the `$s[]` expression for D4.
- Severity: error.
- Messages:
  - D3: `String offsets cannot be written as nested arrays (fatal error).`
  - D4: `Appending with [] is not supported on strings (fatal error).`

## Fix

None.

## Options

None.

## PHP versions

Reported only at language level ≥ 7.1. The upstream fixture runs at 7.1.

## Examples

```php
<?php
/** @param string[] $words */
function mangle(string $text, array $words, $other) {
    $first = $text[0][0];
    $copy  = [$text[1][0]];
    $text[2] = 'z';

    <error descr="String offsets cannot be written as nested arrays (fatal error).">$text[3][1]</error> = 'y';
    <error descr="String offsets cannot be written as nested arrays (fatal error).">$text['a']['b']</error> .= 'q';
    [<error descr="String offsets cannot be written as nested arrays (fatal error).">$text[4]['c']</error>, $tail] = $words;
    <error descr="Appending with [] is not supported on strings (fatal error).">$text[]</error> = 'w';
    <error descr="String offsets cannot be written as nested arrays (fatal error).">$text[5][]</error> = 'v';
    <error descr="String offsets cannot be written as nested arrays (fatal error).">$words[0][1][2]</error> = 'u';

    <error descr="Appending with [] is not supported on strings (fatal error).">$words[0][]</error> = 't';
    $words[] = 'ok';
    $other[0][0] = 1;
    $_SESSION['cart'][] = 'item';
    $_COOKIE['a']['b'] = 'c';
}

$str = 'top';
$str[0][0] = 'x';
```

## Divergences

- Compound assignment counts as a write context; this matches how PHP treats
  it (still fatal). No upstream fixture; behaviour derived from upstream's
  assignment check, which covers compound assignments as well.
- The `$words[0][1][2]` and `$words[0][]` cases rely on the element-type
  rule (`string[]` → `string`); the upstream fixture does not cover them.
- **Offsets of local variables (custos diverges).** For a nested target
  (`C` is itself an offset access), D6 trusts the inferred element type only
  when the chain's root has a declared type: a property (typed or
  `@var`), or a parameter of the enclosing function. A local variable
  initialised with an array literal (`$page = ['#markup' => 'x'];`) gets
  the literal's element type for *every* key, so `$page['#attached'][] =
  ...` — an absent key that PHP autovivifies as an array — was reported as
  a fatal error. Locals are no longer trusted there; the rare real fatal
  on a present string key of a local literal is missed.
