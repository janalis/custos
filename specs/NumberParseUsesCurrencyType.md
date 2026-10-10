---
id: NumberParseUsesCurrencyType
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "5.3", max: "" }
---

# NumberParseUsesCurrencyType

## Summary

Parse currency with parseCurrency. Enabled by default; require positive semantic or local-flow proof.

## Detection

- D1. Report resolved NumberFormatter::parse with its type argument proven NumberFormatter::TYPE_CURRENCY. Do not infer misuse merely from a CURRENCY formatter style.
- D2. Resolve builtin functions, extension classes, method receivers, constants and named arguments. Follow unchanged locals and aliases only within the same bounded lexical scope. Require a reachable path; merge branch facts only when all incoming paths establish them.
- D3. Emit one finding at the violating operation, not one per evidence statement. Rebinding, writes, by-reference escape, unresolved calls that can mutate relevant state, and exhausted analysis budgets remove proof.

## Exceptions (no report)

- E1. Exclude shadowed builtins, unresolved receivers, incomplete syntax and unknown prerequisite facts. An extension function name alone does not prove an application function is that builtin.
- E2. A dominating valid guard or repair described in Detection prevents the finding. Both legacy noinspection and custos-ignore forms suppress this ID.

## Report

- Range: the violating expression or call described in D1; stateful absence checks highlight the final unsafe operation or loop.
- Severity: error.
- Enabled by default: true.
- Message: `Parse currency with parseCurrency.`

## Fix

No automatic fix. The required repair depends on error handling, data selection, lifecycle ordering or application policy.

All edits use exact byte ranges and preserve comments and evaluation order. Offer no fix when its stated prerequisites cannot be proven or the replacement is unsupported by the configured PHP version.

## Options

None.

## PHP versions

Requires PHP 5.3 or later and the referenced extension API. Respect method-specific introduction versions within PHP 5.3–8.5; unsupported optional APIs cannot establish facts. Named arguments require PHP 8.0 or later.

## Examples

```php
<?php
$f = new NumberFormatter("en_US", NumberFormatter::CURRENCY); <error descr="Parse currency with parseCurrency.">$f->parse("$37.20", NumberFormatter::TYPE_CURRENCY)</error>;
```

Valid case:

```php
<?php
$f = new NumberFormatter("en_US", NumberFormatter::CURRENCY); $f->parseCurrency("$37.20", $currency);
```

## Divergences

Native custos inspection, independently specified. No EA counterpart or EA fixture conformance applies.
