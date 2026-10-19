---
id: DomMissingAttributeCheckedAsNull
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "", max: "" }
---

# DomMissingAttributeCheckedAsNull

## Summary

Use hasAttribute to test whether an attribute exists.

## Detection

- D1. DOMElement getAttribute result is compared strictly with null, directly or via unchanged local binding. Both === and !== are covered. Do not report comparisons to empty string or tests using hasAttribute.
- D2. Resolve builtin functions, extension classes, constants and method receivers, including aliases and supported named arguments. Inspect reachable operations in one lexical scope with bounded local flow.
- D3. Emit one finding per violating operation. Never infer policy from names or comments. When the proof is inconclusive, nothing is reported.

## Exceptions (no report)

- E1. Exclude shadowed or unresolved symbols, incomplete code, unknown required constants, ambiguous receiver identity and unsupported versions. Reassignment, escaping aliases, and unknown calls invalidate dependent flow facts.
- E2. Do not report a call that a nullsafe operator can skip: either the getAttribute call itself uses `?->`, or an earlier property fetch or method call in the same unparenthesised member chain does. Such an expression yields null whenever the nullsafe receiver is null, so the null comparison is meaningful and hasAttribute is not an equivalent test. This also applies when the result reaches the comparison through a local binding.
- E3. Both @custos-ignore DomMissingAttributeCheckedAsNull and @noinspection DomMissingAttributeCheckedAsNull suppress this native rule.

## Report

- Range: the complete violating expression or call identified by D1; source-to-use findings highlight the final unsafe use, except where D1 specifies the producer.
- Severity: warning.
- Enabled by default: true.
- Message: `Use hasAttribute to test whether an attribute exists.`

## Fix

None. A safe repair requires choosing values, error handling, or application behavior.

## Options

None. Use normal rule configuration to enable or disable the inspection.

## PHP versions

Apply throughout PHP 5.3–8.5 only where the extension API and syntax exist.  Do not flag API unavailability through this rule. Legacy DOM classes are the initial target; modern Dom namespace classes require independently established matching contracts.

## Examples

```php
<?php
function f(DOMElement $e) { if (<warning descr="Use hasAttribute to test whether an attribute exists.">$e->getAttribute('key') === null</warning>) {} }
```

Valid case:

```php
<?php
function f(DOMElement $e) { if (!$e->hasAttribute('key')) {} }
```

Nullsafe receiver, where null is a real outcome:

```php
<?php
function g(?DOMElement $e) { return $e?->getAttribute('key') === null; }
```

## Divergences

Native custos inspection, independently specified. No EA counterpart or EA fixture conformance applies.
