---
id: DateIntervalSpecification
group: Probable bugs
kind: semantic
needs: [names, index]
php: { min: "", max: "" }
---

# DateIntervalSpecification

## Summary

`new DateInterval($spec)` throws for a malformed ISO-8601 duration string.
Check literal specifications against the accepted shapes so the mistake is
caught before runtime.

## Detection

Visit every `new` expression.

- **D1** Exactly one constructor argument is passed.
- **D2** The instantiated class name, resolved with the file's namespace and
  imports, is `\DateInterval`, compared case-insensitively as PHP compares
  class names (`new \dateinterval(...)` qualifies; custos diverges); so `new DateInterval(...)` inside a namespace without an
  import does not qualify; subclasses do not qualify).
- **D3** Obtain the specification literal `L`:
  - if the argument is itself a string literal, `L` is that literal;
  - otherwise run *value discovery* (as defined in the
    `CallableMethodValidity` spec) on the argument; among the discovered
    values keep only string literals; if exactly one remains, it is `L`.
  - otherwise stop.
- **D4** `L` contains no interpolation.
- **D5** Let `s` be the raw content of `L` (between the quotes, escape
  sequences not decoded). Report when `s` matches **neither** of these
  (whole-string, case-sensitive, ASCII digits):
  - **R1** `^P((\d+Y)?(\d+M)?(\d+D)?(\d+W)?)?(T(?=\d)(\d+H)?(\d+M)?(\d+S)?)?$`
    — `P`, then optional date parts in the fixed order Y, M, D, W, then an
    optional time part that starts with `T` immediately followed by a digit
    and has optional H, M, S parts in that order.
  - **R2** `^P\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}$` — the "datetime-like"
    form, e.g. `P0001-02-03T04:05:06`.

- **D6** When `L` was found by value discovery, report it only for the
  first `new DateInterval(...)` in the file (source order) that resolves to
  that same literal: a malformed literal shared by several call sites is
  reported once, at the literal.

Notable consequences of R1: a bare `P` is accepted; `PT` (no time digits) is
rejected; parts out of order (`P2D1M`, `P1W2D`) are rejected; lowercase
letters are rejected; fractions (`PT1.5S`) are rejected; `PT` followed by a
digit but no unit (`PT5`) is rejected.

## Exceptions (no report)

- **E1** Zero or 2+ arguments; other classes (`new DateTime('…')`).
- **E2** Interpolated strings, non-literal values that cannot be traced to a
  single string literal.
- **E3** Valid specifications per R1/R2.

## Report

- Range: the string literal `L` including its quotes. When `L` was found via
  value discovery (e.g. a variable assigned earlier in the function), the
  highlighted range is that literal at its own location, not the `new`
  argument.
- Severity: error.
- Message: `Malformed DateInterval specification.`

## Fix

None.

## Options

None.

## PHP versions

None.

## Examples

```php
<?php
function intervals($mode)
{
    $weekly = 'P1W';
    $broken = <error descr="Malformed DateInterval specification.">'1W'</error>;

    return [
        new DateInterval(<error descr="Malformed DateInterval specification.">'3M'</error>),
        new DateInterval(<error descr="Malformed DateInterval specification.">'P2D1Y'</error>),
        new DateInterval(<error descr="Malformed DateInterval specification.">'PT1D'</error>),
        new DateInterval(<error descr="Malformed DateInterval specification.">"PT"</error>),
        new DateInterval(<error descr="Malformed DateInterval specification.">'PT10:30:00'</error>),
        new DateInterval(<error descr="Malformed DateInterval specification.">'p1d'</error>),
        new DateInterval($weekly),
        new DateInterval($broken),
        new DateInterval("P{$mode}D"),
        new DateInterval('P1Y2M3DT4H5M6S'),
        new DateInterval('PT36H'),
        new DateInterval('P0002-00-10T12:00:00'),
        new DateTimeImmutable('3M'),
    ];
}
```

The literal assigned to `$broken` is the one highlighted:

```php
<?php
function later()
{
    $spec = <error descr="Malformed DateInterval specification.">'10D'</error>;
    return new DateInterval($spec);
}
```

## Divergences

- **Class name case (custos diverges):** upstream compares the resolved
  class name case-sensitively, so `new dateinterval('3M')` or an import
  written `use DATEINTERVAL;` escapes the check although PHP instantiates
  the same class. custos compares it case-insensitively (D2).
- The regular expressions use `$`, which in the upstream engine also matches
  just before a final line break; a literal whose raw content ends with an
  actual newline would be accepted. Recommendation: anchor at the true end of
  the string (no fixture covers it).
- **Shared literals (custos diverges):** upstream reports a literal reached
  through value discovery once per `new DateInterval(...)` that uses it, so
  the same range carries duplicate findings. custos reports it once (D6).
  The highlight stays on the literal itself, even far from the `new`
  expression: the literal is what needs fixing, and moving the range to
  the call site would only differ from upstream without helping.
- **Other discovered values (custos diverges, D3).** Upstream keeps the
  single string literal among the discovered values even when other,
  non-literal values were discovered too, so `$v = $q['v'] ?? ''; if ($v
  === '') { throw …; } new DateInterval($v)` reports `''` (error) although
  the guard keeps it from the constructor and the other value is the one
  used (Shopware query parser). custos reports a discovered literal only
  when it is the only discovered value.
