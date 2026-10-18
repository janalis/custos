---
id: CurlResponseZeroRejected
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "", max: "" }
---

# CurlResponseZeroRejected

## Summary

Identify curl_exec result used in a truthiness-based failure branch when the same handle has RETURNTRANSFER proven true, and failure branch throws or returns false. Unknown response still merits this contract warning.

## Detection

- D1. Report curl_exec result used in a truthiness-based failure branch when the same handle has RETURNTRANSFER proven true, and failure branch throws or returns false. Unknown response still merits this contract warning.
- D2. Resolve builtin functions, classes, constants and methods, including imports and supported named arguments. Follow only bounded local facts within one lexical scope. Receiver and argument provenance must establish the specific API contract; variable names and comments never prove intent.
- D3. Report one finding per violating operation. A reachable path, unchanged value identity and applicable API behavior are required. Intervening mutation, unknown calls involving tracked values, escaped aliases or exhausted analysis budgets invalidate proof.
- D4. The RETURNTRANSFER proof is the result of replaying, in source order, every dominating `curl_setopt` and `curl_setopt_array` write on the handle. A truthy `CURLOPT_RETURNTRANSFER` establishes it; a falsy or unknown value clears it. `CURLOPT_FILE` and `CURLOPT_WRITEFUNCTION` replace the write sink, so `curl_exec` returns a boolean again: they clear the proof until a later truthy `CURLOPT_RETURNTRANSFER`. Any other option named by a builtin constant leaves it unchanged.
- D5. The failure condition must negate the call an odd number of times, counting each `!` and each loose `== false` between the call and the `if`. Parentheses and assignments are transparent.

- D6. A failure test must select a falsy result: count negation and loose equality with literal false, on either side, by parity. Parentheses, error suppression and the right-hand value of a plain non-reference assignment preserve the test. Compound assignments and assignment targets do not.
- D7. Require every supported path through the failure body to throw or return literal false. A conditional rejection alone is insufficient. Structured blocks and conditional branches are bounded; suspension, jumps, unsupported control flow and budget exhaustion establish no rejection proof.

## Exceptions (no report)

- E1. Exclude unresolved required symbols, user-defined lookalikes, incomplete syntax, unknown required values and unsupported versions. Cases outside D1 receive no finding.
- E3. No report when a write on the handle cannot be read: a `curl_setopt` whose option is not a builtin constant (variable, class constant, user constant, number), or a `curl_setopt_array` whose options are not a literal array with builtin constant keys. Such a write may disable RETURNTRANSFER, so it clears the proof like D4's sink options.
- E4. No report for an even number of negations such as `!!curl_exec($ch)`; that branch runs when a body was received.
- E2. Both `@custos-ignore CurlResponseZeroRejected` and `@noinspection CurlResponseZeroRejected` suppress this native rule.

## Report

- Range: the complete violating call or expression. Follow D1 if it specifies a narrower range.
- Severity: warning.
- Enabled by default: true.
- Message: `Compare cURL transport failure strictly with false.`

## Fix

None. A repair requires a semantic decision or application error-handling policy.

## Options

None.

## PHP versions

Available in PHP 5.3–8.5 where the referenced API exists. Do not apply extension contracts when the resolved API is unavailable. Gate syntax and version-specific behavior to the configured target.

## Examples

Finding (enable the rule explicitly when disabled by default):

```php
<?php
curl_setopt($ch, CURLOPT_RETURNTRANSFER, true); if (!curl_exec($ch)) { throw new RuntimeException(); }
```

Valid case:

```php
<?php
curl_setopt($ch, CURLOPT_RETURNTRANSFER, true); if (curl_exec($ch) === false) { throw new RuntimeException(); }
```

Valid case, the body goes to a file and the call returns a boolean:

```php
<?php
curl_setopt($ch, CURLOPT_RETURNTRANSFER, true);
curl_setopt($ch, CURLOPT_FILE, $fp);
if (!curl_exec($ch)) { throw new RuntimeException(); }
```

## Divergences

Native custos inspection, independently specified; no EA counterpart or EA conformance requirement. Conservative proof intentionally excludes unknown runtime state and intent.
