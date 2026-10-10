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
- D2. Resolve builtin functions, extension classes, constants and method receivers, including aliases and supported named arguments. Inspect reachable operations in one lexical scope with bounded local flow.
- D3. Emit one finding per violating operation. Never infer policy from names or comments. Unsupported proof and exhausted analysis budgets produce no report.

## Exceptions (no report)

- E1. Exclude shadowed or unresolved symbols, incomplete code, unknown required constants, ambiguous receiver identity and unsupported versions. Reassignment, escaping aliases, and unknown calls invalidate dependent flow facts.
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
