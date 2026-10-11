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
- D1a. The result is either false or a node list, so each of these tests on the bound variable separates the two: a strict or loose comparison with the `false` constant (either operand order), plain truthiness (`$nodes`, `!$nodes`), and `instanceof` any class. A test may also be applied to the assignment that binds the variable, as in `($nodes = $x->query($e)) === false`. Tests combine through `!`, `&&` and `||` with their usual meaning: a conjunction holding proves each operand held, a disjunction failing proves each operand failed.
- D1b. A test guards the dereference when its outcome is already decided on every path to it: the dereference sits in the matching branch of an `if` / `elseif` / `else` chain (an earlier condition of the chain having failed counts), in the matching arm of a full ternary, in the right operand of `&&` / `||` whose left operand decides it, or after an earlier statement of the same or an enclosing statement list that is an `if` without alternatives whose body always leaves the path when the result may be false. Any number of statements may separate that exit from the dereference.
- D1c. The tested variable must hold the same reaching query result as the dereferenced one; a reassignment or invalidation between test and dereference discards the guard.
- D2. Resolve builtin functions, extension classes, constants and method receivers, including aliases and supported named arguments. Inspect reachable operations in one lexical scope with bounded local flow.
- D3. Emit one finding per violating operation. Never infer policy from names or comments. Unsupported proof and exhausted analysis budgets produce no report.

## Exceptions (no report)

- E1. Exclude shadowed or unresolved symbols, incomplete code, unknown required constants, ambiguous receiver identity and unsupported versions. Reassignment, escaping aliases, and unknown calls invalidate dependent flow facts.
- E2. A dereference in the failure branch of a test (`if (!$nodes) { $nodes->length; }`), in the short ternary form, or behind a test of another value stays reported. The `and` / `or` keywords and tests other than those in D1a are not recognised as guards.
- E3. Both @custos-ignore XPathQueryFailureDereferenced and @noinspection XPathQueryFailureDereferenced suppress this native rule.

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

Also valid, an early exit on a falsy result followed by unrelated statements:

```php
<?php
function g(DOMXPath $x, $expr) { $nodes=$x->query($expr); if(!$nodes){return 0;} $extra=1; return $nodes->length+$extra; }
```

## Divergences

Native custos inspection, independently specified. No EA counterpart or EA fixture conformance applies.
