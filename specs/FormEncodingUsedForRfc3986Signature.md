---
id: FormEncodingUsedForRfc3986Signature
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "5.4", max: "" }
---

# FormEncodingUsedForRfc3986Signature

## Summary

Form-style query encoding differs from RFC3986 encoding for spaces. Select RFC3986 encoding when the configured signature protocol requires it. This inspection is disabled by default because the intended policy or use can vary.

## Detection

- D1. Resolved http_build_query result is data argument of hash_hmac and a configured signature protocol requires RFC3986 (default empty configured function list; rule disabled by default). Initial explicit evidence is a configured enclosing function fully qualified name; query contains literal space and encoding is absent/PHP_QUERY_RFC1738. Never infer from variable names.
- D1a. The literal space is searched in the whole literal query data: string values, string keys, and both again inside nested literal arrays (bounded nesting depth). A key is serialized only together with its value, so a spaced key counts only when its value is provably emitted: a known string or integer, or a nested literal array that itself provably emits a pair.
- D2. Resolve builtin functions, classes, constants and methods; user-defined lookalikes are excluded. Follow only bounded local proof in one lexical scope.
- D3. Emit one finding per violating operation, not per supporting evidence statement. When the proof is inconclusive, nothing is reported.

## Exceptions (no report)

- E1. Do not report unresolved or shadowed builtin names, unknown values, or incomplete receiver/type provenance. Unknown calls, aliases escaping, and intervening writes invalidate local proof. Similar code outside the stated detection contract is not a finding.
- E1a. A spaced key whose value is `null`, an empty array, or not statically known is not evidence: PHP drops such a pair entirely or may do so at run time.
- E2. Both standard suppression forms suppress this native ID.

## Report

- Range: the violating expression or call, unless D1 specifies a narrower range.
- Severity: warning.
- Enabled by default: false.
- Message: Use RFC3986 query encoding for this signature protocol.

## Fix

None. Choosing a repair requires intent or additional runtime information.

## Options

| Option | Type | Default | Effect |
| --- | --- | --- | --- |
| signatureFunctions | list | `[]` | Configure the explicit policy evidence described in D1. |

## PHP versions

Requires PHP 5.4 or later.

## Examples

The positive example requires `signatureFunctions: ["signRfc3986"]` in this rule’s options. The default empty list establishes no application signature policy.

```php
<?php
function signRfc3986($key){$q=<warning descr="Use RFC3986 query encoding for this signature protocol.">http_build_query(["q"=>"a b"])</warning>;return hash_hmac("sha256",$q,$key);}
```

Spaces in keys and in nested values are form-encoded too (same option):

```php
<?php
function signRfc3986($key){$q=<warning descr="Use RFC3986 query encoding for this signature protocol.">http_build_query(["full name"=>"ab","tags"=>["dark blue"]])</warning>;return hash_hmac("sha256",$q,$key);}
```

Valid case:

```php
<?php
function signRfc3986($key){$q=http_build_query(["q"=>"a b"],"","&",PHP_QUERY_RFC3986);return hash_hmac("sha256",$q,$key);}
```

## Divergences

Native custos inspection, independently specified; no EA counterpart or EA conformance requirement.
