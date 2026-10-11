---
id: StringNormalization
group: Control flow
kind: syntax
needs: []
php: { min: "", max: "" }
---

# StringNormalization

## Summary

Two smells when chaining string normalisation calls:

1. Changing the case of a string and *then* trimming / cutting it converts
   characters that are thrown away right after; cut first, convert the case
   of the (shorter) result.
2. Wrapping a case conversion inside another case conversion is pointless
   when the outer call overrides what the inner one did (or repeats it); the
   inner call can be dropped.

## Detection

Definitions (a call "is" one of these functions when it resolves to the
global function of that name: names compared case-insensitively as PHP does —
`TRIM`, `\StrToLower` match — written with a single leading `\` or
unqualified with no same-named function declared or imported in the current
namespace; `Ns\trim(...)` and shadowing user functions never match. Names
are compared in this canonical lower-case form, e.g. for D5a):

- *Length* functions: `trim`, `ltrim`, `rtrim`, `substr`, `mb_substr`.
- *Basic case* functions: `strtolower`, `strtoupper`, `mb_convert_case`,
  `mb_strtolower`, `mb_strtoupper`.
- *Case* functions: the basic case functions plus `ucfirst`, `lcfirst`,
  `ucwords`.

Common preconditions, for an outer call `O`:

- **D1** `O` is a plain function call (not a method / static call) with at
  least one argument.
- **D2** `O`'s first argument is itself, directly (not in parentheses), a
  plain function call `I` (not a method / static call) with at least one
  argument. Let `S` be `I`'s first argument.

Pattern A — case conversion before cutting (checked first):

- **D3** `O` is a length function and `I` is a **basic** case function.
  `ucfirst`/`lcfirst`/`ucwords` do not qualify: they act on the first
  character of the string or of each word, so cutting before or after them
  gives different results (`trim(ucfirst(' x'))` is `'x'`,
  `ucfirst(trim(' x'))` is `'X'`).
- **D3a** When `O` is `substr` or `mb_substr` and `I` is `mb_convert_case`,
  `I`'s second argument must be one of the constants `MB_CASE_UPPER`,
  `MB_CASE_LOWER`, `MB_CASE_FOLD` or their `_SIMPLE` forms (optionally
  written with a leading `\`). Title case depends on what precedes each
  character, exactly like `ucfirst`/`ucwords`
  (`substr(mb_convert_case('hello world', MB_CASE_TITLE), 1)` is
  `'ello World'`, the swapped form `'Ello World'`), and a variable, a number
  or any other expression may be title case: no report. The trim functions
  are not restricted (they only remove characters at the ends, which leaves
  word starts where they were).
- **D3b** When `O` is `substr` or `mb_substr` and `I` is one of the `mb_`
  functions, the report carries no fix and uses the second pattern A
  message: a multibyte conversion can change the number of characters and
  of bytes, and needs whole characters, so offsets taken before and after it
  do not designate the same text. `strtolower`/`strtoupper` convert byte by
  byte and keep every offset.
- **D4** Additionally one of:
  - `O` is `substr` or `mb_substr` (any number of arguments); or
  - `O` is a trim function with exactly one argument; or
  - `O` is a trim function whose second argument is a string literal whose
    complete source text (quotes included) is a single-line single- or
    double-quoted literal that contains **no** Unicode letter
    (e.g. `'-'`, `"/ "`, `'.,;'`, `'0..9'`). A quoted literal containing any
    letter — including letters of escape sequences such as `"\n"` or `"\t"`
    — is not reported. A `b` prefix is ignored. A heredoc/nowdoc literal
    without interpolation is judged by its body (the lines between the
    opening and closing labels, multi-line allowed): reported only when it
    contains no letter. A double-quoted or heredoc literal containing an
    octal escape (a backslash that is not itself escaped, followed by a
    digit `0`–`7`, e.g. `"\101"` for `A`) is not reported: the escape may
    spell a letter. Single-quoted and nowdoc literals have no such escapes.
    A character range `x..y` whose span covers a letter (`'@..Z'`, `'!..~'`)
    counts as containing letters. A non-literal second
    argument (variable, constant, call, concatenation, string with
    interpolation, backtick command) is not reported.
- Report `O` (pattern A).

Pattern B — redundant nested case conversion (only when `O` is not a length
function):

- **D5** `O` and `I` are both case functions, and either:
  - **D5a** `O` and `I` have the same name, the same number of arguments,
    and every argument after the first is equivalent at both levels (same
    mode, encoding or delimiters): `ucwords(ucwords($v))`,
    `ucwords(ucwords($v, '-'), '-')` and
    `mb_convert_case(mb_convert_case($v, M), M)` qualify;
    `ucwords(ucwords($v, '-'))`, `ucwords(ucwords($v), '-')`,
    `ucwords(ucwords($v, '-'), '_')`,
    `mb_convert_case(mb_convert_case($v, MB_CASE_UPPER), MB_CASE_LOWER)` and
    `mb_strtolower(mb_strtolower($v, 'ISO-8859-1'))` do not; or
  - **D5b** the names differ, `O` is a **basic** case function, `I` is one
    of `ucfirst`, `lcfirst`, `ucwords`, and if `I` is `ucwords` it has
    exactly one argument. `ucfirst`/`lcfirst`/`ucwords` nested in a
    different one of those three (`lcfirst(ucwords($v))`) do not qualify:
    the outer call does not undo the inner one.
- Report `I` (pattern B).

A single outer call yields at most one report. Nested chains are visited at
every level, so e.g. `trim(strtolower(strtolower($v)))` yields a pattern A
report on the `trim(...)` call and a pattern B report on the inner
`strtolower(...)`.

- **Name case.** Wherever this rule compares two expressions for
  equivalence, the names PHP resolves case-insensitively — function and
  method names, class names in calls, `new`, `instanceof` and `::`
  accesses, and keywords — compare case-insensitively (`Cache::$map['k']`
  matches `cache::$map['k']`, `$o->Name()` matches `$o->name()`); variable,
  property and constant names stay case-sensitive.

## Exceptions (no report)

- **E1** Method / static calls at either level (`$o->trim(strtolower($v))`,
  `trim($o->strtolower($v))`).
- **E2** The case call is not the *first* argument of the outer call, or is
  wrapped in parentheses.
- **E3** Trim with a character list that contains letters, or that is not a
  plain quoted literal (see D4).
- **E4** A basic case function wrapped in a *different* case function
  (`ucfirst(strtolower($v))`, `lcfirst(mb_strtoupper($v))`,
  `ucwords(mb_convert_case($v, M))`, `strtoupper(mb_strtolower($v))`):
  the inner call is meaningful.
- **E5** `ucwords` with a delimiter argument nested in a different case
  function (`strtoupper(ucwords($v, '_'))`).
- **E6** The correct order already (`strtolower(trim($v))`) — `O` is a case
  function and `I` a length function: no pattern applies.
- **E7** Inner call without arguments.
- **E8** Pattern A with an inner `ucfirst`/`lcfirst`/`ucwords`
  (`trim(ucfirst($v))`, `substr(ucwords($v), 1)`).
- **E9** `ucfirst`/`lcfirst`/`ucwords` nested in a different one of the three
  (`ucfirst(lcfirst($v))`, `lcfirst(ucwords($v))`), and a function nested in
  itself with different extra arguments (D5a).
- **E10** `substr`/`mb_substr` around `mb_convert_case` whose mode is not a
  known upper/lower/fold constant (D3a).

## Report

- Pattern A
  - Range: the whole outer call `O` (name, including any namespace
    qualifier, through its closing `)`).
  - Message: `Cut first, then change the case: '{replacement}'.` with the
    F1 text; for D3b, which has no fix,
    `Cut first, then change the case (check multibyte input by hand).`
- Pattern B
  - Range: the whole inner call `I` (name through its closing `)`); the outer
    call is not highlighted.
  - Message: `The inner '{name}(...)' call has no effect here.` with `I`'s
    name.
- Severity: info (weak warning) for both.

## Fix

- **F1** (pattern A, except D3b which has no fix) Swap the nesting: the
  outer call `O` is replaced with
  `I`'s source text in which `S` is replaced by `O`'s source text in which
  `I` (its first argument) is replaced by `S`. All other text — names as
  written (including `\` qualifiers), remaining arguments, original spacing
  inside each call — is kept verbatim.
  - `ltrim(strtoupper($code))` → `strtoupper(ltrim($code))`
  - `substr(strtolower($code), 0, 3)` → `strtolower(substr($code, 0, 3))`
  - `rtrim(mb_convert_case($code, MB_CASE_TITLE), '-')` →
    `mb_convert_case(rtrim($code, '-'), MB_CASE_TITLE)`
- **F2** (pattern B) Replace the inner call `I` with the source text of `S`
  (other arguments of `I` are dropped): `strtoupper(ucfirst($code))` →
  `strtoupper($code)`.

## Options

None.

## PHP versions

No gating. (Upstream fixture runs at the IDE test default level, below 7.1;
nothing in the rule is version-dependent.)

## Examples

```php
<?php
function tidy($code, $mask, $o) {
    $r = [];
    $r[] = <weak_warning descr="Cut first, then change the case: 'strtoupper(ltrim($code))'.">ltrim(strtoupper($code))</weak_warning>;
    $r[] = <weak_warning descr="Cut first, then change the case (check multibyte input by hand).">substr(mb_strtolower($code), 0, 3)</weak_warning>;
    $r[] = <weak_warning descr="Cut first, then change the case: 'mb_convert_case(rtrim($code, '-'), MB_CASE_TITLE)'.">rtrim(mb_convert_case($code, MB_CASE_TITLE), '-')</weak_warning>;
    $r[] = \trim(lcfirst($code), '#/');
    $r[] = mb_substr(ucwords($code), -2);

    $r[] = trim(strtoupper($code), 'xyz');
    $r[] = trim(strtoupper($code), " \n");
    $r[] = trim(strtoupper($code), $mask);
    $r[] = $o->rtrim(strtoupper($code));
    $r[] = rtrim($o->strtoupper($code));
    $r[] = strtoupper(rtrim($code));

    $r[] = mb_strtoupper(<weak_warning descr="The inner 'mb_strtoupper(...)' call has no effect here.">mb_strtoupper($code)</weak_warning>);
    $r[] = lcfirst(<weak_warning descr="The inner 'lcfirst(...)' call has no effect here.">lcfirst($code)</weak_warning>);
    $r[] = strtoupper(<weak_warning descr="The inner 'ucfirst(...)' call has no effect here.">ucfirst($code)</weak_warning>);
    $r[] = mb_strtolower(<weak_warning descr="The inner 'ucwords(...)' call has no effect here.">ucwords($code)</weak_warning>);
    $r[] = ucfirst(lcfirst($code));

    $r[] = ucwords(strtolower($code));
    $r[] = lcfirst(mb_strtolower($code));
    $r[] = strtolower(mb_strtoupper($code));
    $r[] = strtoupper(ucwords($code, '-'));
    return $r;
}
```

```php
<?php
function tidy($code, $mask, $o) {
    $r = [];
    $r[] = strtoupper(ltrim($code));
    $r[] = substr(mb_strtolower($code), 0, 3);
    $r[] = mb_convert_case(rtrim($code, '-'), MB_CASE_TITLE);
    $r[] = \trim(lcfirst($code), '#/');
    $r[] = mb_substr(ucwords($code), -2);

    $r[] = trim(strtoupper($code), 'xyz');
    $r[] = trim(strtoupper($code), " \n");
    $r[] = trim(strtoupper($code), $mask);
    $r[] = $o->rtrim(strtoupper($code));
    $r[] = rtrim($o->strtoupper($code));
    $r[] = strtoupper(rtrim($code));

    $r[] = mb_strtoupper($code);
    $r[] = lcfirst($code);
    $r[] = strtoupper($code);
    $r[] = mb_strtolower($code);
    $r[] = ucfirst(lcfirst($code));

    $r[] = ucwords(strtolower($code));
    $r[] = lcfirst(mb_strtolower($code));
    $r[] = strtolower(mb_strtoupper($code));
    $r[] = strtoupper(ucwords($code, '-'));
    return $r;
}
```

## Divergences

- Upstream builds the F1 text with global *textual* search-and-replace
  (every occurrence of `I`'s text inside `O`'s text, then every occurrence of
  `S`'s text inside `I`'s text), not by node ranges. It produces broken code
  when `S`'s text also appears elsewhere inside `I` (e.g.
  `trim(mb_convert_case($v, $vMode))` corrupts `$vMode`) or when `I`'s text
  repeats inside `O`. Recommendation: substitute by node ranges as described
  in F1; identical for all upstream fixtures.
- **Pattern A with `ucfirst`/`lcfirst`/`ucwords` — custos diverges from
  upstream** (D3). Upstream swaps these inner calls with trim/substr too,
  but the swap is not equivalent: they change the first character of the
  string or of each word, and cutting moves which character comes first
  (`trim(ucfirst(' x'))` → `'x'`, `ucfirst(trim(' x'))` → `'X'`). custos
  only applies pattern A to the basic case functions.
- **Pattern B between `ucfirst`/`lcfirst`/`ucwords` — custos diverges from
  upstream** (D5b). Upstream drops one of these nested in another
  (`lcfirst(ucwords('a b'))` → `lcfirst('a b')`), changing the result from
  `'a B'` to `'a b'`. custos only drops them under a basic case function,
  which overrides every letter.
- **Same function nested with different extra arguments — custos diverges
  from upstream** (D5a). Upstream drops the inner call whenever the two
  names match. `ucwords(ucwords($v, '-'))` capitalises after `-` and after
  whitespace; dropping the inner call loses the `-` part.
  `mb_convert_case(mb_convert_case('Straße', MB_CASE_UPPER), MB_CASE_LOWER)`
  is `'strasse'`, without the inner call `'straße'`; an inner
  `mb_strtolower($v, 'ISO-8859-1')` under a UTF-8 one is not a repetition
  either. custos requires the extra arguments to be equivalent at both
  levels.
- **Cutting around a multibyte conversion — custos diverges from upstream**
  (D3a, D3b). Upstream swaps `substr`/`mb_substr` with every `mb_` case
  function. With title case the swap capitalises another character (D3a),
  so custos does not report it. With the other modes the swap still differs
  on non-ASCII input: `mb_substr(mb_strtoupper('Straße 12'), 0, 6)` is
  `'STRASS'` but the swapped form `'STRASSE'`, and a byte `substr` applied
  first can cut a character in half. custos keeps the report (the hint is
  valid for ASCII data) but offers no fix.
- **Octal escapes in trim character lists — custos diverges from upstream**
  (D4). `"\101"` contains only digits in source but is the letter `A`:
  `trim(strtolower('AxA'), "\101")` is `'axa'`, the swapped form `'x'`.
  custos does not report such literals.
- **Function names (custos diverges):** upstream matches the bare names
  case-sensitively and ignores namespaces, so `TRIM(strtolower($v))` was
  missed while namespaced user functions with these names were reported and
  their calls swapped or dropped. custos matches case-insensitively and only
  calls resolving to the global functions (Definitions). Messages quote the
  inner call's name as written.
- **Name case (custos diverges).** Upstream compares the expressions
  textually, so operands that differ only in the case of a function, method
  or class name (`Stats::$n` vs `stats::$n`), which PHP treats as the same,
  are not recognised as equivalent. custos folds the case of those names
  (Detection, "Name case").
- **Trim character lists — custos diverges from upstream** (D4). Upstream
  reports every heredoc/nowdoc second argument (its text does not start
  with a quote), checks `b'…'` literals the same way (so `b'abc'` was
  reported), accepts interpolated double-quoted strings by their source
  text, and ignores `x..y` ranges. Each of these lets a letter into the
  trimmed set, and then cutting before or after the case conversion gives
  different results (`trim(strtolower('A'), '@..Z')` keeps `'a'`, the
  swapped form returns `''`). custos reads the heredoc/nowdoc body, strips
  the `b` prefix, skips strings with interpolation and treats a range
  covering a letter as a letter.
