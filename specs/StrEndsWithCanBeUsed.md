---
id: StrEndsWithCanBeUsed
group: Language level migration
kind: semantic
needs: [names, stubs]
php: { min: "8.0", max: "" }
---

# StrEndsWithCanBeUsed

## Summary
Cutting the tail off a string with `substr($h, -strlen($n))` and comparing it
to `$n` is the pre-PHP 8 "ends with" idiom. PHP 8 has `str_ends_with()`,
which is clearer and avoids building a temporary string.

## Detection
- **D1** A function call `S` whose name (last segment, case-insensitive, as
  PHP compares function names: `SubStr` matches) is
  `substr` or `mb_substr` and which resolves to that global
  function under PHP's runtime rules: `\substr(...)` and an unqualified call in
  the global namespace match; an unqualified call inside a namespace matches
  only when no function of that name is declared in the namespace or
  imported from elsewhere; a qualified `Foo\substr(...)` does not match.
- **D2** `S` has exactly two arguments, haystack `H` and start `X`.
- **D3** `S`'s **direct** parent is a binary `===` or `!==` expression `B`.
- **D4** `X` is a unary minus whose operand is **directly** (no parentheses) a
  function call named (last segment, case-insensitive) `strlen` or `mb_strlen`
  that resolves to that global function (same rules as D1), with exactly one
  argument `L`. Whitespace between `-` and the call is
  allowed (`- strlen($n)`). The two functions need not "match"
  (`substr` + `mb_strlen` is accepted).
- **D5** The other operand of `B` (call it `N`) is structurally equivalent to
  `L` (same node kind and token sequence ignoring whitespace/comments, or
  identical text; plain variables compare by name).
- **D6** Replacement text:
  - `ns` = `\` when the original call is written fully qualified or when an
    unqualified call to the generated function would not reach the global
    function at that position (a same-named function declared in, or
    imported into, the current namespace); empty otherwise.
  - core = `ns` + `str_ends_with(` + H text + `, ` + N text + `)` — the needle
    text is taken from the comparison operand `N`, not from `L`.
  - `===` → `R = core`; `!==` → `R = !core` (negated).
- **D7** `N` (ignoring parentheses) is not an empty string literal (`''`
  or `""`).
- **Name case.** Wherever this rule compares two expressions for
  equivalence, the names PHP resolves case-insensitively — function and
  method names, class names in calls, `new`, `instanceof` and `::`
  accesses, and keywords — compare case-insensitively (`Cache::$map['k']`
  matches `cache::$map['k']`, `$o->Name()` matches `$o->name()`); variable,
  property and constant names stay case-sensitive.

## Exceptions (no report)
- **E1** Language level below 8.0.
- **E2** A length argument (three or more arguments).
- **E3** Start argument not of the form `-strlen(x)` / `-mb_strlen(x)`
  (a literal offset, `-(strlen($n))`, `strlen($n) * -1`, `mb_strlen` with an
  encoding argument).
- **E4** The compared value differs from the measured needle.
- **E5** Loose comparison or the call wrapped in parentheses.
- **E6** Empty literal needle: `substr($h, -strlen('')) === ''`.

## Report
- Range: the whole comparison `B`.
- Severity: info (weak warning).
- Message: `Replace with '{R}'.` (`{R}` exactly as produced by the fix).

## Fix
- **F1** Replace `B` with `R`; the negated form has no space after `!`
  (`!str_ends_with(H, N)`). Arguments are copied verbatim, separated by `, `.
  - `substr($file, -strlen('.php')) === '.php'` → `str_ends_with($file, '.php')`
  - `$ext !== mb_substr($name, - mb_strlen($ext))` → `!str_ends_with($name, $ext)`

## Options
None.

## PHP versions
Reported only at language level ≥ 8.0. The upstream fixture runs at 8.0.

## Examples

```php
<?php
function isType(string $file, string $ext): array {
    return [
        <weak_warning descr="Replace with 'str_ends_with($file, $ext)'.">substr($file, -strlen($ext)) === $ext</weak_warning>,
        <weak_warning descr="Replace with 'str_ends_with($file, '.gz')'.">'.gz' === mb_substr($file, - mb_strlen('.gz'))</weak_warning>,
        <weak_warning descr="Replace with '!str_ends_with(basename($file), $ext)'.">substr(basename($file), -mb_strlen($ext)) !== $ext</weak_warning>,
        substr($file, -strlen($ext)) === '.txt',
        substr($file, -strlen($ext), 2) === $ext,
        substr($file, -3) === $ext,
        mb_substr($file, -mb_strlen($ext, 'UTF-8')) === $ext,
    ];
}
```

```php
<?php
function isType(string $file, string $ext): array {
    return [
        str_ends_with($file, $ext),
        str_ends_with($file, '.gz'),
        !str_ends_with(basename($file), $ext),
        substr($file, -strlen($ext)) === '.txt',
        substr($file, -strlen($ext), 2) === $ext,
        substr($file, -3) === $ext,
        mb_substr($file, -mb_strlen($ext, 'UTF-8')) === $ext,
    ];
}
```

## Divergences
- **Empty needle — custos diverges from upstream** (D7). With `''` as the
  needle, `-strlen('')` is `0`, so `substr($h, 0)` is the whole haystack and
  the comparison is false for any non-empty `$h`; `str_ends_with($h, '')` is
  always true. Upstream rewrites it anyway; custos skips empty literal
  needles. Needles held in variables are still reported (their emptiness is
  unknown).
- **Callees resolved — custos diverges from upstream** (D1, D4, D6).
  Upstream ignores namespaces for `substr`/`mb_substr` and for
  `strlen`/`mb_strlen`, so user functions with those names (e.g.
  `App\strlen()` counting something else) are reported and rewritten, and
  the generated call keeps a foreign qualifier. custos requires every matched
  call to reach the built-in and qualifies the generated `str_ends_with` with
  `\` when a bare name would not reach the built-in.
- **Function-name case (custos diverges):** upstream compares `substr`,
  `mb_substr`, `strlen` and `mb_strlen` case-sensitively, missing
  `SubStr($h, -StrLen($n)) === $n`. custos matches any case (D1, D4).
- **Name case (custos diverges).** Upstream compares the expressions
  textually, so operands that differ only in the case of a function, method
  or class name (`Stats::$n` vs `stats::$n`), which PHP treats as the same,
  are not recognised as equivalent. custos folds the case of those names
  (Detection, "Name case").
- **Possibly empty needles (custos diverges).** Beyond the empty literal of
  D7, a needle that may be `''` makes the original comparison false and
  `str_ends_with()` true. The fix is only offered when the needle is known
  to be non-empty: a non-empty literal, a concatenation with such a part,
  or a variable or constant whose every possible value is one; other
  needles (`string $ext` parameters) are reported without a fix.
