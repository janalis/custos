---
id: CaseInsensitiveStringFunctionsMissUse
group: Performance
kind: semantic
needs: [names, index]
php: { min: "", max: "" }
---

# CaseInsensitiveStringFunctionsMissUse

## Summary
Case-insensitive search functions (`stripos`, `strripos`, `stristr`) do extra
case-folding work. When the needle has no letters at all (e.g. `'/'`, `'::'`,
`'42'`), case does not matter and the plain case-sensitive counterpart gives
the same result faster.

## Detection
- **D1** A function call (not a method/static call) that resolves to the
  global function (names compared case-insensitively, as PHP does; `\` and a
  global `use function` import are fine, but a same-named function declared in
  the current namespace, one imported from another namespace, or a qualified
  non-global name such as `Ns\stripos` does not count) — custos diverges, see
  Divergences — one of:
  - `stripos` → counterpart `strpos`
  - `strripos` → counterpart `strrpos`
  - `stristr` → counterpart `strstr`
- **D2** The call has **2 or 3** arguments.
- **D3** Obtain the needle literal `L` from the 2nd argument:
  - if the argument is a string literal (single or double quoted), `L` is it;
  - otherwise run *value discovery* (as defined in the `CallableMethodValidity`
    spec) on the argument; among the discovered values keep only string
    literals (other values are simply ignored); if exactly one string literal
    remains, it is `L`; otherwise stop.
  - `L` must be in the same file as the call (always true for in-file
    discovery; a constant defined in another file does not qualify).
- **D4** Let `c` be the string value of `L`: the text between the quotes
  with escape sequences decoded according to the quote style (single quotes:
  only `\\` and `\'`; double quotes: `\n`, `\t`, `\x2F`, `\u{2014}`, octal,
  …). Report when `c` is non-empty, is valid UTF-8, and contains no Unicode
  letter (general category `L*`: Latin, Cyrillic, CJK, … letters). Digits,
  punctuation, whitespace, control characters and symbols are not letters.
  A value that is not valid UTF-8 (`"\xE9"`, a lone high byte) is not
  reported: such a byte may be a letter in a single-byte encoding.

Thus `"\t"` (a tab) and `"\x2F"` (a slash) are reported, while `"\x41"`
(`A`) and `'\t'` (backslash + `t`) are not.

## Exceptions (no report)
- **E1** Needle containing any letter, ASCII or not (`'b'`, `'é'`, `'й'`).
- **E2** Empty needle `''`.
- **E3** Needle that cannot be traced to exactly one string literal (non-literal
  argument with no or several literal candidates, top-level variables).
- **E4** Fewer than 2 or more than 3 arguments; other function names.
  A different letter case (`STRIPOS(...)`) is the same function and is
  reported; the fix writes the counterpart in lower case
  (`STRIPOS($p, '/')` → `strpos($p, '/')`).

## Report
- Range: the whole function call, from the start of its name (including any
  namespace qualifier) to the closing `)`.
- Severity: **info (weak warning)**, regardless of the `warning` default listed
  in `rules.json` (upstream forces the weak level for this rule).
- Message: `Needle has no letters; use '{counterpart}(...)' instead.`

## Fix
- **F1** Rename the function: replace only the last name segment with the
  counterpart (`stripos`→`strpos`, `strripos`→`strrpos`, `stristr`→`strstr`).
  Any namespace qualifier, the argument list and all whitespace/comments are
  kept verbatim: `\stristr($p, ':')` → `\strstr($p, ':')`. When the call has
  no qualifier and an unqualified counterpart at that position would not
  reach the global function (a `use function` import under that name, or a same-named function declared in the current namespace), the new segment is written with a
  leading `\` (`stripos($p, '/')` → `\strpos($p, '/')`).

## Options
None.

## PHP versions
No gating. The EA test runs at the PhpStorm default level (between 5.6 and 7.0).

## Examples

```php
<?php
function parts($path, $mode)
{
    $sep = '::';
    $a = <weak_warning descr="Needle has no letters; use 'strpos(...)' instead.">stripos($path, '/')</weak_warning>;
    $b = <weak_warning descr="Needle has no letters; use 'strrpos(...)' instead.">strripos($path, '.', -2)</weak_warning>;
    $c = <weak_warning descr="Needle has no letters; use 'strstr(...)' instead.">\stristr($path, $sep)</weak_warning>;
    $d = <weak_warning descr="Needle has no letters; use 'strpos(...)' instead.">stripos($path, "2024")</weak_warning>;

    $e = stripos($path, 'x');
    $f = strripos($path, 'ß');
    $g = stristr($path, '');
    $h = <weak_warning descr="Needle has no letters; use 'strpos(...)' instead.">stripos($path, "\t")</weak_warning>;
    $i = stripos($path, $mode);
    $j = stripos($path);
    return [$a, $b, $c, $d, $e, $f, $g, $h, $i, $j];
}
```

```php
<?php
function parts($path, $mode)
{
    $sep = '::';
    $a = strpos($path, '/');
    $b = strrpos($path, '.', -2);
    $c = \strstr($path, $sep);
    $d = strpos($path, "2024");

    $e = stripos($path, 'x');
    $f = strripos($path, 'ß');
    $g = stristr($path, '');
    $h = strpos($path, "\t");
    $i = stripos($path, $mode);
    $j = stripos($path);
    return [$a, $b, $c, $d, $e, $f, $g, $h, $i, $j];
}
```

## Divergences
- **Function-name case (custos diverges):** upstream matches the names
  case-sensitively, so `STRIPOS($p, '/')` or `\StrIPos(...)` is missed
  although PHP calls the same function; any namespace qualifier is accepted
  without resolution, so a namespaced or imported user `stripos()` is
  reported (and renamed). custos compares the name case-insensitively and
  requires the call to reach the global function (D1).
- **Line breaks in the needle (upstream bug):** upstream's letter test does
  not look past a line terminator, so a literal whose raw content contains a
  real newline plus letters (e.g. a double-quoted string spanning two lines,
  `"a⏎b"`) is treated as letter-free and reported. Recommendation: scan the
  whole content; no fixture covers it.
- **Interpolated strings:** the raw content of `"{$x}"`/`"$x"` contains the
  variable name's letters, so it is normally skipped; a variable such as `$_`
  has no letters and would be reported although the needle is dynamic.
  Recommendation: skip literals containing interpolation.
- **Escape sequences (custos diverges):** upstream tests the raw literal
  text, so letter-free needles written with escapes (`"\t"`, `"\x2F"`) are
  missed because of the escape letter, and `"\x41"` would only be skipped by
  accident. custos decodes the literal first (D4) and tests the actual
  needle; non-UTF-8 byte values are skipped to avoid guessing their
  encoding.
- **Builtin spelling (custos diverges).** Upstream inserts a bare counterpart name in its fix, so a function of that name declared in (or imported into) the namespace captures the rewritten call. custos writes `\name(` in that case (F1).
