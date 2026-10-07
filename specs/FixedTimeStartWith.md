---
id: FixedTimeStartWith
group: Performance
kind: syntax
needs: []
php: { min: "", max: "" }
---

# FixedTimeStartWith

## Summary
`strpos($s, 'lit') === 0` scans the whole haystack when the prefix is absent,
so its cost grows with the length of `$s`. Comparing only the first N bytes
with `strncmp`/`strncasecmp` answers the same "starts with" question in time
bounded by the needle length.

## Detection
- **D1** A function call (not a method/static call) that resolves to the
  global function (names compared case-insensitively, as PHP does; `\` and a
  global `use function` import are fine, but a same-named function declared in
  the current namespace, one imported from another namespace, or a qualified
  non-global name such as `Ns\strpos` does not count): `strpos` (counterpart
  `strncmp`) or `stripos` (counterpart `strncasecmp`).
- **D2** The call's **direct** parent (no parentheses in between) is a binary
  expression `B` whose operator is `===` or `!==`. (`==`, `!=`, `<>`, `<` …
  do not qualify.) The call may be the left or the right operand.
- **D3** The call has **exactly 2** arguments.
- **D4** The 2nd argument is a string literal (single or double quoted)
  without any interpolated part (`"$x"`, `"{$x}"` disqualify). A variable,
  constant or concatenation does not qualify.
- **D4a** The needle's decoded value is not empty (`''`, `""` are not
  reported).
- **D5** The other operand of `B` has source text exactly `0` (not `00`,
  `0x0`, `0.0`, `-0`, `(0)`, `false`).

## Exceptions (no report)
- **E1** A 3rd (offset) argument is present.
- **E2** Needle is not a plain literal or contains interpolation.
- **E5** Empty needle (D4a).
- **E3** Loose comparison (`==`/`!=`), or comparison against anything other
  than the literal text `0`.
- **E4** The call is parenthesised or nested in another expression before the
  comparison.

## Report
- Range: the function call only (name including any qualifier through the
  closing `)`), not the comparison.
- Severity: warning (rule disabled by default).
- Message: `Use '{replacement}' for a length-independent prefix check.`
  where `{replacement}` is the F1 text.

## Fix
- **F1** Replace the function call (only the call; the comparison operator and
  the `0` operand stay untouched) with
  `{counterpart}({arg1}, {arg2}, {n})`:
  - `{counterpart}` is `strncmp` / `strncasecmp`, prefixed with `\` when
    the original call was written fully qualified (`\strpos`, `\StrIPos`,
    …) or when an unqualified call at that position would not reach the
    global function (a `use function` import under that name, or a same-named function declared in the current namespace); otherwise it is bare.
  - `{arg1}`, `{arg2}`: verbatim source text of the two arguments.
  - `{n}`: decimal length of the needle's *decoded* value — the literal's
    content with PHP escape sequences of its quote style resolved
    (single-quoted: only `\\` and `\'`; double-quoted: `\n`, `\t`, `\\`, `\$`,
    `\"`, `\x..`, octal, `\u{..}`, …). See Divergences about multibyte text.
  - Separators are `, `.
  Examples: `strpos($u, 'http') === 0` → `strncmp($u, 'http', 4) === 0`;
  `0 !== \stripos($h, "x\ty")` → `0 !== \strncasecmp($h, "x\ty", 3)`.

## Options
None.

## PHP versions
No gating. The EA test runs at the PhpStorm default level (between 5.6 and 7.0).

## Examples

```php
<?php
function prefixes($url, $head, $part)
{
    return [
        <warning descr="Use 'strncmp($url, 'https', 5)' for a length-independent prefix check.">strpos($url, 'https')</warning> === 0,
        0 !== <warning descr="Use 'strncasecmp($head, 'Accept:', 7)' for a length-independent prefix check.">stripos($head, 'Accept:')</warning>,
        <warning descr="Use '\strncmp($url, 'C:\\', 3)' for a length-independent prefix check.">\strpos($url, 'C:\\')</warning> !== 0,
        0 === <warning descr="Use 'strncmp($url, 'it\'s', 4)' for a length-independent prefix check.">strpos($url, 'it\'s')</warning>,

        strpos($url, 'https') == 0,
        strpos($url, 'https', 2) === 0,
        strpos($url, $part) === 0,
        strpos($url, "{$part}/") === 0,
        (strpos($url, 'https')) === 0,
        strpos($url, 'https') === false,
        strpos($url, '') === 0,
    ];
}
```

```php
<?php
function prefixes($url, $head, $part)
{
    return [
        strncmp($url, 'https', 5) === 0,
        0 !== strncasecmp($head, 'Accept:', 7),
        \strncmp($url, 'C:\\', 3) !== 0,
        0 === strncmp($url, 'it\'s', 4),

        strpos($url, 'https') == 0,
        strpos($url, 'https', 2) === 0,
        strpos($url, $part) === 0,
        strpos($url, "{$part}/") === 0,
        (strpos($url, 'https')) === 0,
        strpos($url, 'https') === false,
        strpos($url, '') === 0,
    ];
}
```

## Divergences
- **Multibyte needles (upstream bug):** upstream counts the decoded needle in
  UTF-16 code units (≈ characters), but `strncmp` counts bytes, so `'é/'`
  gets `2` instead of `3` and the rewritten check compares too few bytes.
  Recommendation: use the byte length of the decoded value (identical for
  ASCII, which is all the fixtures use).
- **Empty needle:** custos diverges from upstream (D4a, E5). Upstream
  rewrites `strpos($s, '') === 0` to `strncmp($s, '', 0) === 0`, which is
  always true, whereas before PHP 8 `strpos()` with an empty needle returned
  `false` and emitted a warning, so the rewrite flips the result there. An
  empty prefix check has nothing to gain from a bounded comparison at any
  version, so custos does not report it. The EA conformance case still
  passes.
- **Qualifier kept (custos diverges):** upstream rewrites `\strpos(...)` to
  an unqualified `strncmp(...)`. That still works through the global
  fallback, but it drops the author's explicit qualification (which code
  styles use for opcache-inlined builtins) and lets a namespaced function of
  the same name take over. custos keeps a leading `\` (F1).
- Heredoc/nowdoc needles are untested upstream; recommendation: treat them
  like double-/single-quoted literals respectively.
- **Function-name matching — custos diverges from upstream.** Upstream matches
  `strpos`/`stripos` by the name as written (case-sensitive, any namespace
  qualifier, no resolution), so a differently cased call such as `StrPos($s,
  'x') === 0` is missed while a namespaced or imported user function of the
  same name is reported (and rewritten) as if it were the builtin. custos
  matches case-insensitively and only calls that reach the global function.
- **Builtin spelling (custos diverges).** Upstream inserts a bare `strncmp(`/`strncasecmp(` in its fix, so a function of that name declared in (or imported into) the namespace captures the rewritten call. custos writes `\name(` in that case (F1).
