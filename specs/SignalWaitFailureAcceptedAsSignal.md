---
id: SignalWaitFailureAcceptedAsSignal
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "", max: "" }
---

# SignalWaitFailureAcceptedAsSignal

## Summary

Identify truthiness test of pcntl_sigwaitinfo result followed by passing that unchanged result to a call in true branch. Failure -1 is truthy; explicit >0 or !== -1 excludes.

## Detection

- D1. Report truthiness test of pcntl_sigwaitinfo result followed by passing that unchanged result to a call in true branch. Failure -1 is truthy; explicit >0 or !== -1 excludes.
- D2. Resolve builtin functions, classes, constants and methods, including imports and supported named arguments. Follow only bounded local facts within one lexical scope. Receiver and argument provenance must establish the specific API contract; variable names and comments never prove intent.
- D3. Report one finding per violating operation. A reachable path, unchanged value identity and applicable API behavior are required. Intervening mutation, unknown calls involving tracked values, escaped aliases or exhausted analysis budgets invalidate proof.

## Exceptions (no report)

- E1. Exclude unresolved required symbols, user-defined lookalikes, incomplete syntax, unknown required values and unsupported versions. Cases outside D1 receive no finding.
- E2. Both `@custos-ignore SignalWaitFailureAcceptedAsSignal` and `@noinspection SignalWaitFailureAcceptedAsSignal` suppress this native rule.

## Report

- Range: the complete violating call or expression. Follow D1 if it specifies a narrower range.
- Severity: warning.
- Enabled by default: true.
- Message: `Reject signal-wait failure before handling the result.`

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
if ($signal = pcntl_sigwaitinfo([SIGTERM], $info)) { handleSignal($signal); }
```

Valid case:

```php
<?php
$signal = pcntl_sigwaitinfo([SIGTERM], $info); if ($signal > 0) { handleSignal($signal); }
```

## Divergences

Native custos inspection, independently specified; no EA counterpart or EA conformance requirement. Conservative proof intentionally excludes unknown runtime state and intent.
