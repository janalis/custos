---
id: UnNecessaryDoubleQuotes
group: Code style
kind: syntax
needs: []
php: { min: "", max: "" }
---

# UnNecessaryDoubleQuotes

## Summary

A double-quoted string that contains no interpolation and no escape sequence
behaves exactly like a single-quoted one. Single quotes make it obvious that
nothing is evaluated inside.

## Detection

Visit every string literal.

- **D1** The literal is double-quoted (`"..."`, optionally with a `b` prefix —
  see Divergences). Single-quoted strings, heredoc and nowdoc are ignored.
- **D2** The literal contains no interpolated part (no `$var`, `{$expr}`,
  `${...}` embedded expression). A `$` that does not start an interpolation
  (`"$"`, `"5$"`, `"a $ b"`) is plain text.
- **D3** Take the raw text between the quotes. Replace every `\$` with `$` and
  every `\"` with `"`. The result must contain **no** single quote `'` and
  **no** backslash `\`. (So any other escape such as `\n`, `\t`, `\\`, `\x41`,
  `\u{..}`, `\0`, or a lone backslash prevents the report.)
- **D4** The literal is real code, not text inside a doc comment.
- The empty string `""` qualifies.

## Exceptions (no report)

- **E1** Strings with interpolation.
- **E2** Strings containing `'` (after D3's replacements) or any remaining
  backslash.
- **E3** Heredoc / nowdoc / single-quoted strings.

## Report

- Range: the whole literal including both double quotes.
- Severity: info.
- Message: `No interpolation or escapes here; use single quotes.`

## Fix

- **F1** Replace the literal by `'` + content + `'`, where content is the raw
  text with `\$` → `$` and `\"` → `"` applied (D3 guarantees nothing else needs
  escaping). Examples: `"plain"` → `'plain'`; `"\$\"x"` → `'$"x'`; `""` → `''`.

## Options

None.

## PHP versions

None.

## Examples

```php
<?php
$label  = <weak_warning descr="No interpolation or escapes here; use single quotes.">"ready"</weak_warning>;
$price  = <weak_warning descr="No interpolation or escapes here; use single quotes.">"\$9 \"net\""</weak_warning>;
$none   = <weak_warning descr="No interpolation or escapes here; use single quotes.">""</weak_warning>;
$dollar = <weak_warning descr="No interpolation or escapes here; use single quotes.">"cost: $"</weak_warning>;

$kept1 = 'already single';
$kept2 = "hello {$name}";
$kept3 = "it's";
$kept4 = "line\n";
$kept5 = "C:\\temp";
$kept6 = <<<TXT
body
TXT;
```

```php
<?php
$label  = 'ready';
$price  = '$9 "net"';
$none   = '';
$dollar = 'cost: $';
```

## Divergences

- Binary-prefixed strings (`b"..."`): upstream behaviour unverified. Recommend
  treating them like ordinary double-quoted strings and keeping the `b` prefix
  in the fix (`b"x"` → `b'x'`).
