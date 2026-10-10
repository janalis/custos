---
id: AmbiguousReplacementBackreference
group: Probable bugs
kind: syntax
needs: [names]
php: { min: "", max: "" }
---

# AmbiguousReplacementBackreference

## Summary

Detect preg_replace with a literal replacement containing a numeric backreference followed immediately by a digit where the parsed larger capture number does not exist but the shorter one does.

## Detection

- D1. Report preg_replace with a literal replacement containing a numeric backreference followed immediately by a digit where the parsed larger capture number does not exist but the shorter one does.
- D2. Resolve builtin calls and named arguments before checking the contract. Project-local callees are considered only when their bodies or immutable summaries establish the required facts. The default is disabled; enable only when the stated intent matches the application.
- D3. For flow-dependent conditions, require a reachable source-to-use path with no intervening invalidation. Unknown calls, ambiguous aliases, recursion or analysis-budget exhaustion cannot establish definite facts.

## Exceptions (no report)

- E1. Existing braced reference, actual larger capture groups, escaped dollar signs and unknown patterns are excluded.
- E2. Unresolved targets, incomplete required facts, suppressed findings and PHP versions lacking the referenced API or language feature produce no report. User declarations shadowing builtin names are not builtin contracts.

## Report

- Range: the complete violating call or expression identified by D1; for a declaration-state violation, the offending property access or assignment. Highlight the final unsafe use for source-to-use findings.
- Severity: warning.
- Message: `Brace the replacement backreference before a literal digit.`

## Fix

Rewrite the reference as ${1}1 when the pattern proves only capture 1 exists; preserve replacement quoting and escaping.

A fix is offered only when the concrete prerequisites above are proven. Edits use exact source byte ranges, preserve comments and argument evaluation order, and introduce syntax supported by the configured PHP version. Never invent error-handling policy.

## Options

None. Enable or disable this rule through normal rule configuration.

## PHP versions

Apply only on PHP versions supporting the referenced APIs and syntax (within PHP 5.3–8.5). Builtin behavior and argument contracts follow the configured PHP version. Extension-specific calls require resolved extension API symbols; do not treat same-named application functions as extension calls.

## Examples

```php
<?php
$result=<warning descr="Brace the replacement backreference before a literal digit.">preg_replace('/(r)/','$11','r')</warning>;
```

```php
<?php
$result=preg_replace('/(r)/','${1}1','r');
```

## Divergences

Native custos rule. No upstream inspection or fixture comparison applies. These conditions are independently specified and examples are original.

Audit constraint: PCRE quoted spans, control escapes, nested bracket-class syntax and unsupported delimiter forms remain unknown. Their parentheses cannot establish capture numbers or authorize a replacement rewrite.
