---
id: UnicodeNormalizationFailureConsumed
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "5.3", max: "" }
---

# UnicodeNormalizationFailureConsumed

## Summary

Check normalization before consuming text.

## Detection

- D1. Report a result of resolved Normalizer::normalize or normalizer_normalize reaching strlen or strtoupper without a dominating strict false guard. Follow direct expressions and unchanged local assignments; storage and testing alone are excluded.
- D2. Resolve builtin functions, extension classes, method receivers, constants and named arguments. Follow unchanged locals and aliases only within the same bounded lexical scope. Require a reachable path; merge branch facts only when all incoming paths establish them.
- D3. Emit one finding at the violating operation, not one per evidence statement. Rebinding, writes, by-reference escape, unresolved calls that can mutate relevant state, and exhausted analysis budgets remove proof.
- D4. A strict false guard is a condition that, on the path reaching the consumption, proves the value is not false. Accepted atoms: the value compared with `false` by `===` or `!==` (either operand order), and builtin `is_string()` applied to the value, alone or compared strictly with a boolean literal. The compared operand may be a plain assignment of the normalization result to the consumed variable written inside the condition. Atoms combine through `!`, through `&&`/`and` when the condition holds and through `||`/`or` when it fails.
- D5. A guard dominates the consumption when the consumption is in: the guarded branch of an `if`/`elseif`/`else` chain (a later branch sees every earlier condition failed); the matching operand of a ternary; the right operand of a short-circuit `&&`, `and`, `||` or `or`; an arm of `match (true)` whose single condition proves the value, or any arm after a condition that fails only when the value is not false; an arm of a `match` on the value itself placed after a literal `false` condition; or any statement after an earlier sibling `if` without alternatives whose body always leaves the path, or after an expression statement that short-circuits into `exit` or `throw` when the value is false. The guard must test the same reaching normalization result, after it was produced, with no rebinding or invalidation in between.

## Exceptions (no report)

- E1. Exclude shadowed builtins, unresolved receivers, incomplete syntax and unknown prerequisite facts. An extension function name alone does not prove an application function is that builtin.
- E2. A dominating valid guard or repair described in Detection prevents the finding. Both legacy noinspection and custos-ignore forms suppress this ID.
- E3. A value the shared flow analysis proves is not false at the use is never reported, whatever form the proof takes: a truthiness or `empty()` test, a loose comparison with `false` or `null`, a type predicate, or an early exit whose `if` also has an `else` branch that continues. The result is a string or false, so `if (!$s) { return; }`, `if ($s == false) { return; }` and `$s ? strlen($s) : 0` leave nothing to report. `assert()` can be compiled out and a short ternary on another value tests nothing: both stay reported. The same holds for a statement that leaves on the failing outcome (`$v !== false or exit;`, `$v or throw …;`, `$v ?: throw …;`), for a failing branch that ends in a call that never returns (a callee declared `never`, or a function or a final or private method every path of which throws or exits), for a test stored in a boolean and tested later while neither the boolean nor the tested variable has been written, and, on PHP 8.0 and later outside any `try` with a `catch`, for a builtin call that returned after receiving the value for a parameter that only accepts arrays, objects or callables: such a call throws a `TypeError` for `false`.

## Report

- Range: the violating expression or call described in D1; stateful absence checks highlight the final unsafe operation or loop.
- Severity: warning.
- Enabled by default: true.
- Message: `Check normalization before consuming text.`

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
$s = Normalizer::normalize($input); <warning descr="Check normalization before consuming text.">strlen($s)</warning>;
```

Valid case:

```php
<?php
$s = Normalizer::normalize($input); if ($s !== false) { echo strlen($s); }
```

Guards written as expressions are valid too:

```php
<?php
$s = normalizer_normalize($input);
echo $s === false ? 0 : strlen($s);
echo $s !== false && strlen($s) > 3 ? 'long' : 'short';
echo match (true) { $s === false => '', default => strtoupper($s) };
if (($t = normalizer_normalize($other)) === false) { return; }
echo strlen($t);
```

## Divergences

Native custos inspection, independently specified. No EA counterpart or EA fixture conformance applies.
