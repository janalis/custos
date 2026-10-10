---
id: HostCookiePrefixContractViolation
group: Security
kind: semantic
needs: [names, types, flow]
php: { min: "", max: "" }
---

# HostCookiePrefixContractViolation

## Summary

A __Host- cookie requires Secure, path /, and no Domain attribute. Set those attributes together so browsers accept the intended cookie.

## Detection

- D1. Resolved setcookie/setrawcookie name is literal beginning __Host- and proven options violate Secure=true, Path=/, or omit Domain. Missing path/secure uses their builtin defaults; dynamic options excluded.
- D2. Resolve builtin functions, classes, constants and methods; user-defined lookalikes are excluded. Follow only bounded local proof in one lexical scope.
- D3. Emit one finding per violating operation, not per supporting evidence statement. Use existing analysis budgets; exhausted proof produces no report.

## Exceptions (no report)

- E1. Do not report unresolved or shadowed builtin names, unknown values, or incomplete receiver/type provenance. Unknown calls, aliases escaping, and intervening writes invalidate local proof. Similar code outside the stated detection contract is not a finding.
- E2. Both standard suppression forms suppress this native ID.

## Report

- Range: the violating expression or call, unless D1 specifies a narrower range.
- Severity: warning.
- Enabled by default: true.
- Message: Set Secure and Path=/ and omit Domain for __Host- cookies.

## Fix

None. Choosing a repair requires intent or additional runtime information.

## Options

None.

## PHP versions

Available throughout supported PHP 5.3–8.5 where the referenced builtin/extension is available. Respect syntax/version-specific receiver contracts; unavailable APIs are not this rule’s concern.

## Examples

```php
<?php
<warning descr="Set Secure and Path=/ and omit Domain for __Host- cookies.">setcookie("__Host-auth",$token,["secure"=>true,"path"=>"/app"])</warning>;
```

Valid case:

```php
<?php
setcookie("__Host-auth",$token,["secure"=>true,"path"=>"/"]);
```

## Divergences

Native custos inspection, independently specified; no EA counterpart or EA conformance requirement.
