---
id: FailedClosureBindingInvoked
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "5.4", max: "" }
---

# FailedClosureBindingInvoked

## Summary

Check closure binding before invoking it.

## Detection

- D1. Track an explicitly static closure bound using Closure::bindTo or builtin Closure::bind with a proven non-null object. Report invocation of the failed binding result on the same reachable local path, including an immediately invoked result. The object is proven only when every alternative of its inferred type is a class instance: a nullable, untyped, mixed or otherwise partly non-class type proves nothing, because binding null to a static closure succeeds.
- D2. Resolve builtin functions, classes, constants and methods, including imported aliases and supported named arguments. Follow facts within one lexical scope and use resolved declarations when available; never infer intent from identifiers or comments.
- D3. Emit one finding per violating operation. Flow facts require a reachable path with no intervening mutation, escaped alias, unknown call involving the tracked value or analysis-budget exhaustion. Uncertain facts cannot prove a violation.

## Exceptions (no report)

- E1. Null object bindings, nonstatic closures, binding results guarded against null and unknown closure provenance are excluded. Binding failure alone is not reported here.
- E2. Exclude shadowed builtin symbols, unresolved required types, incomplete syntax and unsupported PHP versions. Both `@custos-ignore FailedClosureBindingInvoked` and `@noinspection FailedClosureBindingInvoked` suppress this inspection.
- E3. A result is guarded when the invocation can only run with a non-null value of the same tracked result:
  - it sits in an `if`/`elseif` body whose condition proves it, in an `elseif`/`else` branch reached only after a condition whose failure proves it, in the matching arm of a ternary, or in the right operand of `&&`/`and` (left operand proves it when true) or `||`/`or` (left operand proves it when false);
  - or an earlier statement of the same or an enclosing statement list is an `if` without `elseif`/`else` whose body always leaves the path and whose condition proves it when false.
  - A condition proves the result when true if it is the result itself (truthiness), `!== null`/`!= null` in either operand order, `instanceof`, `isset(...)`, builtin `is_callable(...)`/`is_object(...)`, the negation of a condition that proves it when false, or a conjunction with such an operand. It proves the result when false if it is `=== null`/`== null`, `empty(...)`, builtin `is_null(...)`, the negation of a condition that proves it when true, or a disjunction with such an operand. Any other condition, including comparisons with other values, proves nothing.
  - A result held in a plain variable is also guarded when the flow facts of a followed scope exclude null at the invocation, for instance a result assigned inside the condition that tests it, a test stored in a boolean flag, or a loop condition. A result held in a property or static property, or invoked in a scope the flow does not follow (`goto`, generator), relies on the forms above only.

## Report

- Range: the complete violating call/expression unless Detection specifies an attribute, return type, header or iterable expression.
- Severity: error.
- Enabled by default: true.
- Message: `Check closure binding before invoking it.`

## Fix

No automatic fix. Choosing a repair requires runtime information, error-handling policy or a semantic decision.

All edits use exact byte ranges, preserve comments and evaluation order, and require syntax supported by the target PHP version. A fix must never add an evaluation or silently change unrelated arguments.

## Options

None. Use ordinary rule configuration to enable or disable this inspection.

## PHP versions

Minimum PHP 5.4. Apply only where the referenced language features and builtin/extension APIs exist. API-specific later syntax or behavior is gated as described in Detection.

## Examples

Finding:

```php
<?php
$f = static function () {}; $bound = $f->bindTo(new stdClass()); <error descr="Check closure binding before invoking it.">$bound()</error>;
```

Valid case:

```php
<?php
$f = static function () {}; $bound = $f->bindTo(null); if ($bound !== null) { $bound(); }
```

Guarded results are not reported either:

```php
<?php
function run(stdClass $target) {
    $f = static function () {};
    $bound = $f->bindTo($target);
    if (!$bound) {
        return null;
    }
    return $bound();
}
```

## Divergences

Native custos inspection, independently specified. No EA counterpart or EA conformance requirement applies. Unknown intent and unproven state are deliberately excluded.
