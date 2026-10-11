---
id: ByteLengthUsedAsCharacterCount
group: Probable bugs
kind: syntax
needs: [names]
php: { min: "", max: "" }
---

# ByteLengthUsedAsCharacterCount

## Summary

Detect strlen of a literal non-ASCII valid UTF-8 string under an explicitly enabled Unicode character-count contract.

## Detection

- D1. Report strlen of a literal non-ASCII valid UTF-8 string under an explicitly enabled Unicode character-count contract.
- D2. Resolve builtin calls and named arguments before checking the contract. Project-local callees are considered only when their bodies or immutable summaries establish the required facts. The default is disabled; enable only when the stated intent matches the application.
- D3. For flow-dependent conditions, require a reachable source-to-use path with no intervening invalidation. Unknown calls, ambiguous aliases, recursion or analysis-budget exhaustion cannot establish definite facts.

## Exceptions (no report)

- E1. ASCII literals and operands of unknown encoding (anything that is not a known literal string, or a literal that is not valid UTF-8) are excluded.
- E3. Byte-length use in protocol/storage operations is excluded: no report when the `strlen` result is passed as an argument to a resolved builtin whose offsets, lengths or sizes are counted in bytes. The recognised consumers are `substr`, `substr_replace`, `substr_count`, `substr_compare`, `mb_strcut`, `strpos`, `stripos`, `strrpos`, `strripos`, `strncmp`, `strncasecmp`, `str_pad`, `pack`, `fread`, `fwrite`, `fputs`, `fgets`, `fseek`, `ftruncate`, `stream_get_contents`, `stream_copy_to_stream`, `file_get_contents`, `socket_read`, `socket_write`, `socket_send` and `socket_recv`. The result may reach the argument through parentheses, unary minus and `+`, `-` or `*` arithmetic (`substr($s, -strlen('é'))`, `fseek($h, strlen('é') + 4)`). Any other route (a variable, concatenation, comparison, a user function or a same-named namespaced function) keeps the report.
- E2. Unresolved targets, incomplete required facts, suppressed findings and PHP versions lacking the referenced API or language feature produce no report. User declarations shadowing builtin names are not builtin contracts.

## Report

- Range: the complete violating call or expression identified by D1; for a declaration-state violation, the offending property access or assignment. Highlight the final unsafe use for source-to-use findings.
- Severity: warning.
- Message: `Count Unicode characters with the intended counting model.`

## Fix

No automatic fix: code-point versus grapheme counting and encoding must be selected explicitly.

A fix is offered only when the concrete prerequisites above are proven. Edits use exact source byte ranges, preserve comments and argument evaluation order, and introduce syntax supported by the configured PHP version. Never invent error-handling policy.

## Options

None. Enable or disable this rule through normal rule configuration.

## PHP versions

Apply only on PHP versions supporting the referenced APIs and syntax (within PHP 5.3–8.5). Builtin behavior and argument contracts follow the configured PHP version. Extension-specific calls require resolved extension API symbols; do not treat same-named application functions as extension calls.

## Examples

```php
<?php
$count=<warning descr="Count Unicode characters with the intended counting model.">strlen('ñ')</warning>;
```

Not reported: the length is consumed as a byte quantity.

```php
<?php
$amount = substr($price, strlen('€'));
$frame = pack('N', strlen('héllo')) . 'héllo';
```

## Divergences

Native custos rule. No upstream inspection or fixture comparison applies. These conditions are independently specified and examples are original.
