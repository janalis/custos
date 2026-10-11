---
id: ImplodeArgumentsOrder
group: Code style
kind: syntax
needs: []
php: { min: "", max: "" }
---

# ImplodeArgumentsOrder

## Summary

`implode()` historically accepted its separator in either position; passing the
pieces first and the separator second is deprecated (7.4) and removed (8.0).
When the second argument is clearly the separator (a string literal), the
arguments are swapped.

## Detection

D1. Node: a plain function call (not a method or static call) that resolves
    to the global function `implode`. The name is compared case-insensitively
    (`IMPLODE`, `Implode` count) and may be written `\implode`; a qualified
    namespaced call (`App\implode`), a non-global `use function` import, or
    an unqualified call in a namespace that declares its own `implode()`
    does not count.
D2. The call has exactly two arguments.
D3. The **second** argument is a string literal (any quoting style, including
    double-quoted strings with interpolation, heredoc, nowdoc).
D4. The **first** argument is not a string literal (same definition as D3).

## Exceptions (no report)

E1. Zero, one, or three-plus arguments.
E2. Second argument not a string literal (variable, constant, concatenation,
    array, call, …): `implode($glue, $parts)`, `implode($parts, SEP)`.
E3. Other functions: the alias `join(...)`; a user function named `implode`
    that the call resolves to (see D1).
E4. Method/static calls named `implode` (`$s->implode([], ',')`).
E5. Both arguments are string literals (`implode(',', 'a')`): nothing
    identifies the second one as the separator, and swapping would not make
    the call valid (the pieces must be an array).
E6. Either argument is named or unpacked
    (`implode(array: $parts, separator: ', ')`): a named separator is already
    bound to the right parameter whatever its position, and positions say
    nothing about what an unpacked argument fills.

## Report

- Range: the whole call expression, from the first character of the function
  name (including a leading namespace qualifier such as `\`) to the closing `)`.
- Severity: info (fixtures tag it `weak_warning`).
- Message: `Pass the separator as the first argument of implode().`

## Fix

F1. Replace the whole call with
    `{name}({second}, {first})`
    where `{name}` is the function name exactly as written, qualifier and
    letter case included (`\implode`, `Implode`), `{second}` and `{first}`
    are the verbatim source texts of the second and first arguments, joined by
    `, ` (comma + one space). Any whitespace/comments originally inside the
    parentheses outside the argument texts are not preserved.
F2. No fix (the report is kept) when the target PHP version is below 7.3 and
    either argument contains a heredoc or nowdoc (the separator itself, or a
    first argument such as `$head + <<<A … A`): until 7.3 the closing marker
    must be the last thing on its line, so neither `, ` nor `)` can be
    written after it.

## Options

| Option | Type | Default | Effect |
|---|---|---|---|

## PHP versions

Reported at every level. Below 7.3 the fix is withheld when an argument
contains a heredoc/nowdoc (F2).

## Examples

```php
<?php

$csv   = <weak_warning descr="Pass the separator as the first argument of implode().">implode($cells, ';')</weak_warning>;
$path  = <weak_warning descr="Pass the separator as the first argument of implode().">\implode( $segments ,  "/" )</weak_warning>;
$ok1   = implode(' | ', $labels);
$ok2   = implode($labels, $sep);
$ok3   = implode($labels);
$ok4   = join($labels, '-');
```

```php
<?php

$csv   = implode(';', $cells);
$path  = \implode("/", $segments);
$ok1   = implode(' | ', $labels);
$ok2   = implode($labels, $sep);
$ok3   = implode($labels);
$ok4   = join($labels, '-');
```

## Divergences

- Two string literals (custos diverges from upstream). Upstream reports
  `implode('a', 'b')` and swaps the arguments, although the call is already
  in the conventional order and swapping only moves the mistake (a string is
  passed as the pieces either way). custos does not report it (D4/E5).
- **Letter case and name resolution (custos diverges).** Upstream compares
  the written name part with `implode` case-sensitively and accepts any
  namespace qualifier, so `IMPLODE($parts, ',')` is missed while a call to a
  user function `App\implode()` is reported and rewritten. PHP function
  names are case-insensitive and a namespaced function is a different
  function; custos matches calls that resolve to the global `implode` in any
  letter case (D1).
- **Named arguments (custos diverges).** Upstream looks at argument
  positions only, so a call that names its arguments in "pieces, separator"
  order is reported and rewritten positionally. Such a call is valid and
  not deprecated; custos does not report it (E6).
- **Heredoc separator below PHP 7.3 (custos diverges).** Upstream always
  offers the swap. Moving a heredoc/nowdoc in front of another argument
  or of the closing parenthesis is a parse error before 7.3, so custos
  keeps the report but offers no fix there (F2).
