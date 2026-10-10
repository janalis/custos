---
id: ParseUrlFailureDereferenced
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "", max: "" }
---

# ParseUrlFailureDereferenced

## Summary

Identify ordinary unguarded reads of dynamic parse_url results, or literal results with a supported parsing-failure proof. Restrict to array-returning mode; missing keys are another concern.

## Detection

- D1. Report ordinary unguarded reads of dynamic parse_url results, or literal results with a supported parsing-failure proof. Empty HTTP(S) authorities and ports above 65535 establish such a proof. Restrict to array-returning mode; missing keys are another concern.
- D2. Resolve builtin functions, classes, constants and methods, including imports and supported named arguments. Follow only bounded local facts within one lexical scope. Receiver and argument provenance must establish the specific API contract; variable names and comments never prove intent.
- D3. Report one finding per violating operation. A reachable path, unchanged value identity and applicable API behavior are required. Intervening mutation, unknown calls involving tracked values, escaped aliases or exhausted analysis budgets invalidate proof.

## Exceptions (no report)

- Array and list destructuring destinations create slots rather than reading them. Destructuring keys and ordinary array values still require reads.

- Literal URL shapes outside the supported failure proof receive no finding. This avoids requesting runtime failure guards for constants whose parsing outcome is unsupported.

- Guarded probes with isset, empty or null coalescing, unsets and plain nested slot creation do not read a failed or missing value. Index expressions and the fallback operand of null coalescing still require normal reads.

- E1. Exclude unresolved required symbols, user-defined lookalikes, incomplete syntax, unknown required values and unsupported versions. Cases outside D1 receive no finding.
- E2. Both `@custos-ignore ParseUrlFailureDereferenced` and `@noinspection ParseUrlFailureDereferenced` suppress this native rule.

## Report

- Range: the complete violating call or expression. Follow D1 if it specifies a narrower range.
- Severity: warning.
- Enabled by default: true.
- Message: `Reject URL parsing failure before indexing.`

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
$parts = parse_url($raw); echo $parts['host'];
```

Valid case:

```php
<?php
$parts = parse_url($raw); if ($parts === false) { throw new RuntimeException(); } echo $parts['host'];
```

## Divergences

Native custos inspection, independently specified; no EA counterpart or EA conformance requirement. Conservative proof intentionally excludes unknown runtime state and intent.
