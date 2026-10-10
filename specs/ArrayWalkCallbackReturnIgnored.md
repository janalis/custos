---
id: ArrayWalkCallbackReturnIgnored
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "", max: "" }
---

# ArrayWalkCallbackReturnIgnored

## Summary

array_walk ignores callback return values. Use array_map to build transformed values, or update elements through a reference parameter.

## Detection

- D1. Resolved array_walk receives a literal arrow function or closure whose only effect is a returned non-null expression and whose first parameter is by value. Return expression must be side-effect-free, allowing resolved pure string transforms. Highlight callback.
- D1e. A supported pure string builtin in the callback requires its argument to have definitely string type. An untyped argument can invoke intentional object-to-string conversion side effects and does not establish a pure discarded transform. Typed string parameters and directly known string expressions can establish the supported argument contract.
- D2. Resolve builtin functions, classes, constants and methods; user-defined lookalikes are excluded. Follow only bounded local proof in one lexical scope.
- D3. Emit one finding per violating operation, not per supporting evidence statement. Use existing analysis budgets; exhausted proof produces no report.

## Exceptions (no report)

- E1. Void callbacks, by-reference mutation, and callbacks with side effects are excluded.
- E2. Both standard suppression forms suppress this native ID.

## Report

- Range: the violating expression or call, unless D1 specifies a narrower range.
- Severity: warning.
- Enabled by default: true.
- Message: Use array_map to retain the callback results.

## Fix

None. Choosing a repair requires intent or additional runtime information.

## Options

None.

## PHP versions

Available throughout supported PHP 5.3–8.5 where the referenced builtin/extension is available. Respect syntax/version-specific receiver contracts; unavailable APIs are not this rule’s concern.

## Examples

```php
<?php
array_walk($names,<warning descr="Use array_map to retain the callback results.">fn(string $name)=>strtoupper($name)</warning>);
```

Valid case:

```php
<?php
$a=["alice"]; array_walk($a,function(&$v){$v=strtoupper($v);});
```

## Divergences

Native custos inspection, independently specified; no EA counterpart or EA conformance requirement.
