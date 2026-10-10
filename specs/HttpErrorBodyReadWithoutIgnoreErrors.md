---
id: HttpErrorBodyReadWithoutIgnoreErrors
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "", max: "" }
---

# HttpErrorBodyReadWithoutIgnoreErrors

## Summary

Enable reading HTTP error response bodies. This opt-in inspection applies the policy stated in Detection.

## Detection

- D1. Enabled policy requires obtaining HTTP error response bodies. Resolve file_get_contents/fopen over a proven http/https URL with a local stream_context_create HTTP context whose ignore_errors option is absent or false. Report the consuming call. The default option false also violates policy. Named arguments and literal contexts are supported.
- D2. Resolve builtin functions, classes, constants and methods, including imported aliases and supported named arguments. Follow facts within one lexical scope and use resolved declarations when available; never infer intent from identifiers or comments.
- D3. Emit one finding per violating operation. Flow facts require a reachable path with no intervening mutation, escaped alias, unknown call involving the tracked value or analysis-budget exhaustion. Uncertain facts cannot prove a violation.

## Exceptions (no report)

- E1. ignore_errors=true, non-HTTP streams, unknown URL schemes and unknown contexts are excluded. Enabling the rule states the response-body policy; comments or variable names are not evidence.
- E2. Exclude shadowed builtin symbols, unresolved required types, incomplete syntax and unsupported PHP versions. Both `@custos-ignore HttpErrorBodyReadWithoutIgnoreErrors` and `@noinspection HttpErrorBodyReadWithoutIgnoreErrors` suppress this inspection.

## Report

- Range: the complete violating call/expression unless Detection specifies an attribute, return type, header or iterable expression.
- Severity: warning.
- Enabled by default: false.
- Message: `Enable reading HTTP error response bodies.`

## Fix

No automatic fix. Choosing a repair requires runtime information, error-handling policy or a semantic decision.

All edits use exact byte ranges, preserve comments and evaluation order, and require syntax supported by the target PHP version. A fix must never add an evaluation or silently change unrelated arguments.

## Options

None. Enabling the rule declares the policy stated in Detection; the rule is disabled by default.

## PHP versions

Available throughout PHP 5.3–8.5. Apply only where the referenced language features and builtin/extension APIs exist. API-specific later syntax or behavior is gated as described in Detection.

## Examples

Finding with the rule explicitly enabled:

```php
<?php
$ctx = stream_context_create(["http" => ["ignore_errors" => false]]); <warning descr="Enable reading HTTP error response bodies.">file_get_contents("https://service.test/data", false, $ctx)</warning>;
```

Valid case:

```php
<?php
$ctx = stream_context_create(["http" => ["ignore_errors" => true]]); file_get_contents("https://service.test/data", false, $ctx);
```

## Divergences

Native custos inspection, independently specified. No EA counterpart or EA conformance requirement applies. Unknown intent and unproven state are deliberately excluded.
