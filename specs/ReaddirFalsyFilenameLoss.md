---
id: ReaddirFalsyFilenameLoss
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "", max: "" }
---

# ReaddirFalsyFilenameLoss

## Summary

Compare the directory entry with false.

## Detection

- D1. Report a resolved readdir call whose result decides an `if`, `elseif`, `while`, `do … while` condition or the last condition expression of a `for`, when it is tested for truthiness: directly, through a plain `=` assignment in the condition, through any number of `!` negations, or through a comparison with the `true`/`false` constant that is loose on the entry itself (`readdir($h) == false`). A legal filename consisting of digit zero is otherwise treated as end-of-directory. Once the entry has been turned into a boolean by `!` or a loose comparison, further `!`, `==`, `!=`, `===`, `!==` tests of that boolean against `true`/`false` are followed too (`(!readdir($h)) === false` is still a truthiness test).
- D2. Resolve builtin functions, classes, constants and methods, including imported aliases and supported named arguments. Follow facts within one lexical scope and use resolved declarations when available; never infer intent from identifiers or comments.
- D3. Emit one finding per violating operation. Flow facts require a reachable path with no intervening mutation, escaped alias, unknown call involving the tracked value or analysis-budget exhaustion. Uncertain facts cannot prove a violation.

## Exceptions (no report)

- E1. Strict comparisons against false, merely assigned results and unrelated functions are excluded. A strict `===`/`!==` comparison applied to the entry itself exempts the whole condition, however it is then negated or compared (`!(readdir($h) === false)`, `(readdir($h) !== false) == true`).
- E3. The tested value must be the entry: compound assignments (`.=`, `??=`, …), by-reference assignments, an assignment of the already negated result (`$done = !readdir($h)`), comparisons with anything other than the `true`/`false` constants, other operators and call arguments are not reported. First-class callable syntax `readdir(...)` is not a call. Earlier comma-separated expressions of a `for` condition are not tested by the loop.
- E2. Exclude shadowed builtin symbols, unresolved required types, incomplete syntax and unsupported PHP versions. Both `@custos-ignore ReaddirFalsyFilenameLoss` and `@noinspection ReaddirFalsyFilenameLoss` suppress this inspection.

## Report

- Range: the complete violating call/expression unless Detection specifies an attribute, return type, header or iterable expression.
- Severity: warning.
- Enabled by default: true.
- Message: `Compare the directory entry with false.`

## Fix

Replace the whole condition with the entry expression, parenthesized, compared strictly: `!== false` when the original condition held for a truthy entry and `=== false` otherwise. The entry expression is the call together with its own parentheses and plain assignment, copied verbatim, so the assignment and the call are evaluated exactly once and comments inside it survive. Negations and boolean comparisons around it are folded into the chosen operator (`!!readdir($h)` becomes `(readdir($h)) !== false`).

The surrounding `!`, parentheses and `true`/`false` comparison are dropped by the rewrite; when that dropped text contains a comment the finding is reported without a fix.

All edits use exact byte ranges, preserve comments and evaluation order, and require syntax supported by the target PHP version. A fix must never add an evaluation or silently change unrelated arguments.

## Options

None. Use ordinary rule configuration to enable or disable this inspection.

## PHP versions

Available throughout PHP 5.3–8.5. Apply only where the referenced language features and builtin/extension APIs exist. API-specific later syntax or behavior is gated as described in Detection.

## Examples

Finding:

```php
<?php
while (<warning descr="Compare the directory entry with false.">$entry = readdir($dir)</warning>) { echo $entry; }
```

Fixed output:

```php
<?php
while (($entry = readdir($dir)) !== false) { echo $entry; }
```

Valid case:

```php
<?php
while (($entry = readdir($dir)) !== false) { echo $entry; }
```

## Divergences

Native custos inspection, independently specified. No EA counterpart or EA conformance requirement applies. Unknown intent and unproven state are deliberately excluded.
