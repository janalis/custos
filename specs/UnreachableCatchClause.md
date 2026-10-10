---
id: UnreachableCatchClause
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "", max: "" }
---

# UnreachableCatchClause

## Summary

An earlier broad catch can intercept every exception a later catch would handle. Put specific exception handlers before broader handlers.

## Detection

- D1. In one try statement an earlier catch type is identical to, or a proven ancestor/interface of, every type in a later catch clause. Highlight the later clause type list; union catches require all alternatives covered.
- D2. Resolve builtin functions, classes, constants and methods; user-defined lookalikes are excluded. Follow only bounded local proof in one lexical scope.
- D3. Emit one finding per violating operation, not per supporting evidence statement. Use existing analysis budgets; exhausted proof produces no report.

## Exceptions (no report)

- E1. Do not report unresolved or shadowed builtin names, unknown values, or incomplete receiver/type provenance. Unknown calls, aliases escaping, and intervening writes invalidate local proof. Similar code outside the stated detection contract is not a finding.
- E2. Both standard suppression forms suppress this native ID.

## Report

- Range: the violating expression or call, unless D1 specifies a narrower range.
- Severity: warning.
- Enabled by default: true.
- Message: Move the specific catch before the broader catch.

## Fix

None. Choosing a repair requires intent or additional runtime information.

## Options

None.

## PHP versions

Available throughout supported PHP 5.3–8.5 for resolved exception classes. Throwable exists from PHP 7.0; do not infer it on earlier targets. Union catches need PHP 7.1.

## Examples

```php
<?php
try { work(); } catch(Throwable $e) {} catch(<warning descr="Move the specific catch before the broader catch.">RuntimeException</warning> $e) {}
```

Valid case:

```php
<?php
try{work();}catch(RuntimeException $e){}catch(Throwable $e){}
```

## Divergences

Native custos inspection, independently specified; no EA counterpart or EA conformance requirement.
