---
id: UsortDiscardsRequiredKeys
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "", max: "" }
---

# UsortDiscardsRequiredKeys

## Summary

Detect usort on a statically keyed array followed by access through an original nonsequential integer or string key that cannot exist after reindexing.

## Detection

- D1. Report usort on a statically keyed array followed by access through an original nonsequential integer or string key that cannot exist after reindexing.
- D2. Resolve builtin calls and named arguments before checking the contract. Project-local callees are considered only when their bodies or immutable summaries establish the required facts. The default is disabled; enable only when the stated intent matches the application.
- D3. For flow-dependent conditions, require a reachable source-to-use path with no intervening invalidation. Unknown calls, ambiguous aliases, recursion or analysis-budget exhaustion cannot establish definite facts.

## Exceptions (no report)

- E1. Unknown keys, sequential lists, reassignment, and escaped arrays invalidate the proof. Assignment-only targets, unset, isset, empty, and null-coalescing fallbacks do not establish that the original key must survive.
- E2. Unresolved targets, incomplete required facts, suppressed findings and PHP versions lacking the referenced API or language feature produce no report. User declarations shadowing builtin names are not builtin contracts.
- E3. A null-coalescing assignment (`??=`) to the key is a fallback, not a read. Other compound assignments read the key and stay reportable.
- E4. A read is not proof when the inspected statement also tests the same key of the same array for presence: isset, empty, the left side of `??` or `??=`, or builtin array_key_exists / key_exists with a statically known key. The test may guard the read (if body, ternary branch, right side of `&&`), so the code already tolerates a missing key. Tests on another key, another array or an unknown key guard nothing.
- E5. A read is not proof when an earlier part of the inspected statement assigns or unsets that key, or assigns the whole array: the value read is no longer the pre-sort entry. A read positioned before the write, including the right side of the assignment to the same key, remains proof.

## Report

- Range: the complete violating call or expression identified by D1; for a declaration-state violation, the offending property access or assignment. Highlight the final unsafe use for source-to-use findings.
- Severity: warning.
- Message: `Preserve record keys with uasort.`

## Fix

Replace the callee with fully qualified \uasort when the later read proves preserved keys are required; preserve arguments. Qualifying the builtin prevents a local function or imported alias from redirecting the corrected call. Assignment targets alone do not establish a read dependency. Because uasort turns the sorted list back into a keyed map, no fix (and no report) is produced when the only uses of the key are presence-tested, defaulted or overwritten ones (E3-E5).

A fix is offered only when the concrete prerequisites above are proven. Edits use exact source byte ranges, preserve comments and argument evaluation order, and introduce syntax supported by the configured PHP version. Never invent error-handling policy.

## Options

None. Enable or disable this rule through normal rule configuration.

## PHP versions

Apply only on PHP versions supporting the referenced APIs and syntax (within PHP 5.3–8.5). Builtin behavior and argument contracts follow the configured PHP version. Extension-specific calls require resolved extension API symbols; do not treat same-named application functions as extension calls.

## Examples

```php
<?php
$users = [31 => 'Mia', 52 => 'Noor']; <warning descr="Preserve record keys with uasort.">usort($users, fn($a, $b) => $a <=> $b)</warning>; echo $users[31];
```

```php
<?php
$users = [31 => 'Mia', 52 => 'Noor']; \uasort($users, fn($a, $b) => $a <=> $b); echo $users[31];
```

## Divergences

Native custos rule. No upstream inspection or fixture comparison applies. These conditions are independently specified and examples are original.
