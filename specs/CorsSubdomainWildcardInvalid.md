---
id: CorsSubdomainWildcardInvalid
group: Probable bugs
kind: flow
needs: [names, types, flow]
php: { min: "", max: "" }
---

# CorsSubdomainWildcardInvalid

## Summary

Send a concrete CORS origin.

## Detection

- D1. Report builtin header setting a literal Access-Control-Allow-Origin value containing star within a serialized origin host or another origin component. The complete value star remains valid.
- D2. Resolve builtin functions, classes, constants and methods, including imported aliases and supported named arguments. Follow facts within one lexical scope and use resolved declarations when available; never infer intent from identifiers or comments.
- D3. Emit one finding per violating operation. Flow facts require a reachable path with no intervening mutation, escaped alias or unknown call involving the tracked value. When the proof is inconclusive, nothing is reported.

## Exceptions (no report)

- E1. Unknown values and a sole star are excluded. Do not apply this rule to unrelated headers that support host patterns.
- E2. Exclude shadowed builtin symbols, unresolved required types, incomplete syntax and unsupported PHP versions. Both `@custos-ignore CorsSubdomainWildcardInvalid` and `@noinspection CorsSubdomainWildcardInvalid` suppress this inspection.

## Report

- Range: the complete violating call/expression unless Detection specifies an attribute, return type, header or iterable expression.
- Severity: warning.
- Enabled by default: true.
- Message: `Send a concrete CORS origin.`

## Fix

No automatic fix. Choosing a repair requires runtime information, error-handling policy or a semantic decision.

## Options

None.

## PHP versions

Available on every supported PHP version. Apply only where the referenced language features and builtin/extension APIs exist. API-specific later syntax or behavior is gated as described in Detection.

## Examples

Finding:

```php
<?php
<warning descr="Send a concrete CORS origin.">header("Access-Control-Allow-Origin: https://*.service.test")</warning>;
```

Valid case:

```php
<?php
header("Access-Control-Allow-Origin: *");
```

## Divergences

Native custos inspection, independently specified; no EA counterpart or EA conformance requirement. Conservative proof intentionally excludes unknown runtime state and intent.
