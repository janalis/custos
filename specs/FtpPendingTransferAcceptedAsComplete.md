---
id: FtpPendingTransferAcceptedAsComplete
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "", max: "" }
---

# FtpPendingTransferAcceptedAsComplete

## Summary

Identify boolean test of ftp_nb_get/put/fget/fput/continue in a branch returning true or emitting constant completion text. Require literal complete output or explicit success return; variable names cannot establish intent.

## Detection

- D1. Report boolean test of ftp_nb_get/put/fget/fput/continue in a branch returning true or emitting constant completion text. Require literal complete output or explicit success return; variable names cannot establish intent.
- D1a. The tested branch is the one that runs when the returned status is non-zero, so `FTP_MOREDATA` reaches it: the body of `if (call)`, or the `else` of `if (!call)`. Parentheses, a plain assignment of the status, `!` and `==`/`!=`/`===`/`!==` against a boolean literal are followed with their polarity. The status is an integer, so a strict comparison of the status itself with a boolean has a fixed outcome: `!==` always holds (the body is the branch), `===` never does (the `else` is).
- D2. Resolve builtin functions, classes, constants and methods, including imports and supported named arguments. Follow only bounded local facts within one lexical scope. Receiver and argument provenance must establish the specific API contract; variable names and comments never prove intent.
- D3. Report one finding per violating operation. A reachable path, unchanged value identity and applicable API behavior are required. Intervening mutation, unknown calls involving tracked values, escaped aliases or exhausted analysis budgets invalidate proof.

## Exceptions (no report)

- E1. Exclude unresolved required symbols, user-defined lookalikes, incomplete syntax, unknown required values and unsupported versions. Cases outside D1 receive no finding.
- E1a. The branch that runs for a zero status only (`FTP_FAILED`), such as the body of `if (!call)` or `if (call == false)`, is a failure path and is not reported whatever it returns.
- E1b. A branch that refers to the global constant `FTP_FINISHED` anywhere outside nested functions and classes checks or drives the transfer to completion itself (a guard on the assigned status, a polling loop on `ftp_nb_continue`); it is not reported.
- E1c. Conditions of loops, compound or by-reference assignments, and comparisons with anything but a boolean literal are not followed. First-class callable syntax `ftp_nb_get(...)` starts no transfer and is not reported.
- E2. Both `@custos-ignore FtpPendingTransferAcceptedAsComplete` and `@noinspection FtpPendingTransferAcceptedAsComplete` suppress this native rule.

## Report

- Range: the complete violating call or expression. Follow D1 if it specifies a narrower range.
- Severity: warning.
- Enabled by default: true.
- Message: `Wait for FTP_FINISHED before reporting success.`

## Fix

None. A repair requires a semantic decision or application error-handling policy.

## Options

None.

## PHP versions

Available in PHP 5.3–8.5 where the referenced API exists. Do not apply extension contracts when the resolved API is unavailable. Gate syntax and version-specific behavior to the configured target.

## Examples

Finding (enable the rule explicitly when disabled by default):

```php
<?php
if (ftp_nb_get($ftp, $local, $remote, FTP_BINARY)) { return true; }
```

Valid case:

```php
<?php
if (ftp_nb_get($ftp, $local, $remote, FTP_BINARY) === FTP_FINISHED) { return true; }
```

Valid cases, failure branch and completion handled inside the branch:

```php
<?php
function failed($ftp, $local, $remote): bool {
    if (!ftp_nb_get($ftp, $local, $remote, FTP_BINARY)) { return true; }
    return false;
}
function fetch($ftp, $local, $remote): bool {
    if ($status = ftp_nb_get($ftp, $local, $remote, FTP_BINARY)) {
        while ($status === FTP_MOREDATA) { $status = ftp_nb_continue($ftp); }
        if ($status !== FTP_FINISHED) { return false; }
        return true;
    }
    return false;
}
```

## Divergences

Native custos inspection, independently specified; no EA counterpart or EA conformance requirement. Conservative proof intentionally excludes unknown runtime state and intent.
