---
id: CompressionLevelOutsideSupportedRange
group: Probable bugs
kind: flow
needs: [names, types, flow]
php: { min: "", max: "" }
---

# CompressionLevelOutsideSupportedRange

## Summary

Use a supported compression level. Enabled by default; require positive semantic or local-flow proof.

## Detection

- D1. Report gzencode, gzcompress or gzdeflate level argument proven integer less than -1 or greater than 9. Skip unknown levels, default omission, and bzip2 APIs with different contracts.
- D2. Resolve builtin functions, extension classes, method receivers, constants and named arguments. Follow unchanged locals and aliases only within the same bounded lexical scope. Require a reachable path; merge branch facts only when all incoming paths establish them.
- D3. Emit one finding at the violating operation, not one per evidence statement. Rebinding, writes, by-reference escape and unresolved calls that can mutate relevant state remove proof. When the proof is inconclusive, nothing is reported.
- D4. A level passed as a variable is proven only when that variable keeps its literal for the whole enclosing scope (function, method, closure or top-level code; nested function, closure and class bodies are not part of it). The same requirement applies to every variable whose value an assignment stores into it, directly or through further assignments.

## Exceptions (no report)

- E1. Exclude shadowed builtins, unresolved receivers, incomplete syntax and unknown prerequisite facts. An extension function name alone does not prove an application function is that builtin.
- E2. A dominating valid guard or repair described in Detection prevents the finding. Both legacy noinspection and custos-ignore forms suppress this ID.
- E3. No report for a variable level when, anywhere in the enclosing scope and wherever it happens relative to the call, one of these reaches the variable: `++` or `--` on it or on one of its elements; a `catch`, `foreach` key or value binding it; a reference to it (a by-reference assignment on either side, a by-reference `foreach` over it, a by-reference array item, a by-reference capture by a nested closure, or an argument, other than the inspected argument itself, that a resolved callee binds by reference or that a callee custos cannot resolve receives); a write naming its target at run time (a variable variable, `extract()`, one-argument `parse_str()` or `mb_parse_str()`, `include`/`require`, `eval`); or a `$GLOBALS` access other than a literal entry naming another variable. An arrow function or a by-value import only works on a copy and keeps the proof. In top-level code a `$GLOBALS` access anywhere in the file counts too. A variable shared with other code (top-level code, `global`, `static`, a by-reference parameter, a by-reference capture by the closure that is the scope) is proven only when the scope holds no code able to run application logic: a call that is not a resolved builtin function, a method or static call, `new`, `include`/`require`, `eval`, or a closure or arrow function expression. A read inside an arrow function also checks the scopes the arrow function captures from.

## Report

- Range: the violating expression or call described in D1; stateful absence checks highlight the final unsafe operation or loop.
- Severity: error.
- Enabled by default: true.
- Message: `Use a supported compression level.`

## Fix

No automatic fix. The required repair depends on error handling, data selection, lifecycle ordering or application policy.

## Options

None.

## PHP versions

Requires PHP 5.3 or later and the referenced extension API. Respect method-specific introduction versions within PHP 5.3–8.5; unsupported optional APIs cannot establish facts. Named arguments require PHP 8.0 or later.

## Examples

```php
<?php
<error descr="Use a supported compression level.">gzencode("sample",14)</error>;
```

Valid case:

```php
<?php
gzencode("sample",9);
```

Changed before the call, so the level is 9:

```php
<?php
function pack_payload($payload) {
    $level = 10;
    $level--;
    return gzencode($payload, $level);
}
```

## Divergences

Native custos inspection, independently specified; no EA counterpart or EA conformance requirement.
