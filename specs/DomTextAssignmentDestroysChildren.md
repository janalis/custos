---
id: DomTextAssignmentDestroysChildren
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "", max: "" }
---

# DomTextAssignmentDestroysChildren

## Summary

Preserve child nodes when adding element text. This inspection is disabled by default; enabling it selects the stated policy.

## Detection

- D1. Assigning nodeValue or textContent on a proven DOMElement with existing element children is reported. Prove children through direct construction or literal loadXML. Enabling establishes child-preservation policy; text-only nodes and unknown children are excluded.
- D2. Resolve builtin functions, extension classes, constants and method receivers, including aliases and supported named arguments. Inspect reachable operations in one lexical scope with bounded local flow.
- D3. Emit one finding per violating operation. Never infer policy from names or comments. Unsupported proof and exhausted analysis budgets produce no report.

## Exceptions (no report)

- E1. Exclude shadowed or unresolved symbols, incomplete code, unknown required constants, ambiguous receiver identity and unsupported versions. Reassignment, escaping aliases, and unknown calls invalidate dependent flow facts.
- E2. Both @custos-ignore DomTextAssignmentDestroysChildren and @noinspection DomTextAssignmentDestroysChildren suppress this native rule.

## Report

- Range: the complete violating expression or call identified by D1; source-to-use findings highlight the final unsafe use, except where D1 specifies the producer.
- Severity: warning.
- Enabled by default: false.
- Message: `Preserve child nodes when adding element text.`

## Fix

None. A safe repair requires choosing values, error handling, or application behavior.

## Options

None. Use normal rule configuration to enable or disable the inspection.

## PHP versions

Apply throughout PHP 5.3–8.5 only where the extension API and syntax exist.  Do not flag API unavailability through this rule. Legacy DOM classes are the initial target; modern Dom namespace classes require independently established matching contracts.

## Examples

```php
<?php
$d=new DOMDocument(); $d->loadXML('<box><part/></box>'); <warning descr="Preserve child nodes when adding element text.">$d->documentElement->nodeValue='added'</warning>;
```

Valid case:

```php
<?php
$d=new DOMDocument(); $d->loadXML('<box><part/></box>'); $d->documentElement->appendChild($d->createTextNode('added'));
```

## Divergences

Native custos inspection, independently specified. No EA counterpart or EA fixture conformance applies.
