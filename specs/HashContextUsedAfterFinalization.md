---
id: HashContextUsedAfterFinalization
group: Security
kind: semantic
needs: [names, types, flow]
php: { min: "", max: "" }
---

# HashContextUsedAfterFinalization

## Summary

Identify hash_update, hash_copy or hash_final on the same local context after unconditional hash_final, with no replacement.

## Detection

- D1. Report hash_update, hash_copy or hash_final on the same local context after unconditional hash_final, with no replacement.
- D2. Resolve builtin functions, classes, constants and methods, including imports and supported named arguments. Follow only bounded local facts within one lexical scope. Receiver and argument provenance must establish the specific API contract; variable names and comments never prove intent.
- D3. Report one finding per violating operation. A reachable path, unchanged value identity and applicable API behavior are required. Intervening mutation, unknown calls involving tracked values, escaped aliases or exhausted analysis budgets invalidate proof.
- D4. The earlier hash_final must run on every path to the later call. A hash_final in a `match` arm body, or in the arguments or member name of a call reached through a `?->` chain, proves nothing for calls outside that arm or chain; calls inside the same arm or chain are still checked.
- D5. Between the earlier hash_final and the later call the context variable must not be rebound. Besides plain assignment, a rebinding is: a `[...]`/`list(...)` destructuring target (positional, keyed or nested), a `foreach` key or value binding, a `catch` variable, a call to `extract` or `parse_str`, and any user function, method, static method or constructor call that receives the variable.
- D6. When the context variable is a by-reference parameter, or is declared `global` or `static` in the scope, any intervening user function, method, static method or constructor call may replace it and invalidates proof. Calls to resolved builtin functions do not.

## Exceptions (no report)

- E1. Exclude unresolved required symbols, user-defined lookalikes, incomplete syntax, unknown required values and unsupported versions. Cases outside D1 receive no finding.
- E2. No report when the only earlier hash_final is conditional in the sense of D4, or when the variable is rebound or may be replaced as described in D5 and D6.
- E3. Both `@custos-ignore HashContextUsedAfterFinalization` and `@noinspection HashContextUsedAfterFinalization` suppress this native rule.

## Report

- Range: the complete violating call or expression. Follow D1 if it specifies a narrower range.
- Severity: error.
- Enabled by default: true.
- Message: `Create a new hash context after finalization.`

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
$h = hash_init('sha256'); hash_final($h); hash_update($h, 'next');
```

Valid case:

```php
<?php
$h = hash_init('sha256'); hash_update($h, 'next'); hash_final($h);
```

Valid cases, the variable holds a fresh context or the finalization is not certain:

```php
<?php
function pair(string $a, string $b): array
{
    $h = hash_init('sha256');
    hash_update($h, $a);
    $first = hash_final($h);
    [$h, $tag] = [hash_init('sha256'), 'second'];
    hash_update($h, $b);
    return [$first, $tag, hash_final($h)];
}

function step($h, int $op, string $chunk): void
{
    match ($op) {
        1 => print(hash_final($h)),
        default => hash_update($h, $chunk),
    };
}
```

## Divergences

Native custos inspection, independently specified; no EA counterpart or EA conformance requirement. Conservative proof intentionally excludes unknown runtime state and intent.
