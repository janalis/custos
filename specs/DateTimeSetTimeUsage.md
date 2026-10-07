---
id: DateTimeSetTimeUsage
group: Probable bugs
kind: semantic
needs: [names, types, hierarchy, index, stubs]
php: { min: "", max: "7.0" }
---

# DateTimeSetTimeUsage

## Summary
The microseconds argument of `DateTime::setTime()` / `date_time_set()` only
exists since PHP 7.1. On older versions passing it makes the call fail and
return `false`, so the extra argument is a bug when targeting PHP < 7.1.

## Detection
Only active when the configured target PHP version is **below 7.1**
(the whole rule is silent for 7.1 and later).

### Method form
- **D1** A method call (`->` or `::`) whose name is `setTime`, compared
  case-insensitively as PHP compares method names (`->SETTIME(...)`
  qualifies; custos diverges).
- **D2** Exactly four arguments.
- **D3** The call resolves (receiver type + class hierarchy) to the method
  declared in `\DateTime` (identity `\DateTime::setTime`). Calls on user
  subclasses that inherit it qualify; `DateTimeImmutable::setTime`, a
  subclass override, or an unresolvable receiver do not.

### Function form
- **D4** A function call whose name (last segment) is `date_time_set`,
  compared case-insensitively as PHP compares function names (custos
  diverges).
- **D5** Exactly five arguments.
- **D6** The call resolves to the global function `\date_time_set` (a
  same-named function declared in a namespace and resolved there does not
  qualify; a `use function` import line itself is not a call).

## Exceptions (no report)
- **E1** Target PHP 7.1+.
- **E2** Three-argument `setTime()` / four-argument `date_time_set()` (or any
  other argument count).
- **E3** `DateTimeImmutable` receivers, unknown receivers.

## Report
- Range: the extra argument only — the 4th argument of `setTime()`, the 5th
  of `date_time_set()` — whatever expression it is (literal, `null`,
  variable, …).
- Severity: error.
- Message: `Microseconds argument requires PHP 7.1+; on this version the call returns false.`

## Fix
None.

## Options
None.

## PHP versions
Active only for target versions `< 7.1` (`php.max = 7.0`). Note: upstream
test cases without an explicit language level run under the IDE's test
default, which is below 7.1, so this rule is active for them; the single
upstream case for this rule sets 7.0 explicitly. Custos' default target
version may be higher — conformance runs must use the case's PHP level.

## Examples
Target PHP 7.0:

```php
<?php
class Clock extends DateTime {}

function rewind_clock(DateTime $d, Clock $c, DateTimeImmutable $i, $usec)
{
    $d->setTime(8, 30, 0, <error descr="Microseconds argument requires PHP 7.1+; on this version the call returns false.">$usec</error>);
    $c->setTime(8, 30, 0, <error descr="Microseconds argument requires PHP 7.1+; on this version the call returns false.">250</error>);
    $d->setTime(8, 30, 0);
    $i->setTime(8, 30, 0, 250);
    date_time_set($d, 8, 30, 0, <error descr="Microseconds argument requires PHP 7.1+; on this version the call returns false.">0</error>);
    \date_time_set($d, 8, 30, 0, <error descr="Microseconds argument requires PHP 7.1+; on this version the call returns false.">null</error>);
    date_time_set($d, 8, 30, 0);
}
```

Target PHP 7.1: the same code produces no findings.

## Divergences
- **Name case (custos diverges).** Upstream compares `setTime` and
  `date_time_set` case-sensitively, so `$d->SetTime(1, 2, 3, 4)` or
  `Date_Time_Set(...)` escape the check although PHP calls the same method
  or function. custos compares both names case-insensitively (D1, D4).
