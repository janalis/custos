---
id: LocalizedNumberTrailingInputAccepted
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "5.3", max: "" }
---

# LocalizedNumberTrailingInputAccepted

## Summary

Reject unparsed number suffixes. This opt-in rule applies the policy described in Detection.

## Detection

- D1. When enabled, require whole-input number parsing. Report an echoed NumberFormatter::parse result for a proven ASCII numeric prefix followed by an alphabetic suffix (excluding exponent syntax), unless a surrounding true branch compares the returned offset strictly with strlen of the same input. Other locale-dependent inputs and UTF-16 offset checks remain outside this bounded detection subset.
- D2. Resolve builtin functions, extension classes, method receivers, constants and named arguments. Follow unchanged locals and aliases only within the same bounded lexical scope. Require a reachable path; merge branch facts only when all incoming paths establish them.
- D3. Emit one finding at the violating operation, not one per evidence statement. Rebinding, writes, by-reference escape, unresolved calls that can mutate relevant state, and exhausted analysis budgets remove proof.

## Exceptions (no report)

- E1. Exclude shadowed builtins, unresolved receivers, incomplete syntax and unknown prerequisite facts. An extension function name alone does not prove an application function is that builtin.
- E2. A dominating valid guard or repair described in Detection prevents the finding. Both legacy noinspection and custos-ignore forms suppress this ID.

## Report

- Range: the violating expression or call described in D1; stateful absence checks highlight the final unsafe operation or loop.
- Severity: warning.
- Enabled by default: false.
- Message: `Reject unparsed number suffixes.`

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
$f = new NumberFormatter("en_US", NumberFormatter::DECIMAL); $n = $f->parse("37tail", NumberFormatter::TYPE_DOUBLE, $end); <warning descr="Reject unparsed number suffixes.">echo $n</warning>;
```

Valid case:

```php
<?php
$f = new NumberFormatter("en_US", NumberFormatter::DECIMAL); $s="37tail"; $n=$f->parse($s,NumberFormatter::TYPE_DOUBLE,$end); if ($n !== false && $end === strlen($s)) { echo $n; }
```

## Divergences

Native custos inspection, independently specified. No EA counterpart or EA fixture conformance applies.
