---
id: ImagickFrameIndexPastEnd
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "", max: "" }
---

# ImagickFrameIndexPastEnd

## Summary

Identify resolved Imagick::setIteratorIndex whose argument is same unchanged receiver getNumberImages result, or a constant index proven at least known positive frame count. Negative indices may have documented special meaning and are excluded.

## Detection

- D1. Report resolved Imagick::setIteratorIndex whose argument is same unchanged receiver getNumberImages result, or a constant index proven at least known positive frame count. Negative indices may have documented special meaning and are excluded.
- D2. Resolve builtin functions, classes, constants and methods, including imports and supported named arguments. Follow only bounded local facts within one lexical scope. Receiver and argument provenance must establish the specific API contract; variable names and comments never prove intent.
- D1a. When the argument is a local variable, it counts as the unchanged count only if the latest binding of that variable before the call is a plain `=` assignment of the `getNumberImages()` result, or of another variable that satisfies the same condition. Any other earlier rebind in between (`$n--`, `$n++`, a compound assignment, a `foreach` key or value target, `list()`/array destructuring, a reference assignment, `global`/`static`, a `catch` variable, `unset`) means the variable may hold a valid index, so nothing is reported.
- D3. Report one finding per violating operation. A reachable path, unchanged value identity and applicable API behavior are required. Intervening mutation, unknown calls involving tracked values, escaped aliases or exhausted analysis budgets invalidate proof.

## Exceptions (no report)

- E1. Exclude unresolved required symbols, user-defined lookalikes, incomplete syntax, unknown required values and unsupported versions. Cases outside D1 receive no finding.
- E2. Both `@custos-ignore ImagickFrameIndexPastEnd` and `@noinspection ImagickFrameIndexPastEnd` suppress this native rule.

## Report

- Range: the complete violating call or expression. Follow D1 if it specifies a narrower range.
- Severity: error.
- Enabled by default: true.
- Message: `Select an index below the image count.`

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
$image = new Imagick($path); $image->setIteratorIndex($image->getNumberImages());
```

Valid case:

```php
<?php
$image = new Imagick($path); $image->setIteratorIndex(0);
```

Valid case, the count is stepped down before it is used as an index:

```php
<?php
$frames = new Imagick($path);
$n = $frames->getNumberImages();
while ($n-- > 0) {
    $frames->setIteratorIndex($n);
}
```

## Divergences

Native custos inspection, independently specified; no EA counterpart or EA conformance requirement. Conservative proof intentionally excludes unknown runtime state and intent.
