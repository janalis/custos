---
id: PregSplitCaptureFlagWithoutGroups
group: Probable bugs
kind: syntax
needs: [names]
php: { min: "", max: "" }
---

# PregSplitCaptureFlagWithoutGroups

## Summary

Detect preg_split with PREG_SPLIT_DELIM_CAPTURE on a parsed literal regex containing no capturing groups.

## Detection

- D1. Report preg_split with PREG_SPLIT_DELIM_CAPTURE on a parsed literal regex containing no capturing groups.
- D2. Resolve builtin calls and named arguments before checking the contract. Project-local callees are considered only when their bodies or immutable summaries establish the required facts. The default is enabled, with positive proof required.
- D3. For flow-dependent conditions, require a reachable source-to-use path with no intervening invalidation. Unknown calls, ambiguous aliases, recursion or analysis-budget exhaustion cannot establish definite facts.

## Exceptions (no report)

- E1. Named or numbered captures, unknown regex/flags, and branch-reset captures are excluded.
- E2. Unresolved targets, incomplete required facts, suppressed findings and PHP versions lacking the referenced API or language feature produce no report. User declarations shadowing builtin names are not builtin contracts.

## Report

- Range: the complete violating call or expression identified by D1; for a declaration-state violation, the offending property access or assignment. Highlight the final unsafe use for source-to-use findings.
- Severity: info.
- Message: `Remove the capture flag or add a capturing group.`

## Fix

Remove the ineffective flag only when flags are a sole constant or a side-effect-free bitwise-or tree; use zero for an otherwise empty flags argument.

A fix is offered only when the concrete prerequisites above are proven. Edits use exact source byte ranges, preserve comments and argument evaluation order, and introduce syntax supported by the configured PHP version. Never invent error-handling policy.

## Options

None. Enable or disable this rule through normal rule configuration.

## PHP versions

Apply only on PHP versions supporting the referenced APIs and syntax (within PHP 5.3–8.5). Builtin behavior and argument contracts follow the configured PHP version. Extension-specific calls require resolved extension API symbols; do not treat same-named application functions as extension calls.

## Examples

```php
<?php
$parts=<info descr="Remove the capture flag or add a capturing group.">preg_split('/,/',$text,-1,PREG_SPLIT_DELIM_CAPTURE)</info>;
```

```php
<?php
$parts=preg_split('/,/',$text,-1,0);
```

## Divergences

Native custos rule. No upstream inspection or fixture comparison applies. These conditions are independently specified and examples are original.

Audit constraint: PHP hash comments inside the flag expression withhold the fix. Unsupported PCRE quoted spans, control escapes and bracket-class syntax cannot prove the absence of captures.
