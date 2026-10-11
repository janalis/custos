---
id: PathContainmentPrefixBoundary
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "", max: "" }
---

# PathContainmentPrefixBoundary

## Summary

Detect canonical-path containment accepted through str_starts_with or equivalent prefix comparison against a literal non-root directory prefix without a directory-separator boundary.

## Detection

- D1. Report canonical-path containment accepted through str_starts_with or equivalent prefix comparison against a literal non-root directory prefix without a directory-separator boundary. The tested path is a `realpath()` result; the prefix is a single known string that is not empty and does not end in `/` or `\`.
- D1a. The equivalent prefix comparisons, with operands in either order and optional parentheses around the call, are:
  - `strpos($path, PREFIX) === 0` or `!== 0`, with no offset argument or an offset known to be 0;
  - `strncmp($path, PREFIX, N)` (strings in either order) compared with integer 0 using `===`, `!==`, `==` or `!=`;
  - `substr($path, 0, N)` compared with PREFIX using `===`, `!==`, `==` or `!=`.

  N must be proven equal to the byte length of PREFIX: a known integer of that value, or `strlen()` of a known string identical to PREFIX. A negated comparison is reported as well, because it expresses the same boundary-less test on the rejection path.
- D2. Resolve builtin calls and named arguments before checking the contract. Project-local callees are considered only when their bodies or immutable summaries establish the required facts. The default is enabled, with positive proof required.
- D3. For flow-dependent conditions, require a reachable source-to-use path with no intervening invalidation. Unknown calls, ambiguous aliases, recursion or analysis-budget exhaustion cannot establish definite facts.

## Exceptions (no report)

- E1. Root prefixes, appended platform separators, exact-root plus child checks, unknown canonicalization and mere textual prefix use are excluded.
- E2. Unresolved targets, incomplete required facts, suppressed findings and PHP versions lacking the referenced API or language feature produce no report. User declarations shadowing builtin names are not builtin contracts.
- E3. Comparisons that are not an exact prefix test are left alone: `strpos(...) == 0` (it also accepts a missing prefix, a different defect), a non-zero or unknown offset, a comparison with anything other than 0 or the prefix, a length that is missing, unknown or different from the prefix length, and a bare or `!`-negated `strncmp` result.

## Report

- Range: the complete violating call or expression identified by D1 (the `str_starts_with` call, or the whole comparison for the D1a forms); for a declaration-state violation, the offending property access or assignment. Highlight the final unsafe use for source-to-use findings.
- Severity: warning.
- Message: `Require a directory boundary in the containment check.`

## Fix

No automatic fix: root equality and platform/symlink policy must be accounted for explicitly.

A fix is offered only when the concrete prerequisites above are proven. Edits use exact source byte ranges, preserve comments and argument evaluation order, and introduce syntax supported by the configured PHP version. Never invent error-handling policy.

## Options

None. Enable or disable this rule through normal rule configuration.

## PHP versions

Apply only on PHP versions supporting the referenced APIs and syntax (within PHP 5.3–8.5). Builtin behavior and argument contracts follow the configured PHP version. Extension-specific calls require resolved extension API symbols; do not treat same-named application functions as extension calls.

## Examples

```php
<?php
function bad($path) { $resolved = realpath($path); if ($resolved === false) { return false; } if (<warning descr="Require a directory boundary in the containment check.">str_starts_with($resolved, '/srv/files')</warning>) { return true; } return false; }
function alsoBad($path) { $resolved = realpath($path); return <warning descr="Require a directory boundary in the containment check.">strncmp($resolved, '/srv/files', 10) === 0</warning>; }
function fine($path) { $resolved = realpath($path); return $resolved === '/srv/files' || strpos($resolved, '/srv/files/') === 0; }
```

## Divergences

Native custos rule. No upstream inspection or fixture comparison applies. These conditions are independently specified and examples are original.
