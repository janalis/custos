---
id: XPathQueryFailureDereferenced
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "", max: "" }
---

# XPathQueryFailureDereferenced

## Summary

Check the XPath query result before accessing its nodes.

## Detection

- D1. DOMXPath query result directly or via local binding is dereferenced as DOMNodeList without a dominating guard excluding false. Guards may throw, return, or enter a proven nonfalse branch. Passing to unknown callees is not itself a finding.
- D2. Resolve builtin functions, extension classes, constants and method receivers, including aliases and supported named arguments. Inspect reachable operations in one lexical scope with bounded local flow.
- D3. Emit one finding per violating operation. Never infer policy from names or comments. Unsupported proof and exhausted analysis budgets produce no report.

## Exceptions (no report)

- E1. Exclude shadowed or unresolved symbols, incomplete code, unknown required constants, ambiguous receiver identity and unsupported versions. Reassignment, escaping aliases, and unknown calls invalidate dependent flow facts.
- E2. Both @custos-ignore XPathQueryFailureDereferenced and @noinspection XPathQueryFailureDereferenced suppress this native rule.

## Report

- Range: the complete violating expression or call identified by D1; source-to-use findings highlight the final unsafe use, except where D1 specifies the producer.
- Severity: warning.
- Enabled by default: true.
- Message: `Check the XPath query result before accessing its nodes.`

## Fix

None. A safe repair requires choosing values, error handling, or application behavior.

## Options

None. Use normal rule configuration to enable or disable the inspection.

## PHP versions

Apply throughout PHP 5.3–8.5 only where the extension API and syntax exist.  Do not flag API unavailability through this rule. Legacy DOM classes are the initial target; modern Dom namespace classes require independently established matching contracts.

## Examples

```php
<?php
function f(DOMXPath $x, $expr) { $nodes=$x->query($expr); echo <warning descr="Check the XPath query result before accessing its nodes.">$nodes->length</warning>; }
```

Valid case:

```php
<?php
function f(DOMXPath $x, $expr) { $nodes=$x->query($expr); if($nodes!==false){echo $nodes->length;} }
```

## Divergences

Native custos inspection, independently specified. No EA counterpart or EA fixture conformance applies.
