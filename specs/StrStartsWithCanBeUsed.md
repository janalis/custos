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
  `strpos` or `mb_strpos` and which resolves to that global
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
  `, `.
  - `0 === strpos($path, '/')` → `str_starts_with($path, '/')`
  - `mb_strpos($s, $p) !== 0` → `!str_starts_with($s, $p)`

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
        <weak_warning descr="Replace with 'str_starts_with($uri, '/api')'.">0 === mb_strpos($uri, '/api')</weak_warning>,
        <weak_warning descr="Replace with '!\str_starts_with(strtolower($uri), $prefix)'.">\strpos(strtolower($uri), $prefix) !== 0</weak_warning>,
        strpos($uri, $prefix, 1) === 0,
        strpos($uri, $prefix) == 0,
        strpos($uri, $prefix) === 00,
        strpos($uri, $prefix) === false,
    ];
}
```

```php
<?php
function route(string $uri, string $prefix): array {
    return [
        str_starts_with($uri, $prefix),
        str_starts_with($uri, '/api'),
        !\str_starts_with(strtolower($uri), $prefix),
        strpos($uri, $prefix, 1) === 0,
        strpos($uri, $prefix) == 0,
        strpos($uri, $prefix) === 00,
        strpos($uri, $prefix) === false,
    ];
}
```

## Divergences
- **Callee resolved — custos diverges from upstream** (D1, D5). Upstream
  ignores namespaces: a user `App\strpos()` is reported, and the generated
  call keeps the original qualifier (`Foo\strpos` → `Foo\str_starts_with`,
  usually an undefined function). custos reports only calls that reach the
  built-in, and qualifies the generated `str_starts_with` with `\` when a
  bare name would not reach the built-in.
- **Function-name case (custos diverges):** upstream compares `strpos` /
  `mb_strpos` case-sensitively, missing `StrPos(...)` or `MB_STRPOS(...)`,
  which call the same built-ins. custos matches any case (D1).
