---
id: ShortStreamReadUnchecked
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "", max: "" }
---

# ShortStreamReadUnchecked

## Summary

Detect fread with a fixed positive byte count whose result reaches unpack requiring that exact length without a length check.

## Detection

A sentinel or length guard applies only to the same reaching value. Reassignment or reevaluating an identical-looking call invalidates the earlier guard.

- D1. Report fread with a fixed positive byte count whose result reaches unpack requiring that exact length without a length check.
- D2. Resolve builtin calls and named arguments before checking the contract. Project-local callees are considered only when their bodies or immutable summaries establish the required facts. The default is enabled, with positive proof required.
- D3. For flow-dependent conditions, require a reachable source-to-use path with no intervening invalidation. Unknown calls, ambiguous aliases, recursion or analysis-budget exhaustion cannot establish definite facts.
- D4. A length check is a condition known to hold on every path to the unpack call and applied to the same reaching read result. Because fread never returns more than the requested count N, each of these proves a complete read:
  - `strlen($d)` compared with a known integer so that the length is N: `=== N`, `== N`, `>= N`, `> N-1`, or the negation of `!== N`, `!= N`, `< N`, `<= N-1`; operands may be written in either order. `mb_strlen($d, '8bit')` counts bytes and is treated like strlen; mb_strlen with any other or no encoding is not.
  - `isset($d[N-1])` with a single operand.
  - For a one-byte read only: a non-zero length, or the value being proven both non-empty and not `false` (strict comparisons against `''` and `false` together, one loose comparison against either, a truthiness test, or `!empty($d)`).
- D5. A condition holds on the path when the unpack call is inside the true branch of an `if`/`elseif`, inside an `else` after the negated conditions, in the matching arm of a ternary, on the right of `&&`/`||`, or after an earlier sibling `if` without `else`/`elseif` whose body always leaves (return, throw, exit, break, continue). `!` negates; a true `&&` and a false `||` establish both operands; a false `&&` or a true `||` establishes nothing.
- D6. A check of a different value, of a wrong bound (for example `strlen($d) < 3` before a four-byte decode), of character count, or one whose guarding branch does not leave, is not a length check: the report stays.

## Exceptions (no report)

- E1. Accumulation loops, length checks as defined by D4 and D5, bounded formats accepting shorter inputs, and unknown formats are excluded.
- E2. Unresolved targets, incomplete required facts, suppressed findings and PHP versions lacking the referenced API or language feature produce no report. User declarations shadowing builtin names are not builtin contracts.

## Report

- Range: the complete violating call or expression identified by D1; for a declaration-state violation, the offending property access or assignment. Highlight the final unsafe use for source-to-use findings.
- Severity: warning.
- Message: `Read the complete record before decoding it.`

## Fix

No automatic fix: EOF, timeouts and retry behavior require explicit handling.

## Options

None. Enable or disable this rule through normal rule configuration.

## PHP versions

Apply only on PHP versions supporting the referenced APIs and syntax (within PHP 5.3–8.5). Builtin behavior and argument contracts follow the configured PHP version. Extension-specific calls require resolved extension API symbols; do not treat same-named application functions as extension calls.

## Examples

```php
<?php
$header = fread($stream, 4); <warning descr="Read the complete record before decoding it.">unpack('Nsize', $header)</warning>;
```

Not reported, the short read is rejected first:

```php
<?php
$header = fread($stream, 4);
if ($header === false || strlen($header) < 4) {
    throw new RuntimeException('truncated header');
}
unpack('Nsize', $header);
```

## Divergences

Native custos rule. No upstream inspection or fixture comparison applies. These conditions are independently specified and examples are original.
