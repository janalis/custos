---
id: DomLiveNodeListRemovalSkipsNodes
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "", max: "" }
---

# DomLiveNodeListRemovalSkipsNodes

## Summary

Remove live DOM matches without skipping shifted nodes.

## Detection

- D1. A forward for loop starts index at zero, tests index against a live getElementsByTagName/getElementsByTagNameNS result length, increments by one, and unconditionally removes item(index) from its parent. Exclude backwards loops, snapshots, index correction and removals outside this exact pattern.
- D2. Resolve builtin functions, extension classes, constants and method receivers, including aliases and supported named arguments. Inspect reachable operations in one lexical scope with bounded local flow.
- D3. Emit one finding per violating operation. Never infer policy from names or comments. Unsupported proof and exhausted analysis budgets produce no report.

## Exceptions (no report)

- E1. Exclude shadowed or unresolved symbols, incomplete code, unknown required constants, ambiguous receiver identity and unsupported versions. Reassignment, escaping aliases, and unknown calls invalidate dependent flow facts.
- E2. Both @custos-ignore DomLiveNodeListRemovalSkipsNodes and @noinspection DomLiveNodeListRemovalSkipsNodes suppress this native rule.

## Report

- Range: the complete violating expression or call identified by D1; source-to-use findings highlight the final unsafe use, except where D1 specifies the producer.
- Severity: warning.
- Enabled by default: true.
- Message: `Remove live DOM matches without skipping shifted nodes.`

## Fix

None. A safe repair requires choosing values, error handling, or application behavior.

## Options

None. Use normal rule configuration to enable or disable the inspection.

## PHP versions

Apply throughout PHP 5.3–8.5 only where the extension API and syntax exist.  Do not flag API unavailability through this rule. Legacy DOM classes are the initial target; modern Dom namespace classes require independently established matching contracts.

## Examples

```php
<?php
function f(DOMDocument $d) { $nodes=$d->getElementsByTagName('part'); for($i=0;$i<$nodes->length;$i++){ $n=$nodes->item($i); <warning descr="Remove live DOM matches without skipping shifted nodes.">$n->parentNode->removeChild($n)</warning>; } }
```

Valid case:

```php
<?php
function f(DOMDocument $d) { $nodes=$d->getElementsByTagName('part'); for($i=$nodes->length-1;$i>=0;$i--){ $n=$nodes->item($i); $n->parentNode->removeChild($n); } }
```

## Divergences

Native custos inspection, independently specified. No EA counterpart or EA fixture conformance applies.
