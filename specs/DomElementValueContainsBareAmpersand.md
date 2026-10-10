---
id: DomElementValueContainsBareAmpersand
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "", max: "" }
---

# DomElementValueContainsBareAmpersand

## Summary

Append a text node for literal ampersand text.

## Detection

- D1. DOMDocument createElement second argument is a constant string containing an ampersand not starting a syntactically valid numeric reference or a declared/predefined XML entity. Accept amp, lt, gt, apos and quot; custom declarations require proven DTD provenance.
- D2. Resolve builtin functions, extension classes, constants and method receivers, including aliases and supported named arguments. Inspect reachable operations in one lexical scope with bounded local flow.
- D3. Emit one finding per violating operation. Never infer policy from names or comments. Unsupported proof and exhausted analysis budgets produce no report.

## Exceptions (no report)

- E1. Exclude shadowed or unresolved symbols, incomplete code, unknown required constants, ambiguous receiver identity and unsupported versions. Reassignment, escaping aliases, and unknown calls invalidate dependent flow facts.
- E2. Both @custos-ignore DomElementValueContainsBareAmpersand and @noinspection DomElementValueContainsBareAmpersand suppress this native rule.

## Report

- Range: the complete violating expression or call identified by D1; source-to-use findings highlight the final unsafe use, except where D1 specifies the producer.
- Severity: warning.
- Enabled by default: true.
- Message: `Append a text node for literal ampersand text.`

## Fix

None. A safe repair requires choosing values, error handling, or application behavior.

## Options

None. Use normal rule configuration to enable or disable the inspection.

## PHP versions

Apply throughout PHP 5.3–8.5 only where the extension API and syntax exist.  Do not flag API unavailability through this rule. Legacy DOM classes are the initial target; modern Dom namespace classes require independently established matching contracts.

## Examples

```php
<?php
$d=new DOMDocument(); $n=<warning descr="Append a text node for literal ampersand text.">$d->createElement('label', 'Salt & pepper')</warning>;
```

Valid case:

```php
<?php
$d=new DOMDocument(); $n=$d->createElement('label'); $n->appendChild($d->createTextNode('Salt & pepper'));
```

## References

- [PHP API contract](https://www.php.net/manual/en/domdocument.createelement.php).

## Divergences

Native custos inspection, independently specified. No EA counterpart or EA fixture conformance applies.
