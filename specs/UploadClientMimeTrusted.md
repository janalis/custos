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

- D1. Report acceptance of an upload based solely on $_FILES entry type compared to an allowed MIME string without server-side content inspection.
- D2. Resolve builtin calls and named arguments before checking the contract. Project-local callees are considered only when their bodies or immutable summaries establish the required facts. The default is enabled, with positive proof required.
- D3. For flow-dependent conditions, require a reachable source-to-use path with no intervening invalidation. Unknown calls, ambiguous aliases, recursion or analysis-budget exhaustion cannot establish definite facts.

## Exceptions (no report)

- E1. Client MIME used only for display/logging, server-side finfo or decoder validation, upload-error rejection and unknown acceptance paths are excluded.
- E2. Unresolved targets, incomplete required facts, suppressed findings and PHP versions lacking the referenced API or language feature produce no report. User declarations shadowing builtin names are not builtin contracts.

## Report

- Range: the complete violating call or expression identified by D1; for a declaration-state violation, the offending property access or assignment. Highlight the final unsafe use for source-to-use findings.
- Severity: warning.
- Message: `Inspect uploaded content instead of trusting its client MIME label.`

## Fix

No automatic fix: content types, file size limits and image decoding policy must be selected by the application.

A fix is offered only when the concrete prerequisites above are proven. Edits use exact source byte ranges, preserve comments and argument evaluation order, and introduce syntax supported by the configured PHP version. Never invent error-handling policy.

## Options

None. Enable or disable this rule through normal rule configuration.

## PHP versions

Apply only on PHP versions supporting the referenced APIs and syntax (within PHP 5.3–8.5). Builtin behavior and argument contracts follow the configured PHP version. Extension-specific calls require resolved extension API symbols; do not treat same-named application functions as extension calls.

## Examples

```php
<?php
if ($_FILES['file']['type'] === 'image/png') { <warning descr="Inspect uploaded content instead of trusting its client MIME label.">move_uploaded_file($_FILES['file']['tmp_name'], '/srv/uploads/image.png')</warning>; }
```

## Divergences

Native custos rule. No upstream inspection or fixture comparison applies. These conditions are independently specified and examples are original.
