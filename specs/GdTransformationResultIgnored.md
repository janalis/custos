---
id: GdTransformationResultIgnored
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "5.5", max: "" }
---

# GdTransformationResultIgnored

## Summary

Identify discarded imagescale or imagecrop return followed by an image encoder receiving same unchanged source image in straight-line scope. Standalone discard and unknown consumers exclude.

## Detection

- D1. Report discarded imagescale or imagecrop return followed by an image encoder receiving same unchanged source image in straight-line scope. Standalone discard and unknown consumers exclude.
- D2. Resolve builtin functions, classes, constants and methods, including imports and supported named arguments. Follow only bounded local facts within one lexical scope. Receiver and argument provenance must establish the specific API contract; variable names and comments never prove intent.
- D3. Report one finding per violating operation. A reachable path, unchanged value identity and applicable API behavior are required. Intervening mutation, unknown calls involving tracked values, escaped aliases or exhausted analysis budgets invalidate proof.

## Exceptions (no report)

- E1. Exclude unresolved required symbols, user-defined lookalikes, incomplete syntax, unknown required values and unsupported versions. Cases outside D1 receive no finding.
- E2. Both `@custos-ignore GdTransformationResultIgnored` and `@noinspection GdTransformationResultIgnored` suppress this native rule.

## Report

- Range: the complete violating call or expression. Follow D1 if it specifies a narrower range.
- Severity: warning.
- Enabled by default: true.
- Message: `Use the image returned by the transformation.`

## Fix

None. A repair requires a semantic decision or application error-handling policy.

## Options

None.

## PHP versions

Requires PHP 5.5 or later. Do not apply extension contracts when the resolved API is unavailable. Gate syntax and version-specific behavior to the configured target.

## Examples

Finding (enable the rule explicitly when disabled by default):

```php
<?php
$canvas = imagecreatetruecolor(32, 32); imagescale($canvas, 16, 16); imagepng($canvas, $path);
```

Valid case:

```php
<?php
$canvas = imagecreatetruecolor(32, 32); $scaled = imagescale($canvas, 16, 16); if ($scaled !== false) { imagepng($scaled, $path); }
```

## Divergences

Native custos inspection, independently specified; no EA counterpart or EA conformance requirement. Conservative proof intentionally excludes unknown runtime state and intent.
