---
id: IsoWeekCalendarYearMismatch
group: Probable bugs
kind: syntax
needs: [names]
php: { min: "", max: "" }
---

# IsoWeekCalendarYearMismatch

## Summary

Detect a date format literal containing unescaped Y and W in the same year-week label.

## Detection

- D1. Report a date format literal containing unescaped Y and W in the same year-week label. The Y counts as a week label only when the format has no other part that belongs to the calendar year (see E1); otherwise Y is the correct year for that part and the format is left alone.
- D2. Resolve builtin calls and named arguments before checking the contract. Project-local callees are considered only when their bodies or immutable summaries establish the required facts. The default is enabled, with positive proof required.
- D3. For flow-dependent conditions, require a reachable source-to-use path with no intervening invalidation. Unknown calls, ambiguous aliases, recursion or analysis-budget exhaustion cannot establish definite facts.

## Exceptions (no report)

- E1. Escaped tokens and formats already using the ISO week-year o are excluded. So is any format with an unescaped calendar-dependent token: a month (`m`, `n`, `F`, `M`), a day of the month (`d`, `j`), the day of the year (`z`), the month length (`t`), the leap-year flag (`L`), or a composite date (`c`, `r`, `U`). There Y labels the month, day or full date, and swapping it for o would print that part with the wrong year, as in `F Y, \w\e\e\k W`, `Y-m W`, `Y/z W` or `c W Y`. Weekday tokens (`D`, `l`, `N`, `w`), time and timezone tokens do not exclude the format.
- E2. Unresolved targets, incomplete required facts, suppressed findings and PHP versions lacking the referenced API or language feature produce no report. User declarations shadowing builtin names are not builtin contracts.

## Report

- Range: the complete violating call or expression identified by D1; for a declaration-state violation, the offending property access or assignment. Highlight the final unsafe use for source-to-use findings.
- Severity: warning.
- Message: `Use the ISO week year with the ISO week number.`

## Fix

Replace every unescaped Y token with o in the literal; preserve all other format bytes. E1 guarantees that Y solely labels the ISO week in a reported format, so the replacement cannot change the year of a month, day or composite date. The fix is offered only when the format argument is a plain quoted literal whose source text equals its value; a format reached through a variable, constant, concatenation or a literal with escape sequences is reported without a fix.

A fix is offered only when the concrete prerequisites above are proven. Edits use exact source byte ranges, preserve comments and argument evaluation order, and introduce syntax supported by the configured PHP version. Never invent error-handling policy.

## Options

None. Enable or disable this rule through normal rule configuration.

## PHP versions

Apply only on PHP versions supporting the referenced APIs and syntax (within PHP 5.3–8.5). Builtin behavior and argument contracts follow the configured PHP version. Extension-specific calls require resolved extension API symbols; do not treat same-named application functions as extension calls.

## Examples

```php
<?php
$day=new DateTimeImmutable(); $label=<warning descr="Use the ISO week year with the ISO week number.">$day->format('Y-W')</warning>;
```

```php
<?php
$day=new DateTimeImmutable(); $label=$day->format('o-W');
```

```php
<?php
// Y labels the month here, not the week.
$label = date('F Y, \w\e\e\k W');
$stamp = gmdate('c W Y');
```

## Divergences

Native custos rule. No upstream inspection or fixture comparison applies. These conditions are independently specified and examples are original.
