---
id: GdPaletteIndexDecodedAsRgb
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "", max: "" }
---

# GdPaletteIndexDecodedAsRgb

## Summary

Identify a color id from imagecolorallocate on a proven imagecreate palette image used in packed RGB extraction with rightshift8/16 and mask255. imagecreatetruecolor, imagepalettetotruecolor and imagecolorsforindex exclude.

## Detection

- D1. Report a color id from imagecolorallocate on a proven imagecreate palette image used in packed RGB extraction with rightshift8/16 and mask255. imagecreatetruecolor, imagepalettetotruecolor and imagecolorsforindex exclude.
- D2. Resolve builtin functions, classes, constants and methods, including imports and supported named arguments. Follow only bounded local facts within one lexical scope. Receiver and argument provenance must establish the specific API contract; variable names and comments never prove intent.
- D3. Report one finding per violating operation. A reachable path, unchanged value identity and applicable API behavior are required. Intervening mutation, unknown calls involving tracked values, escaped aliases or exhausted analysis budgets invalidate proof.

## Exceptions (no report)

- E1. Exclude unresolved required symbols, user-defined lookalikes, incomplete syntax, unknown required values and unsupported versions. Cases outside D1 receive no finding.
- E2. Any `imagepalettetotruecolor` call on the same image located between its creation and the allocation removes the palette proof, even when the call sits in a branch or loop and does not run on every path: the image may be truecolor there and the shift is then valid. A conversion of another image, one preceding a new `imagecreate` assignment, or one following the allocation (the id is already a palette index) does not exclude.
- E2a. The index and its image must each be the one value every path to the shift agrees on. When another assignment reaches the shift or the allocation on some path (a packed colour, an index allocated on another image, a truecolor image), or a loop replaces either after the shift, nothing is reported, whichever assignment is written first: the shift may be valid there.
- E3. Both `@custos-ignore GdPaletteIndexDecodedAsRgb` and `@noinspection GdPaletteIndexDecodedAsRgb` suppress this native rule.

## Report

- Range: the complete violating call or expression. Follow D1 if it specifies a narrower range.
- Severity: warning.
- Enabled by default: true.
- Message: `Look up palette components instead of shifting the index.`

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
$image = imagecreate(3, 3); $c = imagecolorallocate($image, 255, 0, 0); $red = ($c >> 16) & 255;
```

Valid case:

```php
<?php
$image = imagecreatetruecolor(3, 3); $c = imagecolorallocate($image, 255, 0, 0); $red = ($c >> 16) & 255;
```

## Divergences

Native custos inspection, independently specified; no EA counterpart or EA conformance requirement. Conservative proof intentionally excludes unknown runtime state and intent.
