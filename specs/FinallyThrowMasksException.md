---
id: FinallyThrowMasksException
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "5.5", max: "" }
---

# FinallyThrowMasksException

## Summary

An exception thrown during final cleanup can replace the exception already being propagated. Preserve the original failure when reporting cleanup errors. This inspection is disabled by default because the intended policy or use can vary.

## Detection

- D1. A try body ends with an unconditional explicit throw, and its finally block also ends with an unconditional explicit throw; neither throw is caught inside its own block. Highlight the finally throw.
- D1d. The original try exception must remain unhandled when entering finally. Exclude a try statement with a catch that handles that thrown exception, including a catch that returns normally; conservative unsupported catch relationships discard the propagating-exception proof.
- D2. Resolve builtin functions, classes, constants and methods; user-defined lookalikes are excluded. Follow only bounded local proof in one lexical scope.
- D3. Emit one finding per violating operation, not per supporting evidence statement. Use existing analysis budgets; exhausted proof produces no report.

## Exceptions (no report)

- E1. Do not report unresolved or shadowed builtin names, unknown values, or incomplete receiver/type provenance. Unknown calls, aliases escaping, and intervening writes invalidate local proof. Similar code outside the stated detection contract is not a finding.
- E2. Both standard suppression forms suppress this native ID.

## Report

- Range: the violating expression or call, unless D1 specifies a narrower range.
- Severity: warning.
- Enabled by default: false.
- Message: Preserve the original exception during cleanup.

## Fix

None. Choosing a repair requires intent or additional runtime information.

## Options

None.

## PHP versions

Requires PHP 5.5 or later.

## Examples

```php
<?php
try { throw new RuntimeException("a"); } finally { <warning descr="Preserve the original exception during cleanup.">throw new LogicException("b")</warning>; }
```

Valid case:

```php
<?php
try{work();}finally{cleanup();}
```

## Divergences

Native custos inspection, independently specified; no EA counterpart or EA conformance requirement.
