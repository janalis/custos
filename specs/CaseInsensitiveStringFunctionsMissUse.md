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
- **D2** The call has **2 or 3** arguments, all positional (no named
  argument, no `...` spread): the needle is then the 2nd written argument.
- **D3** Obtain the needle literal `L` from the 2nd argument (surrounding
  parentheses ignored). The needle must be *proven* to be that literal: when
  any other value could reach the call, the case-sensitive counterpart may
  give another result, so nothing is reported. `L` is:
  - the argument itself when it is a string literal (single or double
    quoted, no interpolation);
  - for a **local variable** inside a function, method or closure: the
    literal of its plain `=` assignment, provided that literal is the only
    value that can reach the call. That requires every reaching assignment
    to be accounted for and to yield the same single literal written
    directly on the right-hand side (not through another variable), the
    entry value not to reach the call (so not a parameter, even one with a
    default, and not a variable imported into a closure or arrow function),
    and no other kind of write to the variable in the function (compound
    assignment, `++`/`--`, destructuring, `foreach` target, `catch`,
    `global`, `static`, `unset`, by-reference closure import). In addition
    the function body must contain nothing that can rebind the variable
    out of sight: no variable-variable, `eval`, `include`/`require`,
    `extract()`, `parse_str()` or `mb_parse_str()`; no reference taken to
    the variable (`$r = &$v`, `[&$v]`), no `yield $v`; and the variable is
    not passed to an unresolved or reference parameter before the read. A
    by-value argument remains stable even when another argument binds by
    reference. Calls later in an enclosing loop may already have run on a
    prior iteration; calls after the read without a back edge do not change
    that read's value;
  - for a **global constant**: its value, when this file declares a
    constant of that name exactly once (`const` or `define()`; a `define()`
    with a non-literal name counts as a declaration of every name);
  - for a **class constant** (`self::X`, `C::X` with `C` declared in this
    file, `static::X` in a final class) declared in that class: its value.
  - Anything else (property, ternary, concatenation, call, …) is not
    traced. `L` is therefore always in the same file as the call.
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
- **E3** Needle not proven to be one string literal (D3): a parameter (with
  or without default), a variable assigned more than one value, written by a
  `foreach`, passed by reference or exposed to `extract()`, a property, a
  top-level variable, a constant declared twice or in another file.
- **E5** Named or unpacked arguments (D2).
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
    foreach (['-', '_'] as $mark) {
        $j = stripos($path, $mark);
    }
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
    foreach (['-', '_'] as $mark) {
        $j = stripos($path, $mark);
    }
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
  `"a⏎b"`) is treated as letter-free and reported. custos scans the whole
  content; no fixture covers it.
- **Interpolated strings:** the raw content of `"{$x}"`/`"$x"` contains the
  variable name's letters, so it is normally skipped; a variable such as `$_`
  has no letters and would be reported although the needle is dynamic.
  custos skips literals containing interpolation.
- **Escape sequences (custos diverges):** upstream tests the raw literal
  text, so letter-free needles written with escapes (`"\t"`, `"\x2F"`) are
  missed because of the escape letter, and `"\x41"` would only be skipped by
  accident. custos decodes the literal first (D4) and tests the actual
  needle; non-UTF-8 byte values are skipped to avoid guessing their
  encoding.
- **Needle tracing (custos diverges):** upstream accepts a non-literal
  needle as soon as its value discovery yields exactly one string literal,
  ignoring every other candidate. A parameter with a letter-free default, a
  variable reassigned from a non-literal, a `foreach` target, a property
  with a default or a variable later filled by reference is then reported
  and renamed although callers or later writes can supply a needle with
  letters. custos requires the literal to be the only possible value (D3)
  and stays silent otherwise; named arguments are skipped as well (D2).
- **Builtin spelling (custos diverges).** Upstream inserts a bare counterpart name in its fix, so a function of that name declared in (or imported into) the namespace captures the rewritten call. custos writes `\name(` in that case (F1).
