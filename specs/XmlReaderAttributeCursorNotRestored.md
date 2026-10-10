---
id: XmlReaderAttributeCursorNotRestored
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "", max: "" }
---

# XmlReaderAttributeCursorNotRestored

## Summary

Restore the XMLReader element cursor before element operations. This inspection is disabled by default; enabling it selects the stated policy.

## Detection

- D1. After a proven successful moveToAttribute or moveToFirstAttribute, with no moveToElement/read/next, compare nodeType to XMLReader::ELEMENT or perform element-only handling. Initial supported sink is strict equality with ELEMENT. Enabling establishes that this comparison intends to handle the containing element.
- D2. Resolve builtin functions, extension classes, constants and method receivers, including aliases and supported named arguments. Inspect reachable operations in one lexical scope with bounded local flow.
- D3. Emit one finding per violating operation. Never infer policy from names or comments. Unsupported proof and exhausted analysis budgets produce no report.

## Exceptions (no report)

- E1. Exclude shadowed or unresolved symbols, incomplete code, unknown required constants, ambiguous receiver identity and unsupported versions. Reassignment, escaping aliases, and unknown calls invalidate dependent flow facts.
- E2. Both @custos-ignore XmlReaderAttributeCursorNotRestored and @noinspection XmlReaderAttributeCursorNotRestored suppress this native rule.

## Report

- Range: the complete violating expression or call identified by D1; source-to-use findings highlight the final unsafe use, except where D1 specifies the producer.
- Severity: warning.
- Enabled by default: false.
- Message: `Restore the XMLReader element cursor before element operations.`

## Fix

None. A safe repair requires choosing values, error handling, or application behavior.

## Options

None. Use normal rule configuration to enable or disable the inspection.

## PHP versions

Apply throughout PHP 5.3–8.5 only where the extension API and syntax exist.  Do not flag API unavailability through this rule.

## Examples

```php
<?php
function f(XMLReader $r) { if($r->moveToAttribute('key')){ if(<warning descr="Restore the XMLReader element cursor before element operations.">$r->nodeType === XMLReader::ELEMENT</warning>){} } }
```

Valid case:

```php
<?php
function f(XMLReader $r) { if($r->moveToAttribute('key')){ $r->moveToElement(); if($r->nodeType === XMLReader::ELEMENT){} } }
```

## Divergences

Native custos inspection, independently specified. No EA counterpart or EA fixture conformance applies.
