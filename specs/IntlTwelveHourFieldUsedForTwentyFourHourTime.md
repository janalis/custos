---
id: IntlTwelveHourFieldUsedForTwentyFourHourTime
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "5.5", max: "" }
---

# IntlTwelveHourFieldUsedForTwentyFourHourTime

## Summary

Use the 24-hour calendar field.

## Detection

- D1. Report resolved IntlCalendar::set with direct builtin FIELD_HOUR class constant and proven integer value 12 through 23. FIELD_HOUR_OF_DAY, other fields and unknown values are excluded.
- D2. Resolve builtin functions, extension classes, method receivers, constants and named arguments. Follow unchanged locals and aliases only within the same bounded lexical scope. Require a reachable path; merge branch facts only when all incoming paths establish them.
- D3. Emit one finding at the violating operation, not one per evidence statement. Rebinding, writes, by-reference escape, unresolved calls that can mutate relevant state, and exhausted analysis budgets remove proof.
- D4. A value passed as a local variable is proven only while plain assignments are the sole way that variable changes in its scope. Anywhere in the same scope, an increment or decrement of the variable, a catch clause binding it, or a closure importing it by reference removes proof regardless of position, so the hour is treated as unknown. Closures importing it by value and arrow functions work on a copy and do not count.

## Exceptions (no report)

- E1. Exclude shadowed builtins, unresolved receivers, incomplete syntax and unknown prerequisite facts. An extension function name alone does not prove an application function is that builtin.
- E2. A dominating valid guard or repair described in Detection prevents the finding. Both legacy noinspection and custos-ignore forms suppress this ID.
- E3. No report when the hour variable is stepped or shared as described in D4: `$h = 12; $h--;` leaves 11, a valid 12-hour value, and renaming the field would move the time by twelve hours.

## Report

- Range: the violating expression or call described in D1; stateful absence checks highlight the final unsafe operation or loop.
- Severity: warning.
- Enabled by default: true.
- Message: `Use the 24-hour calendar field.`

## Fix

Replace only the FIELD_HOUR class constant name with FIELD_HOUR_OF_DAY on the same resolved class qualifier; retain qualifier and all other arguments.

All edits use exact byte ranges and preserve comments and evaluation order. Offer no fix when its stated prerequisites cannot be proven or the replacement is unsupported by the configured PHP version.

## Options

None.

## PHP versions

Requires PHP 5.5 or later and the referenced extension API. Respect method-specific introduction versions within PHP 5.3–8.5; unsupported optional APIs cannot establish facts. Named arguments require PHP 8.0 or later.

## Examples

```php
<?php
$c = IntlCalendar::createInstance(); <warning descr="Use the 24-hour calendar field.">$c->set(IntlCalendar::FIELD_HOUR, 19)</warning>;
```

Fixed result:

```php
<?php
$c = IntlCalendar::createInstance(); $c->set(IntlCalendar::FIELD_HOUR_OF_DAY, 19);
```

Valid case:

```php
<?php
$c = IntlCalendar::createInstance(); $c->set(IntlCalendar::FIELD_HOUR_OF_DAY, 19);
```

Not reported, the hour is no longer the assigned literal:

```php
<?php
function noon(IntlCalendar $c) { $h = 12; $h--; $c->set(IntlCalendar::FIELD_HOUR, $h); }
```

## Divergences

Native custos inspection, independently specified. No EA counterpart or EA fixture conformance applies.
