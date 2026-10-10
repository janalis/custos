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

## Exceptions (no report)

- E1. Exclude shadowed builtins, unresolved receivers, incomplete syntax and unknown prerequisite facts. An extension function name alone does not prove an application function is that builtin.
- E2. A dominating valid guard or repair described in Detection prevents the finding. Both legacy noinspection and custos-ignore forms suppress this ID.

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

## Divergences

Native custos inspection, independently specified. No EA counterpart or EA fixture conformance applies.
