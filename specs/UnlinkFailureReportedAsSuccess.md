---
id: UnlinkFailureReportedAsSuccess
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "", max: "" }
---

# UnlinkFailureReportedAsSuccess

## Summary

Identify discarded unlink call immediately followed by unconditional return true in the same function.

## Detection

- D1. Report discarded unlink call immediately followed by unconditional return true in the same function.
- D2. Resolve builtin functions, classes, constants and methods, including imports and supported named arguments. Follow only bounded local facts within one lexical scope. Receiver and argument provenance must establish the specific API contract; variable names and comments never prove intent.
- D3. Report one finding per violating operation. A reachable path, unchanged value identity and applicable API behavior are required. Intervening mutation, unknown calls involving tracked values, escaped aliases or exhausted analysis budgets invalidate proof.

## Exceptions (no report)

- E1. Exclude unresolved required symbols, user-defined lookalikes, incomplete syntax, unknown required values and unsupported versions. Cases outside D1 receive no finding.
- E2. Both `@custos-ignore UnlinkFailureReportedAsSuccess` and `@noinspection UnlinkFailureReportedAsSuccess` suppress this native rule.
- E3. A file the function made itself with `tempnam()` is not reported: when
  the path given to `unlink` holds, as its only local value, the result of a
  `tempnam()` call, the deletion cleans up after the function's own work,
  and the returned success is that of this work, not of the deletion. Once
  the variable is reassigned to another path, the deletion is reported again.

## Report

- Range: the complete violating call or expression. Follow D1 if it specifies a narrower range.
- Severity: warning.
- Enabled by default: true.
- Message: `Return the deletion result or handle failure.`

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
function removeFile($p) { unlink($p); return true; }
```

Valid cases:

```php
<?php
function removeFile($p) { return unlink($p); }

function canWrite(string $dir) {
    $probe = tempnam($dir, 'probe');
    if (false === $probe) {
        return false;
    }
    unlink($probe);
    return true;
}
```

## Divergences

Native custos inspection, independently specified; no EA counterpart or EA conformance requirement. Conservative proof intentionally excludes unknown runtime state and intent.
