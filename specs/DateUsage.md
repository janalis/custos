---
id: DateUsage
group: Unused
kind: syntax
needs: []
php: { min: "", max: "" }
---

# DateUsage

## Summary

`date()` already formats the current moment when its timestamp argument is
omitted, so passing `time()` explicitly is redundant noise (and costs an extra
call). Drop the second argument.

## Detection

Visit every function call (not method or static calls).

- **D1** The call resolves to the global function (names compared
  case-insensitively, as PHP does; `\` and a global `use function` import are
  fine, but a same-named function declared in the current namespace, one
  imported from another namespace, or a qualified non-global name such as
  `Ns\date` does not count): `date` (`date`, `\date`, `Date` match).
- **D2** The call has exactly two arguments.
- **D3** The second argument is itself a plain function call (not a method
  call, static call, `new`, or a call wrapped in parentheses) resolving (as
  in D1) to the global `time`.
- **D4** That inner `time` call has zero arguments.

## Exceptions (no report)

- **E1** `date()` with one or three-plus arguments.
- **E2** Second argument `time(...)` with any argument (`time($x)`).
- **E3** Second argument is a method/static call named `time`
  (`$clock->time()`, `Clock::time()`), a variable, or any other expression.
- **E4** Calls to other functions (`gmdate($f, time())` is not reported).

## Report

- Range: the inner `time()` call expression, from the first character of its
  name (including a leading `\` if written) to its closing `)`.
- Severity: info (rendered as an "unused" highlight; fixture markup
  `weak_warning`).
- Message: `Redundant time() argument: date() uses the current time by default.`

## Fix

- **F1** Delete everything after the end of the first argument up to and
  including the end of the second argument: the comma, any whitespace or
  comments around it, and the `time()` call. Text after the second argument
  (e.g. whitespace before `)`) is kept.
  - `date('Y', time())` → `date('Y')`
  - `date($fmt ,  \time() )` → `date($fmt )`

## Options

None.

## PHP versions

None.

## Examples

```php
<?php
function stamp($pattern, Clock $clock)
{
    echo date('D, d M', <weak_warning descr="Redundant time() argument: date() uses the current time by default.">time()</weak_warning>);
    echo \date($pattern, <weak_warning descr="Redundant time() argument: date() uses the current time by default.">\time()</weak_warning>);
    echo date('H:i', time(42));
    echo date('H:i', $clock->time());
    echo date('H:i', Clock::time());
    echo gmdate('H:i', time());
    echo date('H:i');
}
```

```php
<?php
function stamp($pattern, Clock $clock)
{
    echo date('D, d M');
    echo \date($pattern);
    echo date('H:i', time(42));
    echo date('H:i', $clock->time());
    echo date('H:i', Clock::time());
    echo gmdate('H:i', time());
    echo date('H:i');
}
```

## Divergences

- With a trailing comma (`date('Y', time(),)`), upstream's deletion stops at
  the second argument and leaves `date('Y',)`. Recommendation: same behaviour
  (still valid PHP 7.3+); not covered by fixtures.
- **Function-name matching — custos diverges from upstream.** Upstream matches
  `date`/`time` by the name as written (case-sensitive, any namespace
  qualifier, no resolution), so a differently cased call such as `Date('Y',
  Time())` is missed while a namespaced or imported user function of the same
  name is reported (and rewritten) as if it were the builtin. custos matches
  case-insensitively and only calls that reach the global function.
