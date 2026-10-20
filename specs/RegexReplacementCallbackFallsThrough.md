---
id: RegexReplacementCallbackFallsThrough
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "", max: "" }
---

# RegexReplacementCallbackFallsThrough

## Summary

Identify preg_replace_callback with an inline callback having a reachable fallthrough path and no generator semantics. An explicit return on every path excludes.

## Detection

- D1. Report preg_replace_callback with an inline callback having a reachable fallthrough path and no generator semantics. An explicit return on every path excludes.
- D1a. A path ends, without falling through, at `return`, `throw`, `exit`/`die`, or a statement whose expression has the resolved type `never` (a call to a function, static method or instance method declared `never`). An `if` with a literal condition contributes only its selected branch; any other `if` falls through when it has no `else` or when at least one of its branches does.
- D2. Resolve builtin functions, classes, constants and methods, including imports and supported named arguments. Follow only bounded local facts within one lexical scope. Receiver and argument provenance must establish the specific API contract; variable names and comments never prove intent.
- D3. Report one finding per violating operation. A reachable path, unchanged value identity and applicable API behavior are required. Intervening mutation, unknown calls involving tracked values, escaped aliases or exhausted analysis budgets invalidate proof.

## Exceptions (no report)

- E1. Exclude unresolved required symbols, user-defined lookalikes, incomplete syntax, unknown required values and unsupported versions. Cases outside D1 receive no finding.
- E1a. A statement-level call whose target cannot be resolved ends the path as well: a function that is neither a builtin nor a known declaration, a dynamic function call, or a method call on a receiver of unknown type. Such a callee may be declared `never`, so fallthrough is unproven and nothing is reported.
- E1b. Callbacks containing loops, `try`, `switch` or `goto` are not analyzed.
- E2. Both `@custos-ignore RegexReplacementCallbackFallsThrough` and `@noinspection RegexReplacementCallbackFallsThrough` suppress this native rule.

## Report

- Range: the complete violating call or expression. Follow D1 if it specifies a narrower range.
- Severity: warning.
- Enabled by default: true.
- Message: `Return the replacement from the callback.`

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
preg_replace_callback('~[a-z]+~', function($m) { strtoupper($m[0]); }, $text);
```

Valid case:

```php
<?php
preg_replace_callback('~[a-z]+~', function($m) { return strtoupper($m[0]); }, $text);
```

Valid case, the remaining path ends in a function that never returns:

```php
<?php
function reject(string $why): never { throw new DomainException($why); }
preg_replace_callback('~[a-z]~', function ($m) {
    if ($m[0] !== 'q') {
        return strtoupper($m[0]);
    }
    reject('unsupported letter');
}, $text);
```

## Divergences

Native custos inspection, independently specified; no EA counterpart or EA conformance requirement. Conservative proof intentionally excludes unknown runtime state and intent.
