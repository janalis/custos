---
id: StrContainsCanBeUsed
group: Language level migration
kind: semantic
needs: [names, stubs]
php: { min: "8.0", max: "" }
---

# StrContainsCanBeUsed

## Summary

Comparing a `strpos()`/`mb_strpos()` result strictly against `false` is the
pre-PHP 8 idiom for "does the string contain this substring". PHP 8 has
`str_contains()` for exactly that.

## Detection

- **D1** A function call `P` whose name (last segment, case-insensitive, as
  PHP compares function names: `StrPos` matches) is
  `strpos` or `mb_strpos` and which resolves to that global
  function under PHP's runtime rules: `\strpos(...)` and an unqualified call in
  the global namespace match; an unqualified call inside a namespace matches
  only when no function of that name is declared in the namespace or
  imported from elsewhere; a qualified `Foo\strpos(...)` does not match.
- **D2** `P` has exactly two arguments, haystack `H` and needle `N`.
- **D3** `P`'s **direct** parent is a binary `===` or `!==` expression `B`
  (a parenthesised call has a parenthesised-expression parent → no match).
- **D4** The other operand of `B` (whichever side `P` is not on) is the
  constant `false`, case-insensitive (`false`, `FALSE`).
- **D5** Replacement text:
  - `ns` = `\` when the original call is written fully qualified or when an
    unqualified call to the generated function would not reach the global
    function at that position (a same-named function declared in, or
    imported into, the current namespace); empty otherwise.
  - core = `ns` + `str_contains(` + H text + `, ` + N text + `)`.
  - `!==` → `R = core`; `===` → `R = !core` (negated).

## Exceptions (no report)

- **E1** Language level below 8.0.
- **E2** A third argument (offset), or a fourth (`mb_strpos` encoding), or
  fewer than two arguments.
- **E3** Loose comparison (`==`, `!=`), comparison with anything other than
  `false` (`0`, `null`, `-1`), or a comparison where the call is wrapped in
  parentheses.
- **E4** Other functions (`stripos`, `strrpos`, `mb_stripos`, `strstr`, …).

## Report

- Range: the whole comparison `B` (from its left operand start to its right
  operand end).
- Severity: info (weak warning).
- Message: `Replace with '{R}'.` (`{R}` exactly as produced by the fix, e.g.
  `!str_contains($a, $b)` for the negated form).

## Fix

- **F1** Replace `B` with `R`. The negated form is emitted **without** a space
  after `!`: `!str_contains(H, N)`. Arguments are separated by `, ` and copied
  verbatim.
  - `strpos($s, 'x') !== false` → `str_contains($s, 'x')`
  - `false === \mb_strpos($s, $t)` → `!\str_contains($s, $t)`

## Options

None.

## PHP versions

Reported only at language level ≥ 8.0. The upstream fixture runs at 8.0.

## Examples

```php
<?php
function scan(string $line, string $tag) {
    $hit  = <weak_warning descr="Replace with 'str_contains($line, $tag)'.">strpos($line, $tag) !== false</weak_warning>;
    $hit2 = <weak_warning descr="Replace with 'str_contains($line, '#')'.">FALSE !== mb_strpos($line, '#')</weak_warning>;
    $miss = <weak_warning descr="Replace with '!\str_contains(trim($line), $tag)'.">\strpos(trim($line), $tag) === false</weak_warning>;
    $from = strpos($line, $tag, 3) !== false;
    $loose = strpos($line, $tag) != false;
    $zero = strpos($line, $tag) === 0;
    $wrapped = (strpos($line, $tag)) !== false;
    return [$hit, $hit2, $miss, $from, $loose, $zero, $wrapped];
}
```

```php
<?php
function scan(string $line, string $tag) {
    $hit  = str_contains($line, $tag);
    $hit2 = str_contains($line, '#');
    $miss = !\str_contains(trim($line), $tag);
    $from = strpos($line, $tag, 3) !== false;
    $loose = strpos($line, $tag) != false;
    $zero = strpos($line, $tag) === 0;
    $wrapped = (strpos($line, $tag)) !== false;
    return [$hit, $hit2, $miss, $from, $loose, $zero, $wrapped];
}
```

## Divergences

- **Callee resolved — custos diverges from upstream** (D1, D5). Upstream
  ignores namespaces: a user `App\strpos()` (with its own semantics) is
  reported, and the generated call keeps the original qualifier, producing a
  call to a function that may not exist (`Foo\strpos` → `Foo\str_contains`).
  custos reports only calls that reach the built-in, and qualifies the
  generated `str_contains` with `\` when a bare name would not reach the
  built-in.
- Replacing `mb_strpos` by `str_contains` is byte-based; equivalent for
  containment checks with valid encodings. Kept.
- **Function-name case (custos diverges):** upstream compares `strpos` /
  `mb_strpos` case-sensitively, missing `StrPos(...)` or `MB_STRPOS(...)`,
  which call the same built-ins. custos matches any case (D1).
