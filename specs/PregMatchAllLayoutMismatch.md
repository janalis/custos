---
id: PregMatchAllLayoutMismatch
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "", max: "" }
---

# PregMatchAllLayoutMismatch

## Summary

Regex result indices depend on whether output is grouped by capture or by match. Read indices consistent with the selected layout.

## Detection

- D1. Resolved preg_match_all uses explicit PREG_PATTERN_ORDER with one capturing group and known literal subject containing at least two matches. Subsequent access reads `result[1][0]` while `result[0][1]` is explicitly compared to that value for equality, establishing mismatched group-versus-match indexing. More generally report only provably out-of-range literal capture/match indices with proven pattern shape and match count.
- D1a. Regex proof supports ordinary ASCII patterns compatible with RE2, no modifiers, lookarounds, backreferences, or unsupported PCRE extensions. Unknown matching behavior is excluded.
- D1b. Regex proof requires an ASCII subject and an ordinary compatible PCRE subset without modifiers or whitespace escapes \s or \S. Reject patterns matching an empty string, subjects larger than 32768 bytes, and results with 1024 or more matches.
- D1d. Require an actual value read of the proven absent result dimension. Exclude writes, unset, by-reference uses and unknown argument uses, and exclude statements that also mutate the result array. The regex offset must be omitted or statically known zero; nonzero or unknown offsets are unsupported.
- D1e. All argument uses are excluded because they can involve reference passing. Also exclude increment/decrement, unset and write contexts. Any same-statement mutation or call receiving the result array discards the actual-read proof.
- D2. Resolve builtin functions, classes, constants and methods; user-defined lookalikes are excluded. Follow only bounded local proof in one lexical scope.
- D3. Emit one finding per violating operation, not per supporting evidence statement. Use existing analysis budgets; exhausted proof produces no report.

## Exceptions (no report)

- E1. Unknown pattern shape/count, complex regex constructs, intentionally grouped output, or mere `m[1][0]` access without incompatible use is excluded.
- E2. Both standard suppression forms suppress this native ID.

## Report

- Range: the violating expression or call, unless D1 specifies a narrower range.
- Severity: warning.
- Enabled by default: true.
- Message: Use indices matching the selected regex result layout.

## Fix

None. Choosing a repair requires intent or additional runtime information.

## Options

None.

## PHP versions

Available throughout supported PHP 5.3–8.5 where the referenced builtin/extension is available. Respect syntax/version-specific receiver contracts; unavailable APIs are not this rule’s concern.

See the [PHP regex result and offset contract](https://www.php.net/manual/en/function.preg-match-all.php) for the relevant builtin behavior.

## Examples

```php
<?php
preg_match_all("/(\d+)/","12 34",$m,PREG_PATTERN_ORDER);echo <warning descr="Use indices matching the selected regex result layout.">$m[2][0]</warning>;
```

Valid case:

```php
<?php
preg_match_all("/(\d+)/","12 34",$m,PREG_SET_ORDER);echo $`m[1][0]`;
```

## Divergences

Native custos inspection, independently specified; no EA counterpart or EA conformance requirement.
