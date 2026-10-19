---
id: AuthenticationCookieAllowsScriptAccess
group: Security
kind: semantic
needs: [names, types, flow]
php: { min: "", max: "" }
---

# AuthenticationCookieAllowsScriptAccess

## Summary

An authentication cookie without HttpOnly is available to scripts. Enable HttpOnly when the configured authentication policy requires it. This inspection is disabled by default because the intended policy or use can vary.

## Detection

- D1. Resolved setcookie/setrawcookie has a configured authentication cookie name (default auth_token) and an explicit false httponly, given either in a literal options array or as the legacy positional/named `httponly` argument. Missing httponly is excluded in initial contract to avoid policy ambiguity.
  - Options array: the third argument is a literal array whose keys are all known. Option names are matched case-insensitively, as PHP does (`"HttpOnly"`, `"HTTPONLY"`); when several spellings are present the element that comes last in array order decides.
  - Legacy argument: whenever the third argument is not a literal array, an explicit `httponly` argument decides, whatever the expiry expression is (`0`, `time() + 3600`, a variable, a call). PHP refuses to combine an options array with the legacy arguments, so their presence proves the legacy signature.
- D2. Resolve builtin functions, classes, constants and methods; user-defined lookalikes are excluded. Follow only bounded local proof in one lexical scope.
- D3. Emit one finding per violating operation, not per supporting evidence statement. When the proof is inconclusive, nothing is reported.

## Exceptions (no report)

- E1. Do not report unresolved or shadowed builtin names, unknown values, or incomplete receiver/type provenance. Unknown calls, aliases escaping, and intervening writes invalidate local proof. Similar code outside the stated detection contract is not a finding.
- E2. Both standard suppression forms suppress this native ID.

## Report

- Range: the violating expression or call, unless D1 specifies a narrower range.
- Severity: warning.
- Enabled by default: false.
- Message: Enable HttpOnly for the authentication cookie.

## Fix

None. Choosing a repair requires intent or additional runtime information.

## Options

| Option | Type | Default | Effect |
| --- | --- | --- | --- |
| authenticationCookies | list | `["auth_token"]` | Configure the explicit policy evidence described in D1. |

## PHP versions

Available throughout supported PHP 5.3–8.5 where the referenced builtin/extension is available. Respect syntax/version-specific receiver contracts; unavailable APIs are not this rule’s concern.

## Examples

```php
<?php
setcookie("auth_token",$token,["secure"=>true,"httponly"=><warning descr="Enable HttpOnly for the authentication cookie.">false</warning>]);
setcookie("auth_token",$token,["secure"=>true,"HttpOnly"=><warning descr="Enable HttpOnly for the authentication cookie.">false</warning>]);
setcookie("auth_token",$token,time()+3600,"/","",true,<warning descr="Enable HttpOnly for the authentication cookie.">false</warning>);
```

Valid case:

```php
<?php
setcookie("auth_token",$token,["secure"=>true,"httponly"=>true]);
setcookie("auth_token",$token,["httponly"=>false,"HttpOnly"=>true]);
setcookie("auth_token",$token,time()+3600,"/","",true,true);
setcookie("auth_token",$token,$options);
```

## Divergences

Native custos inspection, independently specified; no EA counterpart or EA conformance requirement.
