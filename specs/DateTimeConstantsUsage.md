---
id: DateTimeConstantsUsage
group: Probable bugs
kind: semantic
needs: [names, index, hierarchy, stubs]
php: { min: "", max: "" }
---

# DateTimeConstantsUsage

## Summary
The `ISO8601` date format constants do not actually produce ISO-8601 output
(the offset lacks a colon). The `ATOM` variants produce the compliant format
and should be used instead.

## Detection

### Global constant
- **D1** A constant reference that resolves to the global constant
  `DATE_ISO8601` (case-sensitive): written `\DATE_ISO8601`, or unqualified
  with PHP's global fallback (no constant `DATE_ISO8601` declared in the
  current namespace), or through a `use const DATE_ISO8601 [as X];` import.
  A qualified `X\DATE_ISO8601`, a `use const` import of another constant
  under that name, or a namespace declaring its own `DATE_ISO8601` is not
  reported (those are user constants).

### Class constant
- **D2** A class constant access `X::ISO8601` (constant name exactly
  `ISO8601`, case-sensitive) — `X` may be a class name, `self`/`static`/
  `parent`, or an expression such as `$date::ISO8601`.
- **D3** The access resolves (through the class hierarchy, using the inferred
  type for expression receivers) to a class constant declaration whose
  identity is `\DateTime::ISO8601` or `\DateTimeInterface::ISO8601` (i.e. the
  constant declared in one of those two built-in types). This covers
  `DateTime::ISO8601`, `DateTimeInterface::ISO8601`, `DateTimeImmutable::ISO8601`
  (inherited from the interface) and user subclasses of `DateTime` that do not
  redeclare the constant.

## Exceptions (no report)
- **E1** Constants with other names (`__DATE_ISO8601`, `DATE_ISO8601_X`,
  `date_iso8601`).
- **E2** `ISO8601` constants declared by user classes (including a redeclared
  `ISO8601` in a `DateTime` subclass) or unresolvable accesses.

## Report
- Range: D1 — the whole constant reference including any leading `\` or
  namespace qualifier. D2 — the whole class constant access, from the start
  of `X` to the end of `ISO8601` (e.g. `DateTime::ISO8601`).
- Severity: error.
- Messages:
  - D1: `DATE_ISO8601 is not ISO-8601 compliant; use DATE_ATOM.`
  - D2: `ISO8601 is not ISO-8601 compliant; use the ATOM constant.`

## Fix
- **F1** D1: replace the constant's name by `DATE_ATOM`:
  `DATE_ISO8601` → `DATE_ATOM`, `\DATE_ISO8601` → `\DATE_ATOM`. An
  unqualified (or aliased) reference becomes `\DATE_ATOM` when a bare
  `DATE_ATOM` there would not reach the global constant (a `use const`
  import under that name, or a `DATE_ATOM` constant declared in the current
  namespace).
- **F2** D2: replace only the constant name `ISO8601` by `ATOM`, keeping the
  class part and the `::` (and any whitespace around it) untouched:
  `DateTime::ISO8601` → `DateTime::ATOM`, `$when::ISO8601` → `$when::ATOM`.

## Options
None.

## PHP versions
None.

## Examples

```php
<?php
class Stamp { const ISO8601 = 'Y'; }
class LocalTime extends DateTime {}

function formats(DateTimeImmutable $at)
{
    return [
        <error descr="DATE_ISO8601 is not ISO-8601 compliant; use DATE_ATOM.">DATE_ISO8601</error>,
        <error descr="DATE_ISO8601 is not ISO-8601 compliant; use DATE_ATOM.">\DATE_ISO8601</error>,
        <error descr="ISO8601 is not ISO-8601 compliant; use the ATOM constant.">DateTime::ISO8601</error>,
        <error descr="ISO8601 is not ISO-8601 compliant; use the ATOM constant.">DateTimeInterface::ISO8601</error>,
        <error descr="ISO8601 is not ISO-8601 compliant; use the ATOM constant.">LocalTime::ISO8601</error>,
        Stamp::ISO8601,
        DATE_ISO8601_LEGACY,
        DATE_RFC3339,
        $at->format('c'),
    ];
}
```

```php
<?php
class Stamp { const ISO8601 = 'Y'; }
class LocalTime extends DateTime {}

function formats(DateTimeImmutable $at)
{
    return [
        DATE_ATOM,
        \DATE_ATOM,
        DateTime::ATOM,
        DateTimeInterface::ATOM,
        LocalTime::ATOM,
        Stamp::ISO8601,
        DATE_ISO8601_LEGACY,
        DATE_RFC3339,
        $at->format('c'),
    ];
}
```

## Divergences
- Depending on the stub set, `DateTimeImmutable` may declare its own copy of
  the constants (identity `\DateTimeImmutable::ISO8601`), in which case
  upstream would not report `DateTimeImmutable::ISO8601`. Recommendation:
  report it as well (it is the same format); not covered by upstream
  fixtures.
- **Namespaced constants (custos diverges).** Upstream compares only the
  last segment of the written name, so `X\DATE_ISO8601` (or a namespace's
  own `DATE_ISO8601`) is reported and rewritten to a constant `X\DATE_ATOM`
  that may not exist. custos reports only references that resolve to the
  global constant (D1) and spells the replacement so it reaches the global
  `DATE_ATOM` (F1).
