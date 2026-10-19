---
id: UploadClientMimeTrusted
group: Security
kind: semantic
needs: [names, types, flow]
php: { min: "", max: "" }
---

# UploadClientMimeTrusted

## Summary

Detect acceptance of an upload based solely on $_FILES entry type compared to an allowed MIME string without server-side content inspection.

## Detection

- D1. Report a builtin `move_uploaded_file()` call whose source is `$_FILES[key]['tmp_name']` (literal key) when the call is reached only if the client-supplied `$_FILES[key]['type']` of the same entry matched an allowed MIME string, and no server-side content inspection of that temporary file is found (E1).
- D1a. A client type test is `$_FILES[key]['type']` compared (`==`, `===`, either operand order) with a string literal containing `/`, or a builtin `in_array()` call whose needle is that entry type and whose haystack is a non-empty literal array of such strings (strict or loose).
- D1b. Accepting branch: the call is in the then-body of an `if` or `elseif` whose condition implies a positive test. Parentheses are transparent, every operand of `&&` holds in the body, and `!` inverts a negative test (`!=`, `!==`, or a `||` containing one).
- D1c. Guard clause: an earlier statement in the same statement list as the call, or as one of its enclosing statements up to the function or script boundary, is an `if` whose body always leaves (return, throw, exit, break, continue) and whose condition is true whenever the test fails: a `!=`/`!==` comparison, a negated positive test, or a `||` with such an operand.
- D1d. Alternative after a mismatch: the call is in an `elseif` body or the `else` of an `if` chain, and the `if` condition or an earlier `elseif` condition of that chain is true whenever the test fails (as in D1c). `if ($_FILES[key]['type'] !== 'image/png') { … } else { move_uploaded_file(…); }` accepts on the label like the positive form.
- D2. Resolve builtin calls and named arguments before checking the contract. Project-local callees are considered only when their bodies or immutable summaries establish the required facts. The default is enabled, with positive proof required.
- D3. For flow-dependent conditions, require a reachable source-to-use path with no intervening invalidation. Unknown calls, ambiguous aliases, recursion or analysis-budget exhaustion cannot establish definite facts.

## Exceptions (no report)

- E1. Server-side content inspection: anywhere earlier in the same function (or script) the same temporary file, written as `$_FILES[key]['tmp_name']` or as a local variable assigned from it, is an argument of builtin `finfo_file()`, `mime_content_type()`, `getimagesize()`, `exif_imagetype()`, an `imagecreatefrom*()` decoder, or a method call named `file` (`finfo::file()`; the receiver type need not be proven). The inspection does not have to dominate the move: an acceptance path that may depend on stored bytes is not provably label-only.
- E1a. Client MIME used only for display, logging or choosing a name; a test that rejects the allowed type; a guard that does not leave; a `||` alternative in an accepting branch or a `&&` in a guard (the test is then not required); a different `$_FILES` entry; and upload-error or size checks alone are not client type tests.
- E2. Unresolved targets, incomplete required facts, suppressed findings and PHP versions lacking the referenced API or language feature produce no report. User declarations shadowing builtin names are not builtin contracts.

## Report

- Range: the complete violating call or expression identified by D1; for a declaration-state violation, the offending property access or assignment. Highlight the final unsafe use for source-to-use findings.
- Severity: warning.
- Message: `Inspect uploaded content instead of trusting its client MIME label.`

## Fix

No automatic fix: content types, file size limits and image decoding policy must be selected by the application.

## Options

None. Enable or disable this rule through normal rule configuration.

## PHP versions

Apply only on PHP versions supporting the referenced APIs and syntax (within PHP 5.3–8.5). Builtin behavior and argument contracts follow the configured PHP version. Extension-specific calls require resolved extension API symbols; do not treat same-named application functions as extension calls.

## Examples

```php
<?php
if ($_FILES['file']['type'] === 'image/png') { <warning descr="Inspect uploaded content instead of trusting its client MIME label.">move_uploaded_file($_FILES['file']['tmp_name'], '/srv/uploads/image.png')</warning>; }
function storeAvatar() {
    if (!in_array($_FILES['avatar']['type'], ['image/png', 'image/jpeg'], true)) { return; }
    <warning descr="Inspect uploaded content instead of trusting its client MIME label.">move_uploaded_file($_FILES['avatar']['tmp_name'], '/srv/avatars/new')</warning>;
}
function storeScan() {
    if ($_FILES['scan']['type'] !== 'image/png') { return; }
    if (getimagesize($_FILES['scan']['tmp_name']) === false) { return; } // content inspected: no report
    move_uploaded_file($_FILES['scan']['tmp_name'], '/srv/scans/new');
}
```

## Divergences

Native custos rule. No upstream inspection or fixture comparison applies. These conditions are independently specified and examples are original.
