---
id: RegexBodyDoesNotCompile
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "", max: "" }
---

# RegexBodyDoesNotCompile

## Summary

Identify a constant slash or tilde delimited regex body with an unterminated ordinary character class or group, or unmatched closing parenthesis. Validate delimiter/modifier shape first. Exclude advanced (? syntax, extended mode, quoted \Q regions, escapes not fully parsed and paired delimiters; no Go regexp parser may reject valid PCRE features.

## Detection

- D1. Report a constant slash or tilde delimited regex body with an unterminated ordinary character class or group, or unmatched closing parenthesis. Validate delimiter/modifier shape first. Exclude advanced (? syntax, extended mode, quoted \Q regions, escapes not fully parsed and paired delimiters; no Go regexp parser may reject valid PCRE features.
- D2. Resolve builtin functions, classes, constants and methods, including imports and supported named arguments. Follow only bounded local facts within one lexical scope. Receiver and argument provenance must establish the specific API contract; variable names and comments never prove intent.
- D3. Report one finding per violating operation. A reachable path, unchanged value identity and applicable API behavior are required. Intervening mutation, unknown calls involving tracked values, escaped aliases or exhausted analysis budgets invalidate proof.

## Exceptions (no report)

- E1. Exclude unresolved required symbols, user-defined lookalikes, incomplete syntax, unknown required values and unsupported versions. Cases outside D1 receive no finding.
- E2. Both `@custos-ignore RegexBodyDoesNotCompile` and `@noinspection RegexBodyDoesNotCompile` suppress this native rule.

## Report

- Range: the complete violating call or expression. Follow D1 if it specifies a narrower range.
- Severity: error.
- Enabled by default: true.
- Message: `Correct the regular expression body.`

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
preg_match('~[~', $text);
```

Valid case:

```php
<?php
preg_match('~[a-z]~', $text);
```

## Divergences

Native custos inspection, independently specified; no EA counterpart or EA conformance requirement. Conservative proof intentionally excludes unknown runtime state and intent.
