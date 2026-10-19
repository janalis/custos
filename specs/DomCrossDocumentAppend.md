---
id: DomCrossDocumentAppend
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "", max: "" }
---

# DomCrossDocumentAppend

## Summary

Import the node into the destination document.

## Detection

- D1. appendChild, insertBefore or replaceChild receives a node whose owner document is proven distinct from the destination owner document, from local DOMDocument allocations and createElement/createTextNode provenance. Imported, adopted or cloned nodes with correct owner provenance are excluded.
- D1a. A node held in a variable takes the receiver of the latest preceding `adoptNode()` call as its owner when that call is a statement executed on every path to the operation and adopts that same node, whether or not its result is used. This applies to the destination receiver as well as to the inserted node, so a node adopted by a third document, or adopted away from the destination, is still reported.
- D2. Resolve builtin functions, extension classes, constants and method receivers, including aliases and supported named arguments. Inspect reachable operations in one lexical scope with bounded local flow.
- D3. Emit one finding per violating operation. Never infer policy from names or comments. When the proof is inconclusive, nothing is reported.

## Exceptions (no report)

- E1. Exclude shadowed or unresolved symbols, incomplete code, unknown required constants, ambiguous receiver identity and unsupported versions. Reassignment, escaping aliases, and unknown calls invalidate dependent flow facts.
- E1a. When the latest preceding `adoptNode()` call is conditional, adopts a different or unproven node (it may be an ancestor, whose whole subtree moves), or has a receiver without a proven document, the owner of a node variable is unknown and nothing is reported. Documents, `documentElement` and nodes created inline in the operation are unaffected by adoptions.
- E2. Both @custos-ignore DomCrossDocumentAppend and @noinspection DomCrossDocumentAppend suppress this native rule.

## Report

- Range: the complete violating expression or call identified by D1; source-to-use findings highlight the final unsafe use, except where D1 specifies the producer.
- Severity: error.
- Enabled by default: true.
- Message: `Import the node into the destination document.`

## Fix

None. A safe repair requires choosing values, error handling, or application behavior.

## Options

None. Use normal rule configuration to enable or disable the inspection.

## PHP versions

Apply throughout PHP 5.3–8.5 only where the extension API and syntax exist.  Do not flag API unavailability through this rule. Legacy DOM classes are the initial target; modern Dom namespace classes require independently established matching contracts.

## Examples

```php
<?php
$a=new DOMDocument(); $b=new DOMDocument(); <error descr="Import the node into the destination document.">$a->appendChild($b->createElement('entry'))</error>;
```

Valid case:

```php
<?php
$a=new DOMDocument(); $b=new DOMDocument(); $a->appendChild($a->importNode($b->createElement('entry'), true));
```

Adopted in place, also valid:

```php
<?php
$a=new DOMDocument(); $b=new DOMDocument(); $n=$b->createElement('entry'); $a->adoptNode($n); $a->appendChild($n);
```

## Divergences

Native custos inspection, independently specified. No EA counterpart or EA fixture conformance applies.
