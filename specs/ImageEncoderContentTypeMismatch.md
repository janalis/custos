---
id: ImageEncoderContentTypeMismatch
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "", max: "" }
---

# ImageEncoderContentTypeMismatch

## Summary

The response content type must describe the image bytes being written. Match it to the selected image encoder.

## Detection

- D1. Resolved header writes a literal Content-Type of image/png, image/jpeg, image/gif, image/webp, image/avif or image/bmp, followed in straight-line scope by a different corresponding image encoder outputting to response (absent/null output path). No intervening content-type header.
- D1d. The response-output encoder statement must be the final statement in its lexical block. Account for preceding corrective Content-Type headers and removals; unknown response operations discard the MIME proof. Later statements, corrective headers or unknown calls prevent a finding rather than assuming the earlier MIME header remains final.
- D1e. Restrict the initial proof to an encoder that is the final top-level statement in the file. Finality inside a function or nested block cannot establish that later execution leaves the MIME header unchanged, so those contexts are excluded.
- D1f. Exclude earlier ob_start calls with any explicitly supplied callback: that callback can transform the eventual output bytes. Scan at most 4096 prior syntax nodes; exhausted scan cannot establish the response MIME proof.
- D2. Resolve builtin functions, classes, constants and methods; user-defined lookalikes are excluded. Follow only bounded local proof in one lexical scope.
- D3. Emit one finding per violating operation, not per supporting evidence statement. Use existing analysis budgets; exhausted proof produces no report.

## Exceptions (no report)

- E1. Do not report unresolved or shadowed builtin names, unknown values, or incomplete receiver/type provenance. Unknown calls, aliases escaping, and intervening writes invalidate local proof. Similar code outside the stated detection contract is not a finding.
- E2. Both standard suppression forms suppress this native ID.

## Report

- Range: the violating expression or call, unless D1 specifies a narrower range.
- Severity: warning.
- Enabled by default: true.
- Message: Match the response content type to the image encoder.

## Fix

None. Choosing a repair requires intent or additional runtime information.

## Options

None.

## PHP versions

Available throughout supported PHP 5.3–8.5 where the referenced builtin/extension is available. Respect syntax/version-specific receiver contracts; unavailable APIs are not this rule’s concern.

## Examples

```php
<?php
header("Content-Type: image/jpeg"); <warning descr="Match the response content type to the image encoder.">imagepng($im)</warning>;
```

Valid case:

```php
<?php
header("Content-Type: image/png"); imagepng($im);
```

## Divergences

Native custos inspection, independently specified; no EA counterpart or EA conformance requirement.
