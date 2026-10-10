---
id: CollatorSortDiscardsRequiredKeys
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "5.3", max: "" }
---

# CollatorSortDiscardsRequiredKeys

## Summary

Preserve keys during collation sorting. Disabled by default. Explicit enablement establishes the policy described in Detection; names, comments and example labels never establish intent.

## Detection

- D1. Enabled policy requires preserving known noninteger record keys. Report Collator::sort applied to a local array proven to have at least one string key, including a known literal. Skip list arrays, unknown keys, and Collator::asort.
- D2. Resolve builtin functions, extension classes, method receivers, constants and named arguments. Follow unchanged locals and aliases only within the same bounded lexical scope. Require a reachable path; merge branch facts only when all incoming paths establish them.
- D3. Emit one finding at the violating operation, not one per evidence statement. Rebinding, writes, by-reference escape, unresolved calls that can mutate relevant state, and exhausted analysis budgets remove proof.

## Exceptions (no report)

- E1. Exclude shadowed builtins, unresolved receivers, incomplete syntax and unknown prerequisite facts. An extension function name alone does not prove an application function is that builtin.
- E2. A dominating valid guard or repair described in Detection prevents the finding. Both legacy noinspection and custos-ignore forms suppress this ID.

## Report

- Range: the violating expression or call described in D1; stateful absence checks highlight the final unsafe operation or loop.
- Severity: warning.
- Enabled by default: false.
- Message: `Preserve keys during collation sorting.`

## Fix

Replace resolved method token sort with asort; preserve arguments and receiver evaluation.

All edits use exact byte ranges and preserve comments and evaluation order. Offer no fix when its stated prerequisites cannot be proven or the replacement is unsupported by the configured PHP version.

## Options

None.

## PHP versions

Requires PHP 5.3 or later and the referenced extension API. Respect method-specific introduction versions within PHP 5.3–8.5; unsupported optional APIs cannot establish facts. Named arguments require PHP 8.0 or later.

## Examples

```php
<?php
$c=new Collator("en_US"); $a=["part-q"=>"Quartz","part-a"=>"Amber"]; <warning descr="Preserve keys during collation sorting.">$c->sort($a)</warning>;
```

Fixed result:

```php
<?php
$c=new Collator("en_US"); $a=["part-q"=>"Quartz","part-a"=>"Amber"]; $c->asort($a);
```

Valid case:

```php
<?php
$c=new Collator("en_US"); $a=["part-q"=>"Quartz","part-a"=>"Amber"]; $c->asort($a);
```

## Divergences

Native custos inspection, independently specified. No EA counterpart or EA fixture conformance applies.
