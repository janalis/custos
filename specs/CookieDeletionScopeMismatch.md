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

- D1. Two resolved cookie setters (`setcookie`/`setrawcookie`) are consecutive statements and use the same literal name; the first stores a nonempty literal value, the second clears the value with a known positive expiry <=1. Highlight the deletion call when their scopes definitely differ:
  - the paths are both explicit nonempty literals and are not the same string, or
  - the domains are both explicit nonempty literals and still differ after lower-casing and dropping one leading dot.
- D1a. Path and domain are read from the options array when the third argument is a literal array, otherwise from the fourth/fifth (or named `path`/`domain`) arguments. Option names are matched case-insensitively, as PHP does. The scalar form is proven when the third argument is absent or a known integer, or when a `path`, `domain`, `secure` or `httponly` argument is present (PHP refuses those next to an options array); the stored cookie's expiry may then be any expression such as `time() + 3600`.
- D2. Resolve builtin functions, classes, constants and methods; user-defined lookalikes are excluded. Follow only bounded local proof in one lexical scope.
- D3. Emit one finding per violating operation, not per supporting evidence statement. Use existing analysis budgets; exhausted proof produces no report.

## Exceptions (no report)

- E1. Do not report unresolved or shadowed builtin names, unknown values, or incomplete receiver/type provenance. Unknown calls, aliases escaping, and intervening writes invalidate local proof. Similar code outside the stated detection contract is not a finding.
- E2. An omitted or empty path or domain is not a definite scope: the browser derives it from the request. A setting omitted or empty on either side never makes that setting differ, so `setcookie("a", "v"); setcookie("a", "", 1, "/");` is not reported.
- E3. Domains differing only by letter case or by a leading dot are the same scope.
- E4. No report when a path or domain is not a literal string, when the third argument is neither a literal array nor a proven scalar expiry, when the options array has unknown keys, or when it spells one option name twice in different letter cases.
- E5. Both standard suppression forms suppress this native ID.

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
setcookie("sid","token",["Path"=>"/app","Domain"=>".Example.com"]);setcookie("sid","",["expires"=>1,"path"=>"/app","domain"=>"example.com"]);
setcookie("lang","fr");setcookie("lang","",1,"/");
```

A computed expiry on the stored cookie still proves the scalar form:

```php
<?php
setcookie("auth","token",time()+3600,"/app"); <warning descr="Delete the cookie using its original path and domain.">setcookie("auth","",1,"/")</warning>;
```

## Divergences

Native custos inspection, independently specified; no EA counterpart or EA conformance requirement.
