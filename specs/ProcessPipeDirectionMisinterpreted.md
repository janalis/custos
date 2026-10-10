---
id: ProcessPipeDirectionMisinterpreted
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "", max: "" }
---

# ProcessPipeDirectionMisinterpreted

## Summary

Identify parent fwrite to proc_open pipe whose known descriptor mode is w, or parent fread/stream_get_contents on mode r. Child-relative r yields parent writable pipe; child-relative w yields parent readable pipe. Require successful process creation and exact pipe descriptor provenance.

## Detection

- D1. Report parent fwrite to proc_open pipe whose known descriptor mode is w, or parent fread/stream_get_contents on mode r. Child-relative r yields parent writable pipe; child-relative w yields parent readable pipe. Require successful process creation and exact pipe descriptor provenance.
- D2. Resolve builtin functions, classes, constants and methods, including imports and supported named arguments. Follow only bounded local facts within one lexical scope. Receiver and argument provenance must establish the specific API contract; variable names and comments never prove intent.
- D3. Report one finding per violating operation. A reachable path, unchanged value identity and applicable API behavior are required. Intervening mutation, unknown calls involving tracked values, escaped aliases or exhausted analysis budgets invalidate proof.

## Exceptions (no report)

- E1. Exclude unresolved required symbols, user-defined lookalikes, incomplete syntax, unknown required values and unsupported versions. Cases outside D1 receive no finding.
- E2. Both `@custos-ignore ProcessPipeDirectionMisinterpreted` and `@noinspection ProcessPipeDirectionMisinterpreted` suppress this native rule.

## Report

- Range: the complete violating call or expression. Follow D1 if it specifies a narrower range.
- Severity: warning.
- Enabled by default: true.
- Message: `Choose pipe modes from the child perspective.`

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
$p = proc_open(['cat'], [0 => ['pipe', 'w']], $pipes); if (is_resource($p)) { fwrite($pipes[0], 'input'); }
```

Valid case:

```php
<?php
$p = proc_open(['cat'], [0 => ['pipe', 'r']], $pipes); if (is_resource($p)) { fwrite($pipes[0], 'input'); }
```

## Divergences

Native custos inspection, independently specified; no EA counterpart or EA conformance requirement. Conservative proof intentionally excludes unknown runtime state and intent.
