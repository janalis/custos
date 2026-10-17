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
    values keep only string literals; if exactly one remains, no other
    value was discovered, and the discovery is *settled* (below), it is `L`.
  - otherwise stop.

  Value discovery only follows plain `=` assignments, so the discovery is
  settled only when every variable it went through (the argument, then the
  variables assigned to it, across ternary and `??` operands) is a local
  that can hold nothing else:
  - it is not a parameter of the enclosing function (a default is only one
    of the values callers pass) nor a variable captured by `use`;
  - it has at least one plain `=` assignment in its own scope, and no other
    kind of write there: foreach key/value target, `list()`/`[]`
    destructuring slot, reference assignment target, `global`/`static`
    declaration, catch variable, `unset()`, an offset write (`$v[0] = …`);
  - no reference to it is handed out: right-hand side of `$r = &$v`,
    `[&$v]`, `use (&$v)`, a call-time `&$v` argument, or an argument bound
    to a by-reference parameter (positional, named or variadic tail) of a
    callee resolved through the index (user function, method, static
    method or builtin). An unresolved callee cannot prove by-value binding;
    passing the variable to it makes value discovery unknown. A resolved
    by-value call preserves the candidate specification;
  - its scope does not call `extract()`.
- **D4** `L` contains no interpolation.
- **D5** Let `s` be the raw content of `L` (between the quotes, escape
  sequences not decoded). Report when `s` matches **neither** of these
  (whole-string, case-sensitive, ASCII digits):
  - **R1** `^P((\d+Y)?(\d+M)?(\d+W)?(\d+D)?)?(T(?=\d)(\d+H)?(\d+M)?(\d+S)?)?$`
    — `P`, then optional date parts in the fixed order Y, M, W, D, then an
    optional time part that starts with `T` immediately followed by a digit
    and has optional H, M, S parts in that order.
  - **R2** `^P\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}$` — the "datetime-like"
    form, e.g. `P0001-02-03T04:05:06`.

- **D6** When `L` was found by value discovery, report it only for the
  first `new DateInterval(...)` in the file (source order) that resolves to
  that same literal: a malformed literal shared by several call sites is
  reported once, at the literal.

- **D7** When the target PHP version is below 8.0 and `s` matches R1 with
  both a week and a day part (`P1W2D`), report it with the second message:
  PHP only combines the two since 8.0; older versions accept the string
  but silently discard the weeks. D6 applies as well.

Notable consequences of R1: a bare `P` is accepted; `PT` (no time digits) is
rejected; parts out of order (`P2D1M`, `P2D1W`) are rejected; lowercase
letters are rejected; fractions (`PT1.5S`) are rejected; `PT` followed by a
digit but no unit (`PT5`) is rejected.

## Exceptions (no report)

- **E1** Zero or 2+ arguments; other classes (`new DateTime('…')`).
- **E2** Interpolated strings, non-literal values that cannot be traced to a
  single string literal.
- **E4** A discovered literal that is not provably the value used (D3): a
  parameter default, a variable also written by a foreach, a destructuring,
  a reference or a by-reference call.
- **E3** Valid specifications per R1/R2.

## Report

- Range: the string literal `L` including its quotes. When `L` was found via
  value discovery (e.g. a variable assigned earlier in the function), the
  highlighted range is that literal at its own location, not the `new`
  argument.
- Severity: error.
- Message: `Malformed DateInterval specification.`
- Message (D7): `Weeks and days cannot be combined in a DateInterval
  specification before PHP 8.0.`

## Fix

None.

## Options

None.

## PHP versions

- Below 8.0, a specification with both weeks and days is reported (D7).

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

A sentinel default or a variable rewritten through a reference is not the
value the constructor receives:

```php
<?php
function fill(&$spec) { $spec = 'P1D'; }

function ttl(string $spec = '')
{
    if ($spec === '') {
        return null;
    }
    return new DateInterval($spec);
}

function filled()
{
    $spec = 'pending';
    fill($spec);
    return new DateInterval($spec);
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
- **Week and day order (custos diverges, R1/D7).** Upstream expects the day
  part before the week part, so it reports `P1W2D` and `P1Y2M3W4D`, which
  PHP 8 accepts, and accepts `P2D1W`, which every PHP version rejects.
  custos follows PHP: weeks come before days, and the combination is only
  reported for targets below 8.0, where the weeks are lost.
- **Unsettled discovery (custos diverges, D3).** Upstream trusts the
  discovered literal whenever it is the single one, so a parameter default
  guarded as a sentinel (`function ttl(string $spec = '') { if ($spec ===
  '') { return null; } return new DateInterval($spec); }`) and an initial
  value later replaced through a foreach, a destructuring, a reference
  alias or a by-reference call are reported although they never reach the
  constructor. custos reports only a settled discovery.
