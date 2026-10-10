---
id: ImagickResizeStatusUsedAsImage
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "", max: "" }
---

# ImagickResizeStatusUsedAsImage

## Summary

Identify method/property access on unchanged result of resolved Imagick::resizeImage; it returns a boolean, not a replacement image.

## Detection

- D1. Report method/property access on unchanged result of resolved Imagick::resizeImage; it returns a boolean, not a replacement image.
- D2. Resolve builtin functions, classes, constants and methods, including imports and supported named arguments. Follow only bounded local facts within one lexical scope. Receiver and argument provenance must establish the specific API contract; variable names and comments never prove intent.
- D3. Report one finding per violating operation. A reachable path, unchanged value identity and applicable API behavior are required. Intervening mutation, unknown calls involving tracked values, escaped aliases or exhausted analysis budgets invalidate proof.

## Exceptions (no report)

- E1. Exclude unresolved required symbols, user-defined lookalikes, incomplete syntax, unknown required values and unsupported versions. Cases outside D1 receive no finding.
- E2. Both `@custos-ignore ImagickResizeStatusUsedAsImage` and `@noinspection ImagickResizeStatusUsedAsImage` suppress this native rule.

## Report

- Range: the complete violating call or expression. Follow D1 if it specifies a narrower range.
- Severity: warning.
- Enabled by default: true.
- Message: `Use the mutated Imagick object after resizing.`

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
$image = new Imagick(); $small = $image->resizeImage(80, 80, Imagick::FILTER_LANCZOS, 1); $small->getImageBlob();
```

Valid case:

```php
<?php
$image = new Imagick(); $image->resizeImage(80, 80, Imagick::FILTER_LANCZOS, 1); $image->getImageBlob();
```

## Divergences

Native custos inspection, independently specified; no EA counterpart or EA conformance requirement. Conservative proof intentionally excludes unknown runtime state and intent.
