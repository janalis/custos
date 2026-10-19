---
id: ReadonlyPropertyReassignment
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "8.1", max: "" }
---

# ReadonlyPropertyReassignment

## Summary

Detect assignment to a resolved readonly instance property already initialized on every reachable incoming path.

## Detection

- D1. Report assignment to a resolved readonly instance property already initialized on every reachable incoming path.
- D1a. Every writing assignment operator counts as a repeated write: plain `=` and the compound operators (`+=`, `-=`, `.=`, ...) all fail on an initialized readonly property. The null-coalescing assignment `??=` is the exception, see E3.
- D2. Resolve builtin calls and named arguments before checking the contract. Project-local callees are considered only when their bodies or immutable summaries establish the required facts. The default is enabled, with positive proof required.
- D3. For flow-dependent conditions, require a reachable source-to-use path with no intervening invalidation. Unknown calls, ambiguous aliases, recursion or analysis-budget exhaustion cannot establish definite facts.

## Exceptions (no report)

- E1. A first permitted initialization, version-permitted clone reinitialization, unknown initialization state and reflection operations are excluded.
- E2. Unresolved targets, incomplete required facts, suppressed findings and PHP versions lacking the referenced API or language feature produce no report. User declarations shadowing builtin names are not builtin contracts.

- E3. A `??=` assignment is never reported. It writes only when the stored value is null, and the adjacent-assignment proof establishes that the property is initialized, not that it holds null; on a non-null value the statement is a no-op and PHP raises nothing.

## Report

- Range: the complete violating call or expression identified by D1; for a declaration-state violation, the offending property access or assignment. Highlight the final unsafe use for source-to-use findings.
- Severity: error.
- Message: `Initialize this readonly property only once.`

## Fix

No automatic fix: removing an assignment may conceal a missing design change.

## Options

None. Enable or disable this rule through normal rule configuration.

## PHP versions

Available from PHP 8.1. Builtin behavior and argument contracts follow the configured PHP version. Extension-specific calls require resolved extension API symbols; do not treat same-named application functions as extension calls.

## Examples

```php
<?php
class Item { public readonly int $id; public function __construct() {$this->id=1; <error descr="Initialize this readonly property only once.">$this->id</error>=2;} }
```

## Divergences

Native custos rule. No upstream inspection or fixture comparison applies. These conditions are independently specified and examples are original.

Audit constraint: definite repeated initialization is currently proven only by adjacent assignments through the same direct variable receiver, with a literal or variable first value. Receiver replacement, unknown evaluation and intervening statements invalidate the proof. A `??=` write is conditional on a null value that this proof does not establish, so it is left unreported.
