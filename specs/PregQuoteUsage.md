---
id: PregQuoteUsage
group: Probable bugs
kind: semantic
needs: [names, stubs]
php: { min: "", max: "" }
---

# PregQuoteUsage

## Summary
`preg_quote()` only escapes the pattern delimiter when it is given as the
second argument. Called with a single argument, a `/` or `#` inside the quoted
text breaks the surrounding pattern.

## Detection
- **D1** A plain function call (not a method or static call) whose name (last
  segment) is `preg_quote`, compared case-insensitively as PHP does
  (`Preg_Quote(...)` matches; custos diverges), and which resolves to the
  global `\preg_quote` under PHP's runtime rules: `\preg_quote(...)` and an
  unqualified call in the global namespace match; an unqualified call inside
  a namespace matches only when no `preg_quote` function is declared in that
  namespace or imported from elsewhere; a qualified `Foo\preg_quote(...)`
  does not match.
- **D2** The call has exactly **one** argument (whatever its form, including
  a spread `...$args`).

## Exceptions (no report)
- **E1** Two or more arguments (`preg_quote($s, '/')`), or zero arguments.
- **E2** Method calls named `preg_quote`.

## Report
- Range: the function name token `preg_quote` only (not the argument list;
  whitespace between the name and `(` is not included). For a qualified call
  such as `\preg_quote(...)`, highlight only the trailing name identifier.
- Severity: error.
- Message: `Pass the pattern delimiter to preg_quote() as its second argument.`

## Fix
None.

## Options
None.

## PHP versions
None.

## Examples

```php
<?php
$re1 = '~' . <error descr="Pass the pattern delimiter to preg_quote() as its second argument.">preg_quote</error>($needle) . '~i';
$re2 = <error descr="Pass the pattern delimiter to preg_quote() as its second argument.">preg_quote</error> ( $host );
$re3 = '@^' . preg_quote($path, '@') . '$@';
$re4 = $quoter->preg_quote($needle);
```

## Divergences
- **Callee resolved — custos diverges from upstream** (D1). Upstream matches
  the callee by name only, so a user function `App\preg_quote($s)` with its
  own signature (e.g. a wrapper that supplies the delimiter itself) is
  reported. custos reports only calls that reach the built-in.
- **Name case — custos diverges** (D1). Upstream compares the name
  case-sensitively, so `PREG_QUOTE($s)` is missed even though PHP calls the
  same built-in; custos compares it case-insensitively.
