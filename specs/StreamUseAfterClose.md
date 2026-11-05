---
id: StreamUseAfterClose
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "", max: "" }
---

# StreamUseAfterClose

## Summary

Detect a stream operation requiring an open resource on an identity closed by a successful fclose on every incoming path.

## Detection

A bounded local closure may return a proven fresh stream while preserving its
identity for later caller operations. Separate invocation sites create separate
identities. Returning a captured or passed stream preserves the caller's existing
state rather than recording a new open operation. Only operation states shared
by every normal completion are carried across the call. A close performed inside
the closure establishes a new closed state only when its success is proven;
unknown or conflicting effects discard the state proof. Reference captures,
escaped callables, retained static state and ambiguous returned identities do not
establish this transfer.

- D1. Report a stream operation requiring an open resource on an identity closed by a successful fclose on every incoming path.
- D2. Resolve builtin calls and named arguments before checking the contract. Project-local callees are considered only when their bodies or immutable summaries establish the required facts. The default is enabled, with positive proof required.
- D3. For flow-dependent conditions, require a reachable source-to-use path with no intervening invalidation. Unknown calls, ambiguous aliases, recursion or analysis-budget exhaustion cannot establish definite facts.

## Exceptions (no report)

- E1. Fresh reopened identity, guarded failed close, unknown handles and escaped aliases invalidate proof.
- E3. A closure created anywhere in the same scope that binds the handle variable by reference (`use (&$h)`) invalidates proof, whether it is created before or after the close: calling it can store a newly opened stream in the variable. A by-value capture, a by-reference capture of another variable, and a capture written inside a nested function body do not.
- E4. No report where the operation runs only when the reaching `fclose` returned false, since a refused close leaves the stream open. The result may be tested directly or through an unchanged local holding it, by negation, a comparison with a boolean literal, `and`/`or`, a ternary, an `if`/`else` branch, or an early exit of the success branch. An operation reached only when the close returned true is still reported, as is one guarded by a local that was reassigned after the close.
- E2. Unresolved targets, incomplete required facts, suppressed findings and PHP versions lacking the referenced API or language feature produce no report. User declarations shadowing builtin names are not builtin contracts.
- E3. A close whose stream argument comes from an unpacked list (`fclose(...$handles)`) may close the stream or not: the use is not reported as following a close.

Named function and closed method factories preserve fresh resource identity independently at each caller. Pure parameter pass-through functions preserve the supplied resource identity and its current operation state. Reference returns, opaque mutations and uncertain producers establish no resource-return contract.

A first-class native close or callback-helper close establishes a closed stream on the branch where its exact result is true. A saved result can establish the same proof while that operation is still current. Failure, unchecked results, later operations and escaped handles establish no closed-state proof.

A fixed wrapper may project operations on borrowed stream parameters when the
same state reaches every normal completion and the wrapper cannot run unknown
code. Newly closed states require proven success. Repeated actual handle
identities, reference or variadic parameters, escaping handles and uncertain
bindings establish no operation proof.

A fixed quiet wrapper returning the unchanged result of its close can establish
closed state when that result is true. This includes supported wrapper chains
and closed method targets. Saved results retain the relation only while the
same close remains current; later operations, reassignment, escapes and unrelated
return values discard it. Scalar return coercions that change boolean identity
establish no result-dependent close contract.

## Report

- Range: the complete violating call or expression identified by D1; for a declaration-state violation, the offending property access or assignment. Highlight the final unsafe use for source-to-use findings.
- Severity: error.
- Message: `Use an open stream for this operation.`

## Fix

No automatic fix: the correct operation order or reopening policy is unknown.

## Options

None. Enable or disable this rule through normal rule configuration.

## PHP versions

Apply only on PHP versions supporting the referenced APIs and syntax (within PHP 5.3–8.5). Builtin behavior and argument contracts follow the configured PHP version. Extension-specific calls require resolved extension API symbols; do not treat same-named application functions as extension calls.

## Examples

```php
<?php
function bad($path) { $h = fopen($path, 'wb'); if ($h === false) { return; } fclose($h); <error descr="Use an open stream for this operation.">fwrite($h, 'next')</error>; }
```

```php
<?php
function retry($path) { $h = fopen($path, 'wb'); if ($h === false) { return; } if (!fclose($h)) { fwrite($h, 'still open'); } }
function swap($a, $b) { $h = fopen($a, 'wb'); if ($h === false) { return; } $next = function () use (&$h, $b) { $h = fopen($b, 'wb'); }; fclose($h); $next(); fwrite($h, 'second file'); }
```

## Divergences

Native custos rule. No upstream inspection or fixture comparison applies. These conditions are independently specified and examples are original.
