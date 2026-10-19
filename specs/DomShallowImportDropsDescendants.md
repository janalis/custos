---
id: DomShallowImportDropsDescendants
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "", max: "" }
---

# DomShallowImportDropsDescendants

## Summary

Import descendants when copying an element subtree. This inspection is disabled by default; enabling it selects the stated policy.

## Detection

- D1. DOMDocument importNode has omitted or literal false deep argument and the imported element is proven to contain children, by literal loadXML or direct appendChild construction. Enabling establishes subtree-preservation policy. Leaves, attributes and explicit deep imports are excluded.
- D1a. The proof must still hold at the import. Inside the enclosing scope, up to the import (the whole scope when the import sits in a loop), every use of the imported element variable, of the document variable that supplies or holds it, and of each variable appended to it must be one of: a plain rebinding from an object creation or method call; a node-creating or serialising document call (createElement, createElementNS, createTextNode, createComment, createCDATASection, createAttribute, createDocumentFragment, importNode, saveXML, saveHTML); an unconditional loadXML, load, appendChild or removeChild on the document, which D1 already orders; setAttribute, setAttributeNS, getAttribute, hasAttribute, hasChildNodes or appendChild on the element or an appended node; the element as the node of importNode or of an unconditional appendChild on a document variable; an appended variable as the node of appendChild on the element; `$document->documentElement` written directly as the imported node.
- D2. Resolve builtin functions, extension classes, constants and method receivers, including aliases and supported named arguments. Inspect reachable operations in one lexical scope with bounded local flow.
- D3. Emit one finding per violating operation. Never infer policy from names or comments. When the proof is inconclusive, nothing is reported.

## Exceptions (no report)

- E1. Exclude shadowed or unresolved symbols, incomplete code, unknown required constants, ambiguous receiver identity and unsupported versions. Reassignment, escaping aliases, and unknown calls invalidate dependent flow facts.
- E1a. Any other use breaks D1a and drops the report, because the children may be gone: the element, its document or an appended node passed to a function, method or constructor, captured by a closure, copied to another variable or bound by reference; a property read or write on the element; a property of the document other than the directly imported documentElement (so removal through `$document->documentElement->...`); any other method on them, such as remove, removeChild, replaceChild, replaceChildren, loadHTML or loadHTMLFile; a conditional reload of the document; overridden methods of user subclasses; a parenthesised node argument.
- E2. Both @custos-ignore DomShallowImportDropsDescendants and @noinspection DomShallowImportDropsDescendants suppress this native rule.

## Report

- Range: the complete violating expression or call identified by D1; source-to-use findings highlight the final unsafe use, except where D1 specifies the producer.
- Severity: warning.
- Enabled by default: false.
- Message: `Import descendants when copying an element subtree.`

## Fix

Append literal true when deep is omitted, or replace a literal false deep value. Require resolved argument positions, preserved comments and no unpacking.

Use exact byte edits and preserve comments, evaluation count, argument order and supported PHP syntax. Withhold the fix whenever its prerequisites cannot be proven.

## Options

None. Use normal rule configuration to enable or disable the inspection.

## PHP versions

Apply throughout PHP 5.3–8.5 only where the extension API and syntax exist.  Do not flag API unavailability through this rule. Legacy DOM classes are the initial target; modern Dom namespace classes require independently established matching contracts.

## Examples

```php
<?php
$s=new DOMDocument(); $s->loadXML('<list><entry/></list>'); $d=new DOMDocument(); $n=<warning descr="Import descendants when copying an element subtree.">$d->importNode($s->documentElement)</warning>;
```

Fixed result:

```php
<?php
$s=new DOMDocument(); $s->loadXML('<list><entry/></list>'); $d=new DOMDocument(); $n=$d->importNode($s->documentElement, true);
```

Valid case:

```php
<?php
$s=new DOMDocument(); $s->loadXML('<list><entry/></list>'); $d=new DOMDocument(); $n=$d->importNode($s->documentElement, true);
```

## Divergences

Native custos inspection, independently specified. No EA counterpart or EA fixture conformance applies.
