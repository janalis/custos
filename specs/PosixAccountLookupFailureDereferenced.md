---
id: PosixAccountLookupFailureDereferenced
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "", max: "" }
---

# PosixAccountLookupFailureDereferenced

## Summary

Identify array indexing on a direct or unchanged posix_getpwnam/posix_getpwuid result that a failed lookup (`false`) can still reach.

## Detection

- D1. Report array indexing on a direct or unchanged posix_getpwnam/posix_getpwuid result unless a guard proves the failure value `false` cannot reach the read (see D4).
- D2. Resolve builtin functions, classes, constants and methods, including imports and supported named arguments. Follow only bounded local facts within one lexical scope. Receiver and argument provenance must establish the specific API contract; variable names and comments never prove intent.
- D3. Report one finding per violating operation. A reachable path, unchanged value identity and applicable API behavior are required. Intervening mutation, unknown calls involving tracked values, escaped aliases or exhausted analysis budgets invalidate proof.

- D4. A guard is a test of the same unchanged variable, written after the lookup, whose outcome on the path to the read excludes `false`. Accepted tests, each usable positively or through `!`:
  - comparison with `false` using `===`, `!==`, `==` or `!=`, in either operand order;
  - truthiness of the variable itself, or of the assignment `$v = posix_getpwnam(...)` used as a condition (a successful lookup is a non-empty array);
  - `is_array($v)`, `isset($v[...])` with at least one key (plain `isset($v)` accepts `false`), and a false `empty($v)` or `empty($v[...])`;
  - `&&`/`and` and `||`/`or` combinations of the above, where a conjunction holds if either side proves it and a disjunction only if both do (and dually when the combination is false).

  A guard applies to the read when the read sits in the matching `if`/`elseif`/`else` branch or ternary arm, in the right operand of a short-circuit operator whose left operand is the guard, or after an earlier statement of an enclosing statement list that leaves it on failure: an `if` whose body always terminates without a `goto` (which could land later in the same list), or an expression statement `guard or throw ...` / `guard or exit` (respectively `failure and throw/exit`). A direct `posix_getpwuid(...)[...]` read has no variable to test and is always reported.

## Exceptions (no report)

- Array and list destructuring destinations create slots rather than reading them. Destructuring keys and ordinary array values still require reads.

- Guarded probes with isset, empty or null coalescing, unsets and plain nested slot creation do not read a failed or missing value. Index expressions and the fallback operand of null coalescing still require normal reads.

- E1. Exclude unresolved required symbols, user-defined lookalikes, incomplete syntax, unknown required values and unsupported versions. Cases outside D1 receive no finding.
- E2. Both `@custos-ignore PosixAccountLookupFailureDereferenced` and `@noinspection PosixAccountLookupFailureDereferenced` suppress this native rule.

## Report

- Range: the complete violating call or expression. Follow D1 if it specifies a narrower range.
- Severity: warning.
- Enabled by default: true.
- Message: `Reject account lookup failure before indexing.`

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
$account = posix_getpwnam($name); echo $account['uid'];
```

Valid case:

```php
<?php
$account = posix_getpwnam($name); if ($account !== false) { echo $account['uid']; }
```

Also valid, the failure never reaches the read:

```php
<?php
$owner = posix_getpwuid($uid) or exit('unknown owner');
echo $owner['name'];
$home = is_array($owner) ? $owner['dir'] : null;
```

## Divergences

Native custos inspection, independently specified; no EA counterpart or EA conformance requirement. Conservative proof intentionally excludes unknown runtime state and intent.
