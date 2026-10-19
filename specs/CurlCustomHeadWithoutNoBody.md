---
id: CurlCustomHeadWithoutNoBody
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "", max: "" }
---

# CurlCustomHeadWithoutNoBody

## Summary

Setting a custom HEAD method does not configure all of cURL's transfer behavior. Use the no-body option for a HEAD transfer.

## Detection

- D1. Resolved curl handle final CURLOPT_CUSTOMREQUEST is literal HEAD and CURLOPT_NOBODY is definitely false/default before curl_exec. Require handle local ownership and no unknown configuration calls.
- D1e. Account for a later CURLOPT_HTTPGET setter resetting no-body transfer behavior. The finding must describe the final mode rather than an earlier explicit CURLOPT_NOBODY value.
- D1f. Local ownership requires an uncopied handle allocation. Before the reported `curl_exec`, the handle variable may only appear as a direct call argument, as the target of a plain assignment, or in a boolean test (`if`/`elseif` condition, `!`, `isset`, `empty`, `==`, `!=`, `===`, `!==`, `&&`, `||`). Copying it into another binding loses the shared option proof. Uses after the reported call do not matter.
- D2. Resolve builtin functions, classes, constants and methods; user-defined lookalikes are excluded. Follow only bounded local proof in one lexical scope.
- D3. Emit one finding per violating operation, not per supporting evidence statement. When the proof is inconclusive, nothing is reported.

## Exceptions (no report)

- E1. Do not report unresolved or shadowed builtin names, unknown values, or incomplete receiver/type provenance. Unknown calls, aliases escaping, and intervening writes invalidate local proof. Similar code outside the stated detection contract is not a finding.
- E1a. Do not report when the handle or a copy of it leaves the D1f positions before the call, because code the rule cannot follow may then configure it: capture in a closure `use` list, mention inside an arrow function, array literal item (including one passed as an argument), store into a property, array element or static property, reference binding, variable-variable access in the same scope, a `compact()` or `get_defined_vars()` call in the same scope, or any other value position such as `@$h`, `$h ?: $x` or `$h ?? $x`.
- E2. Both standard suppression forms suppress this native ID.

## Report

- Range: the violating expression or call, unless D1 specifies a narrower range.
- Severity: warning.
- Enabled by default: true.
- Message: Use CURLOPT_NOBODY for a HEAD transfer.

## Fix

None. Choosing a repair requires intent or additional runtime information.

## Options

None.

## PHP versions

Available throughout supported PHP 5.3–8.5 where the referenced builtin/extension is available. Respect syntax/version-specific receiver contracts; unavailable APIs are not this rule’s concern.

## Examples

```php
<?php
$h=curl_init();curl_setopt($h,CURLOPT_CUSTOMREQUEST,"HEAD");<warning descr="Use CURLOPT_NOBODY for a HEAD transfer.">curl_exec($h)</warning>;
$other=curl_init();curl_setopt($other,CURLOPT_CUSTOMREQUEST,"HEAD");curl_setopt($other,CURLOPT_NOBODY,false);<warning descr="Use CURLOPT_NOBODY for a HEAD transfer.">curl_exec($other)</warning>;
```

Valid case:

```php
<?php
$h=curl_init($url);curl_setopt($h,CURLOPT_NOBODY,true);curl_exec($h);
```

Escaping handle, not reported:

```php
<?php
function probe($url) {
    $h = curl_init($url);
    curl_setopt($h, CURLOPT_CUSTOMREQUEST, 'HEAD');
    $quiet = function () use ($h) { curl_setopt($h, CURLOPT_NOBODY, true); };
    $quiet();
    return curl_exec($h);
}
```

## Divergences

Native custos inspection, independently specified; no EA counterpart or EA conformance requirement.
