---
id: CollatorComparisonTruthinessReversed
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "5.3", max: "" }
---

# CollatorComparisonTruthinessReversed

## Summary

Compare collation results explicitly.

## Detection

This is an opt-in policy. Truthiness, previous-state checks, raw wait-status comparisons and dispatch acknowledgements can be intentional; enabling this inspection requests explicit handling of the contract described below.

- D1. Report direct truthiness, negation, assignment-in-condition, or equality against true of resolved Collator::compare. Highlight the complete comparison when the misuse is equality against true; otherwise highlight the call. Equality is zero; failure is false. Desired ordering is not inferred.
- D2. Resolve builtin functions, extension classes, method receivers, constants and named arguments. Follow unchanged locals and aliases only within the same bounded lexical scope. Require a reachable path; merge branch facts only when all incoming paths establish them.
- D3. Emit one finding at the violating operation, not one per evidence statement. Rebinding, writes, by-reference escape, unresolved calls that can mutate relevant state, and exhausted analysis budgets remove proof.

## Exceptions (no report)

- E1. Exclude shadowed builtins, unresolved receivers, incomplete syntax and unknown prerequisite facts. An extension function name alone does not prove an application function is that builtin.
- E2. A dominating valid guard or repair described in Detection prevents the finding. Both legacy noinspection and custos-ignore forms suppress this ID.

## Report

- Range: the violating expression or call described in D1; stateful absence checks highlight the final unsafe operation or loop.
- Severity: warning.
- Enabled by default: false.
- Message: `Compare collation results explicitly.`

## Fix

No automatic fix. Neither truthiness nor comparison against true establishes whether equality, ascending order or descending order was intended.

## Options

None.

## PHP versions

Requires PHP 5.3 or later and the referenced extension API. Respect method-specific introduction versions within PHP 5.3–8.5; unsupported optional APIs cannot establish facts. Named arguments require PHP 8.0 or later.

## Examples

```php
<?php
$c=new Collator("en_US"); if (<warning descr="Compare collation results explicitly.">$c->compare("cobalt","cobalt") === true</warning>) { echo "equal"; }
```

Valid case:

```php
<?php
$c=new Collator("en_US"); if ($c->compare("cobalt","cobalt") === 0) { echo "equal"; }
```

## Divergences

Native custos inspection, independently specified. No EA counterpart or EA fixture conformance applies.
