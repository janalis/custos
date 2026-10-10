---
id: ImagickSingleBlobDropsRequiredFrames
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "", max: "" }
---

# ImagickSingleBlobDropsRequiredFrames

## Summary

Identify Imagick::getImageBlob only with preserveFrames true and local provenance proving at least two image frames, such as two successful newImage/addImage operations. Filename extensions and comments cannot establish frame count. getImagesBlob and known single-frame images exclude.

## Detection

- D1. Report Imagick::getImageBlob only with preserveFrames true and local provenance proving at least two image frames, such as two successful newImage/addImage operations. Filename extensions and comments cannot establish frame count. getImagesBlob and known single-frame images exclude.
- D2. Resolve builtin functions, classes, constants and methods, including imports and supported named arguments. Follow only bounded local facts within one lexical scope. Receiver and argument provenance must establish the specific API contract; variable names and comments never prove intent.
- D3. Report one finding per violating operation. A reachable path, unchanged value identity and applicable API behavior are required. Intervening mutation, unknown calls involving tracked values, escaped aliases or exhausted analysis budgets invalidate proof.

## Exceptions (no report)

- E1. Exclude unresolved required symbols, user-defined lookalikes, incomplete syntax, unknown required values and unsupported versions. Cases outside D1 receive no finding.
- E2. Both `@custos-ignore ImagickSingleBlobDropsRequiredFrames` and `@noinspection ImagickSingleBlobDropsRequiredFrames` suppress this native rule.

## Report

- Range: the complete violating call or expression. Follow D1 if it specifies a narrower range.
- Severity: warning.
- Enabled by default: false.
- Message: `Export every frame when frame preservation is enabled.`

## Fix

None. A repair requires a semantic decision or application error-handling policy.

## Options

`preserveFrames`: bool, default `false`. Only report when explicitly enabled and local provenance proves at least two loaded image frames. Frame names or comments do not establish this policy.

## PHP versions

Available in PHP 5.3–8.5 where the referenced API exists. Do not apply extension contracts when the resolved API is unavailable. Gate syntax and version-specific behavior to the configured target.

## Examples

Finding (enable this rule and set `preserveFrames` to `true`):

```php
<?php
$frames = new Imagick(); if ($frames->newImage(4, 4, 'green')) { if ($frames->newImage(4, 4, 'yellow')) { echo $frames->getImageBlob(); } }
```

Valid case:

```php
<?php
$frames = new Imagick(); if ($frames->newImage(4, 4, 'green')) { if ($frames->newImage(4, 4, 'yellow')) { echo $frames->getImagesBlob(); } }
```

## Divergences

Native custos inspection, independently specified; no EA counterpart or EA conformance requirement. Conservative proof intentionally excludes unknown runtime state and intent.
