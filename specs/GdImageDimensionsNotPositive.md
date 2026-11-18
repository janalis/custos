---
id: GdImageDimensionsNotPositive
group: Probable bugs
kind: flow
needs: [names, types, flow]
php: { min: "", max: "" }
---

# GdImageDimensionsNotPositive

## Summary

Identify imagecreate or imagecreatetruecolor whose width or height is a known integer at most zero.

## Detection

- D1. Report imagecreate or imagecreatetruecolor whose width or height is a known integer at most zero.
- D2. Resolve builtin functions, classes, constants and methods, including imports and supported named arguments. Follow only bounded local facts within one lexical scope. Receiver and argument provenance must establish the specific API contract; variable names and comments never prove intent.
- D3. Report one finding per violating operation. A reachable path, unchanged value identity and applicable API behavior are required. Intervening mutation, unknown calls involving tracked values, escaped aliases or code too complex to analyze invalidate proof.

## Exceptions (no report)

- E1. Exclude unresolved required symbols, user-defined lookalikes, incomplete syntax, unknown required values and unsupported versions. Cases outside D1 receive no finding.
- E2. A dimension read from a local variable, or from any variable whose value an assignment stored into it (directly or through further assignments), is not proven when, anywhere in the enclosing scope, one of these reaches it: `++` or `--` on it or on one of its elements; a `catch`, `foreach` key or value binding it; a reference to it (a by-reference assignment on either side, a by-reference `foreach` over it, a by-reference array item, a by-reference capture by a nested closure, or an argument, other than the inspected argument itself, that a resolved callee binds by reference or that a callee custos cannot resolve receives); a write naming its target at run time (a variable variable, `extract()`, one-argument `parse_str()` or `mb_parse_str()`, `include`/`require`, `eval`); or a `$GLOBALS` access other than a literal entry naming another variable. A counter initialized to zero and incremented in a loop is the typical case. In top-level code a `$GLOBALS` access anywhere in the file counts too. A read inside an arrow function also checks the scopes the arrow function captures from.
- E3. Variable variables, `extract()`, one-argument `parse_str()`, `include`/`require` and `eval` write variables by run-time name: no variable dimension is proven in a scope holding one.
- E4. A variable shared with other code (top-level code, `global`, `static`, a by-reference parameter, a by-reference capture by the closure that is the scope) is proven only when the scope holds no code able to run application logic: a call that is not a resolved builtin function, a method or static call, `new`, `include`/`require`, `eval`, or a closure or arrow function expression.
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
