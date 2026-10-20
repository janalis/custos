---
id: AeadDecryptionTagLengthUnchecked
group: Security
kind: semantic
needs: [names, types, flow]
php: { min: "7.1", max: "" }
---

# AeadDecryptionTagLengthUnchecked

## Summary

Authenticated decryption needs a tag of the expected length. Validate request-supplied tag lengths before calling the decryptor. This inspection is disabled by default because the intended policy or use can vary.

## Detection

- D1. Resolved openssl_decrypt uses literal GCM cipher and tag directly derived from request superglobal, without a dominating strlen(tag)===16 guard. Restrict to configured expected GCM tag length 16 by default; CCM and known constant tags excluded.
- D1g. A guard is a comparison of a resolved `strlen` call (or `mb_strlen` with the `8bit` encoding) of the tag with an integer, in either operand order, strict or loose (`strlen` returns an integer, so both read alike), that leaves only the expected length where the tag is read. It may sit in any guarding position: the branch of an `if`/`elseif`/`else` or of a ternary the comparison selects, the right operand of a short-circuit operator, or code after an earlier `if` whose body always exits. The measured expression must be the value passed as the tag: the same variable while it still holds the measured value, or a read of the same key of the same request superglobal while nothing can have written it: no write to that element or the whole superglobal, and no call that may run project code, which can write any superglobal, between the guard and the tag (also later in a loop around the tag). A short ternary, another length and a guard on another element prove nothing.
- D1e. Source provenance additionally requires an untouched lexical superglobal prefix. Exclude earlier explicit writes to or reference escapes of that superglobal and any earlier write or reference target involving GLOBALS. Reject all earlier object construction and unknown calls, including calls with no arguments, method calls and static calls; only resolved strlen, is_string, ctype_alnum, ctype_digit or in_array are supported pure-call exceptions. Traverse at most 4096 AST nodes, excluding nested variable scopes; exhaustion or unsupported source mutation discards the request-origin proof.
- D2. Resolve builtin functions, classes, constants and methods; user-defined lookalikes are excluded. Follow only bounded local proof in one lexical scope.
- D3. Emit one finding per violating operation, not per supporting evidence statement. When the proof is inconclusive, nothing is reported.

## Exceptions (no report)

- E1. Do not report unresolved or shadowed builtin names, unknown values, or incomplete receiver/type provenance. Unknown calls, aliases escaping, and intervening writes invalidate local proof. Similar code outside the stated detection contract is not a finding.
- E2. Both standard suppression forms suppress this native ID.

## Report

- Range: the unchecked authentication tag argument.
- Severity: warning.
- Enabled by default: false.
- Message: Validate the authentication tag length before decrypting.

## Fix

None. Choosing a repair requires intent or additional runtime information.

## Options

| Option | Type | Default | Effect |
| --- | --- | --- | --- |
| tagLength | int | `16` | Expected GCM tag length in bytes that the length guard must check. |

## PHP versions

Requires PHP 7.1 or later.

## Examples

```php
<?php
openssl_decrypt($cipher,"aes-256-gcm",$key,OPENSSL_RAW_DATA,$iv,<warning descr="Validate the authentication tag length before decrypting.">$_POST["tag"]</warning>);
```

Valid case:

```php
<?php
$tag=$_POST["tag"];if(strlen($tag)!==16){throw new RuntimeException("Bad tag");}$p=openssl_decrypt($c,"aes-256-gcm",$key,OPENSSL_RAW_DATA,$iv,$tag);
```

Equivalent guard spellings are also valid:

```php
<?php
function open($data, $key, $iv) {
    if (16 !== strlen($_POST["tag"])) {
        return null;
    }
    return openssl_decrypt($data, "aes-256-gcm", $key, OPENSSL_RAW_DATA, $iv, $_POST["tag"]);
}
function openOrFalse($data, $key, $iv) {
    $tag = $_POST["tag"];
    return strlen($tag) === 16 ? openssl_decrypt($data, "aes-256-gcm", $key, OPENSSL_RAW_DATA, $iv, $tag) : false;
}
```

## Divergences

Native custos inspection, independently specified; no EA counterpart or EA conformance requirement.
