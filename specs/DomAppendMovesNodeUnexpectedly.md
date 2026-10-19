---
id: DomAppendMovesNodeUnexpectedly
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "", max: "" }
---

# DomAppendMovesNodeUnexpectedly

## Summary

Clone the node when preserving its existing parent. This inspection is disabled by default; enabling it selects the stated policy.

## Detection

- D1. The same proven node is appended consecutively to distinct proven parents. Enabling establishes that both parents must retain a copy. An intervening explicit detachment or clone excludes the finding (E3); do not infer this policy from identifiers.
- D2. Resolve builtin functions, extension classes, constants and method receivers, including aliases and supported named arguments. Inspect reachable operations in one lexical scope with bounded local flow.
- D3. Emit one finding per violating operation. Never infer policy from names or comments. When the proof is inconclusive, nothing is reported.

## Exceptions (no report)

- E1. Exclude shadowed or unresolved symbols, incomplete code, unknown required constants, ambiguous receiver identity and unsupported versions. Reassignment, escaping aliases, and unknown calls invalidate dependent flow facts.
- E2. Both @custos-ignore DomAppendMovesNodeUnexpectedly and @noinspection DomAppendMovesNodeUnexpectedly suppress this native rule.
- E3. A call that executes on every path between the two appends counts as an explicit detachment when it is one of:
  - `remove()` or `replaceWith(...)` called on the node itself;
  - `removeChild` or `replaceChild` naming the node, positionally or through the `child`/`node` named arguments, on any receiver;
  - `replaceChildren(...)` called on the first parent;
  - `removeChild` or `replaceChild` called on the first parent with a removed child that is not proven to be another node. The removed child is proven distinct only when it and the appended node both resolve to different local allocation calls (`DOMDocument::createElement`, `createElementNS`, `createTextNode`, `createComment`, `createCDATASection`, `createDocumentFragment`, or `cloneNode`); `$parent->firstChild` and similar navigations are never proven distinct.

  The same calls on unrelated nodes or parents keep the finding, as does removing the first parent itself.

## Report

- Range: the complete violating expression or call identified by D1; source-to-use findings highlight the final unsafe use, except where D1 specifies the producer.
- Severity: warning.
- Enabled by default: false.
- Message: `Clone the node when preserving its existing parent.`

## Fix

None. A safe repair requires choosing values, error handling, or application behavior.

## Options

None. Use normal rule configuration to enable or disable the inspection.

## PHP versions

Apply throughout PHP 5.3–8.5 only where the extension API and syntax exist.  Do not flag API unavailability through this rule. Legacy DOM classes are the initial target; modern Dom namespace classes require independently established matching contracts.

## Examples

```php
<?php
$d=new DOMDocument(); $a=$d->createElement('a'); $b=$d->createElement('b'); $n=$d->createElement('n'); $a->appendChild($n); <warning descr="Clone the node when preserving its existing parent.">$b->appendChild($n)</warning>;
```

Valid case:

```php
<?php
$d=new DOMDocument(); $a=$d->createElement('a'); $b=$d->createElement('b'); $n=$d->createElement('n'); $a->appendChild($n); $b->appendChild($n->cloneNode(true));
```

Detached deliberately, no report:

```php
<?php
$d=new DOMDocument(); $a=$d->createElement('a'); $b=$d->createElement('b'); $n=$d->createElement('n'); $a->appendChild($n); $n->remove(); $b->appendChild($n);
```

## Divergences

Native custos inspection, independently specified. No EA counterpart or EA fixture conformance applies.
