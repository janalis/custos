---
id: PdoExecZeroRejectedAsFailure
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "", max: "" }
---

# PdoExecZeroRejectedAsFailure

## Summary

Identify proven PDO::exec result negated or loosely compared with false in a failure branch throwing or returning false. Strict false excludes.

## Detection

- D1. Report proven PDO::exec result negated or loosely compared with false in a failure branch throwing or returning false. Strict false excludes.
  The failure branch is the body of the `if` whose condition is the negated or loosely compared result. It counts when it holds a reachable `throw` or `return false`, outside nested functions and classes, that satisfies one of:
  - no `errorCode()` or `errorInfo()` call on the same receiver expression starts before it inside the branch (a call inside the thrown expression itself, such as an exception message built from `errorInfo()`, starts after the `throw` and does not count);
  - it is a statement of the branch itself, not nested in another construct or a `throw` sub-expression, and every statement before it in the branch is a plain expression or `echo` statement, so nothing in the branch decides whether it runs.
- D2. Resolve builtin functions, classes, constants and methods, including imports and supported named arguments. Follow only bounded local facts within one lexical scope. Receiver and argument provenance must establish the specific API contract; variable names and comments never prove intent.
- D3. Report one finding per violating operation. A reachable path, unchanged value identity and applicable API behavior are required. Intervening mutation, unknown calls involving tracked values, escaped aliases or exhausted analysis budgets invalidate proof.

## Exceptions (no report)

- E1. Exclude unresolved required symbols, user-defined lookalikes, incomplete syntax, unknown required values and unsupported versions. Cases outside D1 receive no finding.
- E2. No report when every `throw` and `return false` of the failure branch comes after an `errorCode()` or `errorInfo()` read on the same receiver and is either nested in another construct or preceded by a statement other than a plain expression or `echo`. Such a branch can tell zero affected rows from a failure before rejecting, for example by throwing only when `errorCode()` is not `'00000'`, or by returning early when it is. The receivers are compared textually; the error state held in a variable is covered because the read still precedes the guarded exit.
- E3. Both `@custos-ignore PdoExecZeroRejectedAsFailure` and `@noinspection PdoExecZeroRejectedAsFailure` suppress this native rule.

## Report

- Range: the complete violating call or expression. Follow D1 if it specifies a narrower range.
- Severity: warning.
- Enabled by default: true.
- Message: `Compare database execution failure strictly with false.`

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
$pdo = new PDO($dsn); if (!$pdo->exec('DELETE FROM jobs WHERE done = 1')) { throw new RuntimeException(); }
```

Valid case:

```php
<?php
$pdo = new PDO($dsn); if ($pdo->exec('DELETE FROM jobs WHERE done = 1') === false) { throw new RuntimeException(); }
```

Valid case, the branch separates zero rows from failure:

```php
<?php
function purge(PDO $pdo, string $sql): int {
    if (!$count = $pdo->exec($sql)) {
        if ($pdo->errorCode() !== '00000') { throw new RuntimeException('purge failed'); }
    }
    return (int) $count;
}
```

## Divergences

Native custos inspection, independently specified; no EA counterpart or EA conformance requirement. Conservative proof intentionally excludes unknown runtime state and intent.
