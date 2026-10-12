---
id: ParseUrlFailureDereferenced
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "", max: "" }
---

# ParseUrlFailureDereferenced

## Summary

Identify ordinary unguarded reads of dynamic parse_url results, or literal results with a supported parsing-failure proof. Restrict to array-returning mode; missing keys are another concern.

## Detection

- D1. Report ordinary unguarded reads of dynamic parse_url results, or literal results with a supported parsing-failure proof. Empty HTTP(S) authorities and ports above 65535 establish such a proof. Restrict to array-returning mode; missing keys are another concern.
- D2. Resolve builtin functions, classes, constants and methods, including imports and supported named arguments. Follow only bounded local facts within one lexical scope. Receiver and argument provenance must establish the specific API contract; variable names and comments never prove intent.
- D3. Report one finding per violating operation. A reachable path, unchanged value identity and applicable API behavior are required. Intervening mutation, unknown calls involving tracked values, escaped aliases or exhausted analysis budgets invalidate proof.
- D3a. When several assignments reach the read (a loop that reassigns the variable, or an assignment made on one path only), the read is reported if one of them is a result of this API that no guard on its own path covers. A guard written before the other assignment does not cover it; a test inside the loop, or after the conditional assignment, covers every assignment. A reaching value that is not a result of this API (a literal, another call, an unknown value) never causes a report by itself.
- D4. A read is guarded when a condition on the same variable, written after the producing call and before any reassignment, excludes `false` on every path to the read. Accepted condition shapes:
  - strict or loose comparison with `false` (`$parts === false`, `false != $parts`);
  - plain truthiness (`$parts`, `!$parts`) and `empty($parts)`;
  - `is_array($parts)` and `array_key_exists($key, $parts)` resolved to the builtins;
  - `isset` or `empty` on a key of the variable (`isset($parts['host'])`, `!empty($parts['path'])`); `isset($parts)` alone proves nothing because `false` is set.
  - negation, and `&&`/`||` (also `and`/`or`) combinations: a true conjunction needs one proving operand, a false conjunction needs all of them, and symmetrically for disjunctions.
- D5. A guard applies in these positions: the matching branch of an enclosing `if`/`elseif`/`else` (later branches also know the earlier conditions failed); the matching arm of an enclosing ternary; the right operand of `&&` or `||` whose left operand is the guard; and any statement after an earlier sibling `if` whose main body always exits, at every enclosing block level inside the same function scope.

## Exceptions (no report)

- Array and list destructuring destinations create slots rather than reading them. Destructuring keys and ordinary array values still require reads.

- Literal URL shapes outside the supported failure proof receive no finding. This avoids requesting runtime failure guards for constants whose parsing outcome is unsupported.

- Guarded probes with isset, empty or null coalescing, unsets and plain nested slot creation do not read a failed or missing value. Index expressions and the fallback operand of null coalescing still require normal reads.

- Reads guarded as described by D4 and D5 never index the failure value.
- A value the shared flow analysis proves is not false at the read is never reported, whatever form the proof takes: the conditions of D4 in the positions of D5, a loop condition, a statement that leaves on the failing outcome (`$parts !== false or exit;`, `$parts or throw …;`, `$parts ?: throw …;`), a failing branch that ends in a call that never returns (a callee declared `never`, or a function or a final or private method every path of which throws or exits), a test stored in a boolean and tested later while neither the boolean nor the tested variable has been written, and, on PHP 8.0 and later outside any `try` with a `catch`, a builtin call that returned after receiving the value for a parameter that only accepts arrays, objects or callables (`count($parts)` throws a `TypeError` for `false`; before PHP 8 it warns and returns, which proves nothing). Positions PHP may skip prove nothing: the right side of `??=`, the arguments of a null-safe call and anything chained after `?->`. A null-safe call to a never-returning method does not end the path, and `f(...)` creates a closure without calling `f`.

- `assert()` is not a guard: assertions are compiled out in production configurations, so the read stays reported. Indirect evidence (a null-coalesced copy tested later, `switch (true)` cases) is not followed either.

- E1. Exclude unresolved required symbols, user-defined lookalikes, incomplete syntax, unknown required values and unsupported versions. Cases outside D1 receive no finding.
- E2. Both `@custos-ignore ParseUrlFailureDereferenced` and `@noinspection ParseUrlFailureDereferenced` suppress this native rule.

## Report

- Range: the complete violating call or expression. Follow D1 if it specifies a narrower range.
- Severity: warning.
- Enabled by default: true.
- Message: `Reject URL parsing failure before indexing.`

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
$parts = parse_url($raw); echo $parts['host'];
```

Valid case:

```php
<?php
$parts = parse_url($raw); if ($parts === false) { throw new RuntimeException(); } echo $parts['host'];
```

Valid cases with other guard shapes:

```php
<?php
$parts = parse_url($raw);
if (!is_array($parts)) { return null; }
$host = isset($parts['host']) ? $parts['host'] : 'localhost';
return $parts !== false && $parts['scheme'] === 'https';
```

## Divergences

Native custos inspection, independently specified; no EA counterpart or EA conformance requirement. Conservative proof intentionally excludes unknown runtime state and intent.
