---
id: HardcodedCredentialAtKnownSink
group: Security
kind: semantic
needs: [names, types, flow]
php: { min: "", max: "" }
---

# HardcodedCredentialAtKnownSink

## Summary

Literal credentials embedded in connection setup are difficult to rotate and protect. Load them from the application's external configuration. This inspection is disabled by default because the intended policy or use can vary.

## Detection

- D1. Resolved PDO constructor password or mysqli connection password is a nonempty literal string. Inspection is disabled by default and excludes configured example/test path glob list, default **/testdata/** and **/tests/**; no arbitrary password variable naming heuristics.
- D2. Resolve builtin functions, classes, constants and methods; user-defined lookalikes are excluded. Follow only bounded local proof in one lexical scope.
- D3. Emit one finding per violating operation, not per supporting evidence statement. Use existing analysis budgets; exhausted proof produces no report.

## Exceptions (no report)

- E1. Do not report unresolved or shadowed builtin names, unknown values, or incomplete receiver/type provenance. Unknown calls, aliases escaping, and intervening writes invalidate local proof. Similar code outside the stated detection contract is not a finding.
- E2. Both standard suppression forms suppress this native ID.

## Report

- Range: the violating expression or call, unless D1 specifies a narrower range.
- Severity: warning.
- Enabled by default: false.
- Message: Load credentials from external configuration.

## Fix

None. Choosing a repair requires intent or additional runtime information.

## Options

| Option | Type | Default | Effect |
| --- | --- | --- | --- |
| excludedPaths | list | `["**/testdata/**", "**/tests/**"]` | Configure the explicit policy evidence described in D1. |

## PHP versions

Available throughout supported PHP 5.3–8.5 where the referenced builtin/extension is available. Respect syntax/version-specific receiver contracts; unavailable APIs are not this rule’s concern.

## Examples

```php
<?php
new PDO($dsn,"service",<warning descr="Load credentials from external configuration.">"permanent-password"</warning>);
```

Valid case:

```php
<?php
$db=new PDO($dsn,$user,getenv("DB_PASSWORD"));
```

## Divergences

Native custos inspection, independently specified; no EA counterpart or EA conformance requirement.
