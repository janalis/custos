---
id: StringsFirstCharactersCompare
group: Probable bugs
kind: syntax
needs: []
php: { min: "", max: "" }
---

# StringsFirstCharactersCompare

## Summary
`strncmp()`/`strncasecmp()` with a literal prefix and a hard-coded length that
does not match the literal's length compares too few (or too many)
characters — usually a leftover after editing the literal.

## Detection
- **D1** A function call (not a method call) that resolves to the global
  function `strncmp` or `strncasecmp`: names compared case-insensitively, as
  PHP does (`StrNCmp(...)` matches); `\strncmp(...)` and an unqualified call
  in the global namespace match; an unqualified call in a namespace matches
  only when no function of that name is declared in the namespace or
  imported with `use function`; a qualified `Foo\strncmp(...)` does not
  match (custos diverges).
- **D2** Exactly 3 arguments.
- **D3** The third argument is an integer-looking number literal, or a unary
  minus applied directly to a number literal (`-3`).
- **D4** The literal string is the second argument if it is a string literal,
  otherwise the first argument if it is a string literal; if neither is a
  string literal, no report. (String literal = single- or double-quoted
  string; see Divergences for interpolated strings and heredocs.)
- **D5** Compute `L` = length of the literal's decoded value:
  - single-quoted: `\\` → `\`, `\'` → `'`, every other backslash kept;
  - double-quoted: standard PHP escape sequences decoded (`\n`, `\t`, `\r`,
    `\v`, `\e`, `\f`, `\\`, `\$`, `\"`, `\0`–`\777` octal, `\xHH`, `\u{...}`),
    unknown escapes kept as two characters.
  Length is counted in characters of the decoded text (see Divergences on
  multi-byte text).
- **D6** Parse the third argument's source text as a PHP integer literal `N`,
  with PHP's own rules: decimal, octal with a leading `0` (`010` is 8) or
  `0o`, hexadecimal `0x`, binary `0b`, and `_` separators between digits
  (`1_0` is 10); an optional leading `-` directly followed by the literal is
  accepted. Floats (`3.0`), a space between `-` and the digits, malformed
  literals, or values outside the 32-bit signed range → no report.
- **D7** Report when `L > 0` and `L != N`.

## Exceptions (no report)
- **E1** Length already equal to the literal length.
- **E2** Empty literal (`L == 0`).
- **E3** Third argument not a number literal (variable, constant,
  `strlen(...)`, expression) or not parseable as an integer literal (D6).
- **E4** Neither of the first two arguments is a string literal.
- **E5** Two or four+ arguments.

## Report
- Range: the third argument exactly as written (the number, or the whole
  `-N` unary expression).
- Severity: error.
- Message: `Length {N} does not match the {L}-character literal.` where `{N}`
  is the third argument's source text as written (`010`, `-3`).

## Fix
- **F1** Replace the third argument with the decimal text of `L`
  (`strncmp($s, 'abc', 5)` → `strncmp($s, 'abc', 3)`).

## Options
None.

## PHP versions
No gating.

## Examples

```php
<?php
function route(string $uri) {
    if (strncmp($uri, '/api/', <error descr="Length 4 does not match the 5-character literal.">4</error>) === 0) {}
    if (strncmp('/admin', $uri, <error descr="Length 9 does not match the 6-character literal.">9</error>) === 0) {}
    if (\strncasecmp($uri, "HTTP\t", <error descr="Length 6 does not match the 5-character literal.">6</error>) === 0) {}
    if (strncasecmp($uri, 'C:\\x', <error descr="Length -1 does not match the 4-character literal.">-1</error>) === 0) {}
    if (strncmp($uri, '/static/', 8) === 0) {}
    if (strncmp($uri, 'back\\slash', 10) === 0) {}
    if (strncmp($uri, '', 3) === 0) {}
    if (strncmp($uri, '/x', 0x2) === 0) {}
    if (strncmp($uri, '/v1/', strlen('/v1/')) === 0) {}
    if (strncmp($uri, $prefix, 4) === 0) {}
}
```

```php
<?php
function route(string $uri) {
    if (strncmp($uri, '/api/', 5) === 0) {}
    if (strncmp('/admin', $uri, 6) === 0) {}
    if (\strncasecmp($uri, "HTTP\t", 5) === 0) {}
    if (strncasecmp($uri, 'C:\\x', 4) === 0) {}
    if (strncmp($uri, '/static/', 8) === 0) {}
    if (strncmp($uri, 'back\\slash', 10) === 0) {}
    if (strncmp($uri, '', 3) === 0) {}
    if (strncmp($uri, '/x', 0x2) === 0) {}
    if (strncmp($uri, '/v1/', strlen('/v1/')) === 0) {}
    if (strncmp($uri, $prefix, 4) === 0) {}
}
```

## Divergences
- **Function name (custos diverges).** Upstream matches the written name
  case-sensitively and accepts any namespace qualifier without resolution,
  so `STRNCMP($s, 'abc', 2)` is missed while a namespaced user function
  named `strncmp` (with its own meaning for the third argument) is reported
  and "fixed". custos compares the name case-insensitively and requires the
  call to reach the builtin (D1).
- Upstream counts UTF-16 code units of the decoded literal, whereas PHP's
  length argument counts bytes. For ASCII literals both agree. For
  non-ASCII literals (`'é'`) upstream would demand 1 where PHP needs 2.
  Recommendation: count bytes of the UTF-8 decoded value (correct PHP
  semantics); document the divergence (no upstream fixture uses non-ASCII).
- Interpolated double-quoted strings (`"/$base/"`) and heredoc/nowdoc literals:
  upstream measures the raw text including the `$var` placeholders.
  Recommendation: skip literals containing interpolation; treat nowdoc like
  single-quoted and heredoc like double-quoted only when not interpolated.
- **Integer literals read as PHP reads them — custos diverges from upstream**
  (D6). Upstream reads the length as a decimal number, so `010` counts as 10
  although PHP passes 8: `strncmp($s, 'abcdefgh', 010)` is correct but was
  reported (and "fixed" to `8`, the same value), while `strncmp($s, 'abc',
  010)` was reported with the wrong length. Hex, binary and separated
  literals were never checked. custos evaluates the literal with PHP's
  integer grammar.
