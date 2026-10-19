---
id: AbstractClassInstantiation
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "", max: "" }
---

# AbstractClassInstantiation

## Summary

Instantiate a concrete class.

## Detection

- D1. Report a new expression whose resolved class declaration is abstract, including a literal class-string traced to that declaration.
- D1a. `new self` counts as resolved when it is evaluated by the body of a method declared in a named abstract class: the instantiated class is then that enclosing class.
- D2. Resolve builtin functions, classes, constants and methods, including imported aliases and supported named arguments. Follow facts within one lexical scope and use resolved declarations when available; never infer intent from identifiers or comments.
- D3. Emit one finding per violating operation. Flow facts require a reachable path with no intervening mutation, escaped alias or unknown call involving the tracked value. When the proof is inconclusive, nothing is reported.

## Exceptions (no report)

- E1. Dynamic class names, unresolved ancestors and factory calls are excluded.
- E1a. `new static` and `new parent` are never resolved. `new self` stays unresolved inside a closure or arrow function (its class scope can be rebound at run time), inside a trait (it names the using class), inside an anonymous class, in a parameter default or attribute argument, and outside a class method.
- E2. Exclude shadowed builtin symbols, unresolved required types, incomplete syntax and unsupported PHP versions. Both `@custos-ignore AbstractClassInstantiation` and `@noinspection AbstractClassInstantiation` suppress this inspection.

## Report

- Range: the complete violating call/expression unless Detection specifies an attribute, return type, header or iterable expression.
- Severity: error.
- Enabled by default: true.
- Message: `Instantiate a concrete class.`

## Fix

No automatic fix. Choosing a repair requires runtime information, error-handling policy or a semantic decision.

## Options

None. Use ordinary rule configuration to enable or disable this inspection.

## PHP versions

Available throughout PHP 5.3–8.5. Apply only where the referenced language features and builtin/extension APIs exist. API-specific later syntax or behavior is gated as described in Detection.

## Examples

Finding:

```php
<?php
abstract class Work {} <error descr="Instantiate a concrete class.">new Work()</error>;
```

```php
<?php
abstract class Draft {
    public static function blank() { return <error descr="Instantiate a concrete class.">new self()</error>; }
}
```

Valid case:

```php
<?php
abstract class Work {} class ConcreteWork extends Work {} new ConcreteWork();
```

```php
<?php
abstract class Draft {
    public static function blank() { return new static(); }
    public function later() { return function () { return new self(); }; }
}
```

## Divergences

Native custos inspection, independently specified. No EA counterpart or EA conformance requirement applies. Unknown intent and unproven state are deliberately excluded.
