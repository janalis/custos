---
id: TransliterationFailureConsumed
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "5.4", max: "" }
---

# TransliterationFailureConsumed

## Summary

Check transliteration before consuming text.

## Detection

- D1. Track results of resolved transliterator_transliterate or Transliterator::transliterate to strlen or strtoupper without a dominating strict false guard. Direct expressions and unchanged locals are supported; storage alone is excluded.
- D2. Resolve builtin functions, extension classes, method receivers, constants and named arguments. Follow unchanged locals and aliases only within the same bounded lexical scope. Require a reachable path; merge branch facts only when all incoming paths establish them.
- D3. Emit one finding at the violating operation, not one per evidence statement. Rebinding, writes, by-reference escape, unresolved calls that can mutate relevant state, and exhausted analysis budgets remove proof.
- D4. A guard dominates the consumer when the consumer is reached only after a condition that excludes false for the same unchanged value. Accepted conditions are a strict comparison with false (`$s !== false` holding, `$s === false` failing, either operand order), `is_string($s)` holding for the resolved builtin, a negation of these, and `&&`/`and`/`||`/`or` combinations whose outcome implies one of them. The compared operand may be the plain assignment that stores the result, as in `($s = transliterator_transliterate(...)) === false`. Accepted positions are: the matching branch of an `if`, `elseif` or `else`; the matching branch of a ternary; the right operand of `&&`/`and` (left held) or `||`/`or` (left failed); an arm of `match (true)` whose single condition holds or that follows an earlier arm whose condition failed; and the statement directly after an `if` without alternatives whose body always leaves.

## Exceptions (no report)

- E1. Exclude shadowed builtins, unresolved receivers, incomplete syntax and unknown prerequisite facts. An extension function name alone does not prove an application function is that builtin.
- E2. A dominating valid guard or repair described in Detection prevents the finding. Both legacy noinspection and custos-ignore forms suppress this ID.
- E3. A value the shared flow analysis proves is not false at the use is never reported, whatever form the proof takes: a truthiness or `empty()` test, a loose comparison with `false` or `null`, a type predicate, or an early exit whose `if` also has an `else` branch that continues. The result is a string or false, so `if (!$s) { return; }`, `$s == false ? 0 : strlen($s)` and `$s ? strlen($s) : 0` leave nothing to report. A test that false can pass (`isset($s)`, `$s !== null`), the wrong branch of a test, and a test of another value prove nothing. The same holds for a statement that leaves on the failing outcome (`$v !== false or exit;`, `$v or throw …;`, `$v ?: throw …;`), for a failing branch that ends in a call that never returns (a callee declared `never`, or a function or a final or private method every path of which throws or exits), for a test stored in a boolean and tested later while neither the boolean nor the tested variable has been written, and, on PHP 8.0 and later outside any `try` with a `catch`, for a builtin call that returned after receiving the value for a parameter that only accepts arrays, objects or callables: such a call throws a `TypeError` for `false`.

## Report

- Range: the violating expression or call described in D1; stateful absence checks highlight the final unsafe operation or loop.
- Severity: warning.
- Enabled by default: true.
- Message: `Check transliteration before consuming text.`

## Fix

No automatic fix. The required repair depends on error handling, data selection, lifecycle ordering or application policy.

All edits use exact byte ranges and preserve comments and evaluation order. Offer no fix when its stated prerequisites cannot be proven or the replacement is unsupported by the configured PHP version.

## Options

None.

## PHP versions

Requires PHP 5.4 or later and the referenced extension API. Respect method-specific introduction versions within PHP 5.3–8.5; unsupported optional APIs cannot establish facts. Named arguments require PHP 8.0 or later.

## Examples

```php
<?php
$s=transliterator_transliterate($transform,$input); <warning descr="Check transliteration before consuming text.">strlen($s)</warning>;
```

Valid case:

```php
<?php
$s=transliterator_transliterate($transform,$input); if ($s !== false) { echo strlen($s); }
```

Guards written as expressions are valid too:

```php
<?php
$s = transliterator_transliterate($transform, $input);
echo $s === false ? 0 : strlen($s);
echo $s !== false && strlen($s) > 3 ? 'long' : 'short';
echo match (true) { $s === false => 0, default => strlen($s) };
if (($u = transliterator_transliterate($transform, $input)) === false) { return; }
echo strtoupper($u);
```

## Divergences

Native custos inspection, independently specified. No EA counterpart or EA fixture conformance applies.
