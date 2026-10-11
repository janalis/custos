---
id: HashFileFailureUsedAsDigest
group: Security
kind: semantic
needs: [names, types, flow]
php: { min: "", max: "" }
---

# HashFileFailureUsedAsDigest

## Summary

Identify unchecked hash_file result consumed as file contents or as an expected digest in strict comparison; exclude result propagation and guarded false rejection.

## Detection

- D1. Report unchecked hash_file result consumed as file contents or as an expected digest in strict comparison; exclude result propagation and guarded false rejection.
- D2. Resolve builtin functions, classes, constants and methods, including imports and supported named arguments. Follow only bounded local facts within one lexical scope. Receiver and argument provenance must establish the specific API contract; variable names and comments never prove intent.
- D3. Report one finding per violating operation. A reachable path, unchanged value identity and applicable API behavior are required. Intervening mutation, unknown calls involving tracked values, escaped aliases or exhausted analysis budgets invalidate proof.
- D4. A use is guarded when false cannot reach it for the same reaching value (no reassignment or invalidation between guard and use). Recognised guards:
  - an earlier statement of any enclosing block that is an `if` without `elseif`/`else` whose body always leaves the path (return, throw, exit, break, continue) and whose condition holds whenever the value is false; unrelated statements may sit between the guard and the use;
  - an enclosing `if` body or ternary arm selected by a condition that excludes false, or the `else` body/arm when the condition holds for false;
  - the right operand of `&&`/`and` whose left operand excludes false, or of `||`/`or` whose left operand holds for false.
- D5. A condition excludes false when it is `$v !== false`, `$v != false`, plain truthiness of `$v`, or `!empty($v)`; it holds for false when it is `$v === false`, `$v == false`, `!$v` or `empty($v)` (operands in either order, parentheses ignored). `!` swaps the two. A conjunction excludes false when either operand does and holds for false only when both do; a disjunction holds for false when either operand does and excludes false only when both do. Truthiness counts because the digest sink never needs an empty string distinguished from failure to be safe: every falsy value is rejected along with false.

## Exceptions (no report)

- E1. Exclude unresolved required symbols, user-defined lookalikes, incomplete syntax, unknown required values and unsupported versions. Cases outside D1 receive no finding.
- E2. Both `@custos-ignore HashFileFailureUsedAsDigest` and `@noinspection HashFileFailureUsedAsDigest` suppress this native rule.

## Report

- Range: the complete violating call or expression. Follow D1 if it specifies a narrower range.
- Severity: warning.
- Enabled by default: true.
- Message: `Reject file hashing failure before using the digest.`

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
file_put_contents($manifest, hash_file('sha256', $path));
```

Valid case:

```php
<?php
$h = hash_file('sha256', $path); if ($h !== false) { file_put_contents($manifest, $h); }
```

Valid cases (early exit followed by unrelated work, truthiness, short-circuit):

```php
<?php
function store(string $path, string $manifest): bool
{
    $sum = hash_file('sha256', $path);
    if (!$sum || $manifest === '') {
        return false;
    }
    $label = basename($path);
    return file_put_contents($manifest, $sum) !== false;
}

function matches(string $path): bool
{
    $sum = hash_file('sha256', $path);
    return $sum !== false && $sum === 'abc';
}
```

## Divergences

Native custos inspection, independently specified; no EA counterpart or EA conformance requirement. Conservative proof intentionally excludes unknown runtime state and intent.
