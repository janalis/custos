---
id: FilePermissionFailurePassesPolicy
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "", max: "" }
---

# FilePermissionFailurePassesPolicy

## Summary

Identify unchecked fileperms result masked with a nonzero permission bitmask, compared equal to zero in a branch returning true or performing the guarded operation. Require direct result provenance.

## Detection

- D1. Report unchecked fileperms result masked with a nonzero permission bitmask, compared equal to zero in a branch returning true or performing the guarded operation. Require direct result provenance.
- D2. Resolve builtin functions, classes, constants and methods, including imports and supported named arguments. Follow only bounded local facts within one lexical scope. Receiver and argument provenance must establish the specific API contract; variable names and comments never prove intent.
- D3. Report one finding per violating operation. A reachable path, unchanged value identity and applicable API behavior are required. Intervening mutation, unknown calls involving tracked values, escaped aliases or exhausted analysis budgets invalidate proof.

## Exceptions (no report)

- E1. Exclude unresolved required symbols, user-defined lookalikes, incomplete syntax, unknown required values and unsupported versions. Cases outside D1 receive no finding.
- E2. No report when a test of the same reaching lookup result already keeps `false` away from the mask comparison. Accepted tests: truthiness (`$m`, `!$m`, `empty($m)`), `is_int`/`is_integer`/`is_long`, `is_bool`, and loose or strict comparison with `false` in either operand order, combined through `!`, `&&` and `||` when the combination still proves the result is not `false`. Accepted positions: the left operand of a short-circuit `&&`/`||` whose right operand holds the comparison, an enclosing `if`/`elseif` body, an `elseif`/`else` arm reached only after an earlier condition of the same `if` failed, and an earlier sibling `if` without `elseif`/`else` whose body always leaves the block. A test of another variable, of a value reassigned since, or one that only checks the path (`file_exists`) is not a guard.
- E3. Both `@custos-ignore FilePermissionFailurePassesPolicy` and `@noinspection FilePermissionFailurePassesPolicy` suppress this native rule.

## Report

- Range: the complete violating call or expression. Follow D1 if it specifies a narrower range.
- Severity: warning.
- Enabled by default: true.
- Message: `Reject permission lookup failure before checking bits.`

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
if ((fileperms($path) & 0002) === 0) { return true; }
```

Valid case:

```php
<?php
$m = fileperms($path); if ($m !== false && ($m & 0002) === 0) { return true; }
```

```php
<?php
function readOnly($path) {
    $m = fileperms($path);
    if (!is_int($m)) { return false; }
    if (($m & 0222) === 0) { return true; }
    return false;
}
```

## Divergences

Native custos inspection, independently specified; no EA counterpart or EA conformance requirement. Conservative proof intentionally excludes unknown runtime state and intent.
