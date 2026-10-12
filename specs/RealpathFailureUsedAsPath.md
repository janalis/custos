---
id: RealpathFailureUsedAsPath
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "", max: "" }
---

# RealpathFailureUsedAsPath

## Summary

Identify concatenation of an unchecked realpath result followed by consumption by a resolved filesystem path API; require local exact provenance.

## Detection

- D1. Report concatenation of an unchecked realpath result followed by consumption by a resolved filesystem path API; require local exact provenance.
- D2. Resolve builtin functions, classes, constants and methods, including imports and supported named arguments. Follow only bounded local facts within one lexical scope. Receiver and argument provenance must establish the specific API contract; variable names and comments never prove intent.
- D4. A result counts as checked when a test on the same `realpath` value rejects `false` on every path to the consuming call. The test may be a strict or loose comparison with `false`, plain truthiness (`$root`, `!$root`), `empty($root)`, the builtin `is_string($root)`, or an assignment tested in place (`if (!($root = realpath($p)))`). It applies as:
  - the branch of an `if`/`elseif`/`else`, `while` body or ternary arm that runs only for a non-false result;
  - the right operand of `&&`/`and` (left proves non-false) or `||`/`or` (left being false proves non-false);
  - an earlier statement of an enclosing block that leaves it on failure: an `if` whose first branch always terminates, or `test || throw …` (PHP 8.0+), `test or exit`/`die`, `failed && exit`.
  - Composition is exact: inside a condition, a test protects the true branch only as a top-level `&&` conjunct and the false branch (or a terminating guard) only as a top-level `||` disjunct. `if ($root || $other)` and `if (!$root && $other) { return; }` prove nothing.
  - The check may sit between the concatenation and the consuming call when the joined string was stored in a variable first.
- D3. Report one finding per violating operation. A reachable path, unchanged value identity and applicable API behavior are required. Intervening mutation, unknown calls involving tracked values, escaped aliases or exhausted analysis budgets invalidate proof.

## Exceptions (no report)

- E1. Exclude unresolved required symbols, user-defined lookalikes, incomplete syntax, unknown required values and unsupported versions. Cases outside D1 receive no finding.
- E3. `assert()` is not a check: assertions may be compiled out in production, so the finding stays. A test on a different variable, on a value reassigned since the test does not count either.
- E2. Both `@custos-ignore RealpathFailureUsedAsPath` and `@noinspection RealpathFailureUsedAsPath` suppress this native rule.
- E4. A value the shared flow analysis proves is not the failure value at the use is never reported, whatever form the proof takes: a test in an enclosing branch, loop condition or short-circuit operand, an early exit, a statement that leaves on the failing outcome (`$v !== false or exit;`, `$v or throw …;`, `$v ?: throw …;`), a failing branch that ends in a call that never returns (a callee declared `never`, or a function or a final or private method every path of which throws or exits), a test stored in a boolean and tested later while neither the boolean nor the tested variable has been written, and, on PHP 8.0 and later outside any `try` with a `catch`, a builtin call that returned after receiving the value for a parameter that only accepts arrays, objects or callables: such a call throws a `TypeError` for `false`. A `switch (true)` case is read like the condition of an `if`/`elseif` chain. Positions PHP may skip prove nothing: the right side of `??=`, the arguments of a null-safe call and anything chained after `?->`. A null-safe call to a never-returning method does not end the path, and `f(...)` creates a closure without calling `f`.

## Report

- Range: the complete violating call or expression. Follow D1 if it specifies a narrower range.
- Severity: warning.
- Enabled by default: true.
- Message: `Reject canonicalization failure before joining paths.`

## Fix

None. A repair requires a semantic decision or application error-handling policy.

## Options

None.

## PHP versions

Available in PHP 5.3–8.5 where the referenced API exists. Do not apply extension contracts when the resolved API is unavailable. Gate syntax and version-specific behavior to the configured target.

## Examples

Finding (enable the rule explicitly when disabled by default):

```php
<?php
$root = realpath($path); file_get_contents($root . '/config.json');
```

Valid case:

```php
<?php
$root = realpath($path); if ($root === false) { throw new RuntimeException(); } file_get_contents($root . '/config.json');
```

Also valid, the failure is rejected by other test shapes:

```php
<?php
function load(string $path) {
    $root = realpath($path);
    if (!$root || !is_dir($root)) {
        return null;
    }
    return file_get_contents($root . '/config.json');
}
function probe(string $path) {
    $root = realpath($path);
    return is_string($root) ? is_file($root . '/config.json') : false;
}
```

Still a finding, the test does not cover every path:

```php
<?php
function load(string $path, bool $force) {
    $root = realpath($path);
    if ($root || $force) {
        return file_get_contents($root . '/config.json');
    }
    return null;
}
```

## Divergences

Native custos inspection, independently specified; no EA counterpart or EA conformance requirement. Conservative proof intentionally excludes unknown runtime state and intent.
