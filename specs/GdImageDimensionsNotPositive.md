---
id: GdImageDimensionsNotPositive
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "", max: "" }
---

# GdImageDimensionsNotPositive

## Summary

Identify imagecreate or imagecreatetruecolor whose width or height is a known integer at most zero.

## Detection

- D1. Report imagecreate or imagecreatetruecolor whose width or height is a known integer at most zero.
- D2. Resolve builtin functions, classes, constants and methods, including imports and supported named arguments. Follow only bounded local facts within one lexical scope. Receiver and argument provenance must establish the specific API contract; variable names and comments never prove intent.
- D3. Report one finding per violating operation. A reachable path, unchanged value identity and applicable API behavior are required. Intervening mutation, unknown calls involving tracked values, escaped aliases or exhausted analysis budgets invalidate proof.

## Exceptions (no report)

- E1. Exclude unresolved required symbols, user-defined lookalikes, incomplete syntax, unknown required values and unsupported versions. Cases outside D1 receive no finding.
- E2. A dimension read from a local variable is not proven when that variable, or any variable it was assigned from, is changed by `++` or `--`, bound by a `catch` clause or imported by reference into a closure anywhere in the enclosing scope. A counter initialized to zero and incremented in a loop is the typical case.
- E3. No variable dimension is proven in a scope that writes variables by run-time name: `extract()`, `parse_str()`, `$$name++` or a compound assignment to `$$name`.
- E4. A variable shared with other code (top-level code, `global`, `static`, a by-reference parameter or a by-reference closure import of the enclosing closure) is not proven when the scope also runs code that could rewrite it: a call to a user-defined or unresolved function, a method or static call, `new`, `include`/`require`, `eval`, or a closure or arrow function definition.
- E5. Both `@custos-ignore GdImageDimensionsNotPositive` and `@noinspection GdImageDimensionsNotPositive` suppress this native rule.

## Report

- Range: the complete violating call or expression. Follow D1 if it specifies a narrower range.
- Severity: error.
- Enabled by default: true.
- Message: `Supply positive image dimensions.`

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
imagecreatetruecolor(0, 80);
```

Valid case:

```php
<?php
imagecreatetruecolor(80, 80);
```

Valid case, the counter is no longer the literal it started from:

```php
<?php
function canvas(array $lines) {
    $rows = 0;
    foreach ($lines as $line) {
        $rows++;
    }
    return imagecreatetruecolor(200, $rows);
}
```

## Divergences

Native custos inspection, independently specified; no EA counterpart or EA conformance requirement. Conservative proof intentionally excludes unknown runtime state and intent.
