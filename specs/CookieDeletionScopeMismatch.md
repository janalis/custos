---
id: CookieDeletionScopeMismatch
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "", max: "" }
---

# CookieDeletionScopeMismatch

## Summary

Cookie deletion must target the same path and domain as creation. Use the original scope when expiring the cookie.

## Detection

- D1. Two resolved cookie setters in one straight-line scope use same literal name; first stores nonempty value, second clears value with known past positive expiry <=1. Their definite literal path/domain scopes differ. Highlight deletion call.
- D2. Resolve builtin functions, classes, constants and methods; user-defined lookalikes are excluded. Follow only bounded local proof in one lexical scope.
- D3. Emit one finding per violating operation, not per supporting evidence statement. Use existing analysis budgets; exhausted proof produces no report.

## Exceptions (no report)

- E1. Do not report unresolved or shadowed builtin names, unknown values, or incomplete receiver/type provenance. Unknown calls, aliases escaping, and intervening writes invalidate local proof. Similar code outside the stated detection contract is not a finding.
- E2. Both standard suppression forms suppress this native ID.

## Report

- Range: the violating expression or call, unless D1 specifies a narrower range.
- Severity: warning.
- Enabled by default: true.
- Message: Delete the cookie using its original path and domain.

## Fix

None. Choosing a repair requires intent or additional runtime information.

## Options

None.

## PHP versions

Available throughout supported PHP 5.3–8.5 where the referenced builtin/extension is available. Respect syntax/version-specific receiver contracts; unavailable APIs are not this rule’s concern.

## Examples

```php
<?php
setcookie("auth","token",["path"=>"/app"]); <warning descr="Delete the cookie using its original path and domain.">setcookie("auth","",["expires"=>1,"path"=>"/"])</warning>;
```

Valid case:

```php
<?php
setcookie("auth",$token,["path"=>"/app"]);setcookie("auth","",["expires"=>1,"path"=>"/app"]);
```

## Divergences

Native custos inspection, independently specified; no EA counterpart or EA conformance requirement.
