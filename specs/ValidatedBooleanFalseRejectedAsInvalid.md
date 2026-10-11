---
id: ValidatedBooleanFalseRejectedAsInvalid
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "", max: "" }
---

# ValidatedBooleanFalseRejectedAsInvalid

## Summary

Identify FILTER_VALIDATE_BOOLEAN result compared strictly with false in a branch that unconditionally throws or returns explicit failure, when JSON-like known input is a valid false spelling; FILTER_NULL_ON_FAILURE and null comparisons exclude. Unknown input alone is insufficient to prove intent.

## Detection

- D1. Report FILTER_VALIDATE_BOOLEAN result compared strictly with false in a branch that unconditionally throws or returns explicit failure, when JSON-like known input is a valid false spelling; FILTER_NULL_ON_FAILURE and null comparisons exclude. Unknown input alone is insufficient to prove intent.
- D1a. A valid false spelling is the known string that, once the bytes space, tab, line feed, carriage return and vertical tab are stripped from both ends, is empty or equals `0`, `false`, `off` or `no` ignoring ASCII case. No other byte counts as padding: a form feed or a multi-byte space such as U+00A0 or U+2003 makes the input invalid for the filter.
- D1b. The third argument, when present, must be a known integer flag set without FILTER_NULL_ON_FAILURE, FILTER_REQUIRE_ARRAY and FILTER_FORCE_ARRAY.
- D2. Resolve builtin functions, classes, constants and methods, including imports and supported named arguments. Follow only bounded local facts within one lexical scope. Receiver and argument provenance must establish the specific API contract; variable names and comments never prove intent.
- D3. Report one finding per violating operation. A reachable path, unchanged value identity and applicable API behavior are required. Intervening mutation, unknown calls involving tracked values, escaped aliases or exhausted analysis budgets invalidate proof.

## Exceptions (no report)

- E1. Exclude unresolved required symbols, user-defined lookalikes, incomplete syntax, unknown required values and unsupported versions. Cases outside D1 receive no finding.
- E1a. Input padded with anything outside the D1a byte set is not a valid boolean, so the strict false comparison correctly means failure.
- E1b. FILTER_REQUIRE_ARRAY rejects the scalar input and FILTER_FORCE_ARRAY wraps the result in an array; in both cases false no longer stands for a validated false value.
- E2. Both `@custos-ignore ValidatedBooleanFalseRejectedAsInvalid` and `@noinspection ValidatedBooleanFalseRejectedAsInvalid` suppress this native rule.

## Report

- Range: the complete violating call or expression. Follow D1 if it specifies a narrower range.
- Severity: warning.
- Enabled by default: false.
- Message: `Distinguish valid false from invalid boolean input.`

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
if (filter_var('false', FILTER_VALIDATE_BOOLEAN) === false) { throw new InvalidArgumentException(); }
```

Valid case:

```php
<?php
if (filter_var('false', FILTER_VALIDATE_BOOLEAN, FILTER_NULL_ON_FAILURE) === null) { throw new InvalidArgumentException(); }
```

Valid case (the non-breaking space is not stripped, so the input really is invalid):

```php
<?php
if (filter_var("\u{00A0}off", FILTER_VALIDATE_BOOLEAN) === false) { throw new InvalidArgumentException(); }
```

## Divergences

Native custos inspection, independently specified; no EA counterpart or EA conformance requirement. Conservative proof intentionally excludes unknown runtime state and intent.
