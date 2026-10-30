---
id: StrStartsWithCanBeUsed
group: Language level migration
kind: semantic
needs: [names, stubs]
php: { min: "8.0", max: "" }
---

# StrStartsWithCanBeUsed

## Summary

`strpos($h, $n) === 0` is the pre-PHP 8 way to ask "does `$h` begin with
`$n`". PHP 8 provides `str_starts_with()`, which says it directly and does not
scan the whole haystack.

## Detection

- **D1** A function call `P` whose name (last segment, case-insensitive, as
  PHP compares function names: `StrPos` matches) is
  `strpos` and which resolves to that global
  function under PHP's runtime rules: `\strpos(...)` and an unqualified call in
  the global namespace match; an unqualified call inside a namespace matches
  only when no function of that name is declared in the namespace or
  imported from elsewhere; a qualified `Foo\strpos(...)` does not match.
- **D2** `P` has exactly two arguments, haystack `H` and needle `N`.
- **D3** `P`'s **direct** parent is a binary `===` or `!==` expression `B`
  (a parenthesised call does not qualify).
- **D4** The other operand of `B` is an integer literal whose source text is
  exactly `0` (not `00`, `0x0`, `0.0`, `-0`, `'0'`, or a constant).
- **D5** Replacement text:
  - `ns` = `\` when the original call is written fully qualified or when an
    unqualified call to the generated function would not reach the global
    function at that position (a same-named function declared in, or
    imported into, the current namespace); empty otherwise.
  - core = `ns` + `str_starts_with(` + H text + `, ` + N text + `)`.
  - `===` → `R = core`; `!==` → `R = !core` (negated).

### Safety requirements

Only the byte-based built-in `strpos` qualifies. `mb_strpos` is skipped
because its mutable default encoding does not establish byte equivalence.

## Exceptions (no report)

- **E1** Language level below 8.0.
- **E2** Offset or encoding argument present, or fewer than two arguments.
- **E3** Loose comparisons (`==`, `!=`), comparisons with anything other than
  the literal `0`, or the call wrapped in parentheses.
- **E4** Other functions (`stripos`, `strrpos`, `mb_stripos`, …).

## Report

- Range: the whole comparison `B`.
- Severity: info (weak warning).
- Message: `Replace with '{R}'.` (`{R}` exactly as produced by the fix).

## Fix

- **F1** Replace `B` with `R`; the negated form has no space after `!`
  (`!str_starts_with(H, N)`). Arguments are copied verbatim, separated by
  `, `. (custos) When the two arguments were separated by a comma and a line
  break only, the fix keeps that separator; the message shows `, `.
  - `0 === strpos($path, '/')` → `str_starts_with($path, '/')`
  - `strpos($s, $p) !== 0` → `!str_starts_with($s, $p)`
- (custos) No fix when a comment lies in the replaced comparison outside
  the copied arguments: the rebuilt call would drop it.

## Options

None.

## PHP versions

Reported only at language level ≥ 8.0. The upstream fixture runs at 8.0.

## Examples

```php
<?php
function route(string $uri, string $prefix): array {
    return [
        <weak_warning descr="Replace with 'str_starts_with($uri, $prefix)'.">strpos($uri, $prefix) === 0</weak_warning>,
        0 === mb_strpos($uri, '/api'),
        <weak_warning descr="Replace with '!\str_starts_with(strtolower($uri), $prefix)'.">\strpos(strtolower($uri), $prefix) !== 0</weak_warning>,
    ];
}
```

```php
<?php
function route(string $uri, string $prefix): array {
    return [
        str_starts_with($uri, $prefix),
        0 === mb_strpos($uri, '/api'),
        !\str_starts_with(strtolower($uri), $prefix),
    ];
}
```

## Divergences

- **Audit safety (custos diverges).** Multibyte search is not rewritten to byte search. The default encoding
  can make mb_strpos disagree with str_starts_with.

- **Callee resolved — custos diverges from upstream** (D1, D5). Upstream
  ignores namespaces: a user `App\strpos()` is reported, and the generated
  call keeps the original qualifier (`Foo\strpos` → `Foo\str_starts_with`,
  usually an undefined function). custos reports only calls that reach the
  built-in, and qualifies the generated `str_starts_with` with `\` when a
  bare name would not reach the built-in.
- **Function-name case (custos diverges):** upstream compares `strpos` /
  `mb_strpos` case-sensitively, missing `StrPos(...)` or `MB_STRPOS(...)`,
  which call the same built-ins. custos matches any case (D1).

Recorded in `internal/testing/conformance/testdata/ea-divergences.json`.
