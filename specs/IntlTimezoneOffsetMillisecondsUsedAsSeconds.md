---
id: IntlTimezoneOffsetMillisecondsUsedAsSeconds
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "5.5", max: "" }
---

# IntlTimezoneOffsetMillisecondsUsedAsSeconds

## Summary

Convert timezone offsets to seconds.

## Detection

- D1. Track direct rawOffset/dstOffset output locals of a preceding resolved IntlTimeZone::getOffset call in the same lexical scope. Report DateTime/DateTimeImmutable::modify with either local, or their sum, concatenated with the literal seconds unit. Already-divided offsets and unknown provenance are excluded.
- D2. Resolve builtin functions, extension classes, method receivers, constants and named arguments. Follow unchanged locals and aliases only within the same bounded lexical scope. Require a reachable path; merge branch facts only when all incoming paths establish them.
- D3. Emit one finding at the violating operation, not one per evidence statement. Rebinding, writes, by-reference escape, unresolved calls that can mutate relevant state, and exhausted analysis budgets remove proof.

## Exceptions (no report)

Reject offset-unit proof after an intervening use of the output local, a reference alias, a nested scope or proof-budget exhaustion. Compound assignments and opaque calls discard the proof.

- E1. Exclude shadowed builtins, unresolved receivers, incomplete syntax and unknown prerequisite facts. An extension function name alone does not prove an application function is that builtin.
- E2. A dominating valid guard or repair described in Detection prevents the finding. Both legacy noinspection and custos-ignore forms suppress this ID.

## Report

- Range: the violating expression or call described in D1; stateful absence checks highlight the final unsafe operation or loop.
- Severity: warning.
- Enabled by default: true.
- Message: `Convert timezone offsets to seconds.`

## Fix

For the exact proven offset expression concatenated with a seconds literal, wrap only the numeric expression in (original / 1000); preserve its single evaluation.

All edits use exact byte ranges and preserve comments and evaluation order. Offer no fix when its stated prerequisites cannot be proven or the replacement is unsupported by the configured PHP version.

## Options

None.

## PHP versions

Requires PHP 5.5 or later and the referenced extension API. Respect method-specific introduction versions within PHP 5.3–8.5; unsupported optional APIs cannot establish facts. Named arguments require PHP 8.0 or later.

## Examples

```php
<?php
$z=IntlTimeZone::createTimeZone("Europe/Paris"); $z->getOffset($instant,false,$raw,$dst); $d=new DateTime(); <warning descr="Convert timezone offsets to seconds.">$d->modify(($raw+$dst)." seconds")</warning>;
```

Fixed result:

```php
<?php
$z=IntlTimeZone::createTimeZone("Europe/Paris"); $z->getOffset($instant,false,$raw,$dst); $d=new DateTime(); $d->modify((($raw+$dst) / 1000)." seconds");
```

Valid case:

```php
<?php
$z=IntlTimeZone::createTimeZone("Europe/Paris"); $z->getOffset($instant,false,$raw,$dst); $d=new DateTime(); $d->modify((($raw+$dst)/1000)." seconds");
```

## Divergences

Native custos inspection, independently specified. No EA counterpart or EA fixture conformance applies.
