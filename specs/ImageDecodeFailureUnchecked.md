---
id: ImageDecodeFailureUnchecked
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "", max: "" }
---

# ImageDecodeFailureUnchecked

## Summary

Decoding image bytes can fail and return false. Check the result before using it as an image.

## Detection

- D1. A local directly assigned resolved imagecreatefromstring is consumed as first argument of imagepng/imagejpeg/imagegif/imagesx/imagesy without a false-result guard. Require unknown input or known invalid literal bytes; recognize early-return/throw guard.
- D2. Resolve builtin functions, classes, constants and methods; user-defined lookalikes are excluded. Follow only bounded local proof in one lexical scope.
- D3. Emit one finding per violating operation, not per supporting evidence statement. Use existing analysis budgets; exhausted proof produces no report.

## Exceptions (no report)

- E1. Do not report unresolved or shadowed builtin names, unknown values, or incomplete receiver/type provenance. Unknown calls, aliases escaping, and intervening writes invalidate local proof. Similar code outside the stated detection contract is not a finding.
- E2. Both standard suppression forms suppress this native ID.

## Report

- Range: the violating expression or call, unless D1 specifies a narrower range.
- Severity: warning.
- Enabled by default: true.
- Message: Check image decoding success before using the image.

## Fix

None. Choosing a repair requires intent or additional runtime information.

## Options

None.

## PHP versions

Available throughout supported PHP 5.3–8.5 where the referenced builtin/extension is available. Respect syntax/version-specific receiver contracts; unavailable APIs are not this rule’s concern.

## Examples

```php
<?php
$im=imagecreatefromstring($bytes); imagepng(<warning descr="Check image decoding success before using the image.">$im</warning>);
```

Valid case:

```php
<?php
$im=imagecreatefromstring($bytes); if($im===false){throw new RuntimeException("Invalid image");} imagepng($im);
```

## Divergences

Native custos inspection, independently specified; no EA counterpart or EA conformance requirement.
