---
id: ExceptionConstructedNotThrown
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "", max: "" }
---

# ExceptionConstructedNotThrown

## Summary

Constructing an exception and discarding it does not interrupt execution. Throw the exception when the failure should stop the operation.

## Detection

- D1. A standalone new-expression statement constructs a resolved Throwable subtype, with no assignment, return, throw or use of the result. User class ancestry must prove Throwable; recognize resolved builtin exception classes.
- D2. Resolve builtin functions, classes, constants and methods; user-defined lookalikes are excluded. Follow only bounded local proof in one lexical scope.
- D3. Emit one finding per violating operation, not per supporting evidence statement. Use existing analysis budgets; exhausted proof produces no report.

## Exceptions (no report)

- E1. Do not report unresolved or shadowed builtin names, unknown values, or incomplete receiver/type provenance. Unknown calls, aliases escaping, and intervening writes invalidate local proof. Similar code outside the stated detection contract is not a finding.
- E1a. Exclude a Throwable class with a custom constructor or destructor: constructing or disposing of that object can be intentional observable work. Require the supported builtin lifecycle contract, including inherited builtin methods for resolved subtypes.
- E2. Both standard suppression forms suppress this native ID.

## Report

- Range: the violating expression or call, unless D1 specifies a narrower range.
- Severity: warning.
- Enabled by default: true.
- Message: Throw the discarded exception.

## Fix

None. Choosing a repair requires intent or additional runtime information.

## Options

None.

## PHP versions

Available throughout supported PHP 5.3–8.5 where the referenced builtin/extension is available. Respect syntax/version-specific receiver contracts; unavailable APIs are not this rule’s concern.

## Examples

```php
<?php
<warning descr="Throw the discarded exception.">new RuntimeException("Invalid")</warning>;
```

Valid case:

```php
<?php
throw new RuntimeException("Missing record");
```

## Divergences

Native custos inspection, independently specified; no EA counterpart or EA conformance requirement.
