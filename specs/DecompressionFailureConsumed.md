---
id: DecompressionFailureConsumed
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "5.3", max: "" }
---

# DecompressionFailureConsumed

## Summary

Check decompression before consuming data.

## Detection

- D1. Report results of gzdecode, gzuncompress, gzinflate or inflate_add reaching strlen/strtoupper without a dominating strict false guard. Follow direct expressions and unchanged locals. Exclude matching encoder/decoder roundtrips with default encoding and omitted or valid literal compression level.
- D1a. Error suppression does not change the result: a decompression call wrapped in `@` (directly or through the assigned local) is followed like the bare call.
- D1b. A strict false guard is a `=== false` / `!== false` comparison, with the constant on either side, whose other operand is the tracked local or the parenthesised plain assignment that stored the reaching result. It dominates the consumer when the consumer sits in: the `if`/`elseif` branch, `while` body or `for` body (last condition expression) selected by a passing test; a later `elseif`/`else` arm after a failure test; the matching ternary arm (the else arm of a short ternary after a failure test); the right operand of `&&`/`and` after a passing test or of `||`/`or` after a failure test; or the statements following an `if` without alternatives whose failure branch always terminates. Inside a condition, `!` inverts the test, a true conjunction establishes each of its operands and a false disjunction refutes each of its operands. The tested value must be the one that reaches the consumer: reassignment between test and use, or a repeated call, is a different value.
- D2. Resolve builtin functions, extension classes, method receivers, constants and named arguments. Follow unchanged locals and aliases only within the same bounded lexical scope. Require a reachable path; merge branch facts only when all incoming paths establish them.
- D3. Emit one finding at the violating operation, not one per evidence statement. Rebinding, writes, by-reference escape, unresolved calls that can mutate relevant state, and exhausted analysis budgets remove proof.

## Exceptions (no report)

A nested encoder proves matching bytes only when its encoding argument is omitted; argument count does not establish that omission.

- E1. Exclude shadowed builtins, unresolved receivers, incomplete syntax and unknown prerequisite facts. An extension function name alone does not prove an application function is that builtin.
- E2. A dominating valid guard or repair described in Detection prevents the finding. Both legacy noinspection and custos-ignore forms suppress this ID.
- E3. `assert`, `match (true)` arms and a statement-level `$s === false and exit;` are not guards under D1b; such code is still reported unless flow analysis independently excludes false.
- E4. A value the shared flow analysis proves is not false at the use is never reported, whatever form the proof takes: a truthiness or `empty()` test, a loose comparison with `false` or `null`, a type predicate, or an early exit whose `if` also has an `else` branch that continues. So `$s ? strlen($s) : 0` and `$s != false && strlen($s) > 3` are not reported, while the wrong arm or operand of such a test still is. The same holds for a statement that leaves on the failing outcome (`$v !== false or exit;`, `$v or throw …;`, `$v ?: throw …;`), for a failing branch that ends in a call that never returns (a callee declared `never`, or a function or a final or private method every path of which throws or exits), for a test stored in a boolean and tested later while neither the boolean nor the tested variable has been written, and, on PHP 8.0 and later outside any `try` with a `catch`, for a builtin call that returned after receiving the value for a parameter that only accepts arrays, objects or callables: such a call throws a `TypeError` for `false`. Positions PHP may skip prove nothing: the right side of `??=`, the arguments of a null-safe call and anything chained after `?->`. A null-safe call to a never-returning method does not end the path, and `f(...)` creates a closure without calling `f`.

## Report

- Range: the violating expression or call described in D1; stateful absence checks highlight the final unsafe operation or loop.
- Severity: warning.
- Enabled by default: true.
- Message: `Check decompression before consuming data.`

## Fix

No automatic fix. The required repair depends on error handling, data selection, lifecycle ordering or application policy.

All edits use exact byte ranges and preserve comments and evaluation order. Offer no fix when its stated prerequisites cannot be proven or the replacement is unsupported by the configured PHP version.

## Options

None.

## PHP versions

Requires PHP 5.3 or later and the referenced extension API. Respect method-specific introduction versions within PHP 5.3–8.5; unsupported optional APIs cannot establish facts. Named arguments require PHP 8.0 or later.

## Examples

```php
<?php
$s=gzdecode($bytes); <warning descr="Check decompression before consuming data.">strlen($s)</warning>;
```

Valid case:

```php
<?php
$s=gzdecode($bytes); if ($s !== false) { echo strlen($s); }
```

Guards inside expressions and conditions are also valid:

```php
<?php
function size($bytes) {
    if (($s = @gzdecode($bytes)) === false) { return 0; }
    return strlen($s);
}
function nonEmpty($bytes) {
    $s = gzinflate($bytes);
    return $s !== false && strlen($s) > 0 ? strtoupper($s) : null;
}
```

Error suppression does not hide the finding:

```php
<?php
$s = @gzuncompress($bytes); <warning descr="Check decompression before consuming data.">strlen($s)</warning>;
```

## Divergences

Native custos inspection, independently specified. No EA counterpart or EA fixture conformance applies.
