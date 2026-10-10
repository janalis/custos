---
id: WorldWritablePermission
group: Security
kind: semantic
needs: [names, types, flow]
php: { min: "", max: "" }
---

# WorldWritablePermission

## Summary

World-write permission lets other users modify a configured sensitive file. Remove that permission when the project's file policy forbids it. This inspection is disabled by default because the intended policy or use can vary.

## Detection

- D1. Resolved chmod/mkdir mode is a constant mask with other-write bit 0002. Require path matched against configured sensitive path globs (default `**/*.pem`, `**/.env`, `**/config.php`) and a statically known literal path. Unknown paths excluded.
- D2. Resolve builtin functions, classes, constants and methods; user-defined lookalikes are excluded. Follow only bounded local proof in one lexical scope.
- D3. Emit one finding per violating operation, not per supporting evidence statement. Use existing analysis budgets; exhausted proof produces no report.

## Exceptions (no report)

- E1. Do not report unresolved or shadowed builtin names, unknown values, or incomplete receiver/type provenance. Unknown calls, aliases escaping, and intervening writes invalidate local proof. Similar code outside the stated detection contract is not a finding.
- E2. Both standard suppression forms suppress this native ID.

## Report

- Range: the violating expression or call, unless D1 specifies a narrower range.
- Severity: warning.
- Enabled by default: false.
- Message: Remove world-write permission from sensitive files.

## Fix

None. Choosing a repair requires intent or additional runtime information.

## Options

| Option | Type | Default | Effect |
| --- | --- | --- | --- |
| sensitivePaths | list | `["**/*.pem", "**/.env", "**/config.php"]` | Configure the explicit policy evidence described in D1. |

## PHP versions

Available throughout supported PHP 5.3–8.5 where the referenced builtin/extension is available. Respect syntax/version-specific receiver contracts; unavailable APIs are not this rule’s concern.

## Examples

```php
<?php
<warning descr="Remove world-write permission from sensitive files.">chmod("/srv/app/config.php",0777)</warning>;

<warning descr="Remove world-write permission from sensitive files.">mkdir(directory:"/srv/app/.env", permissions:0777)</warning>;
```

Valid case:

```php
<?php
chmod("/srv/app/config.php",0640);
```

## Divergences

Native custos inspection, independently specified; no EA counterpart or EA conformance requirement.
