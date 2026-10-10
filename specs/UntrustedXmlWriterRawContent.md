---
id: UntrustedXmlWriterRawContent
group: Security
kind: semantic
needs: [names, types, flow]
php: { min: "", max: "" }
---

# UntrustedXmlWriterRawContent

## Summary

Write untrusted XML content as escaped text.

## Detection

- D1. XMLWriter writeRaw receives directly or through unchanged local bindings an indexed _GET,_POST, _REQUEST or_COOKIE string. Exclude proven literal content, text(), and values reconstructed by a proven XML parser/serializer. Do not assume arbitrary sanitizers establish safety.
- D2. Resolve builtin functions, extension classes, constants and method receivers, including aliases and supported named arguments. Inspect reachable operations in one lexical scope with bounded local flow.
- D3. Emit one finding per violating operation. Never infer policy from names or comments. Unsupported proof and exhausted analysis budgets produce no report.

## Exceptions (no report)

- E1. Exclude shadowed or unresolved symbols, incomplete code, unknown required constants, ambiguous receiver identity and unsupported versions. Reassignment, escaping aliases, and unknown calls invalidate dependent flow facts.
- E2. Both @custos-ignore UntrustedXmlWriterRawContent and @noinspection UntrustedXmlWriterRawContent suppress this native rule.

## Report

- Range: the complete violating expression or call identified by D1; source-to-use findings highlight the final unsafe use, except where D1 specifies the producer.
- Severity: warning.
- Enabled by default: true.
- Message: `Write untrusted XML content as escaped text.`

## Fix

None. A safe repair requires choosing values, error handling, or application behavior.

## Options

None. Use normal rule configuration to enable or disable the inspection.

## PHP versions

Apply throughout PHP 5.3–8.5 only where the extension API and syntax exist.  Do not flag API unavailability through this rule.

## Examples

```php
<?php
function f(XMLWriter $w) { <warning descr="Write untrusted XML content as escaped text.">$w->writeRaw($_POST['caption'])</warning>; }
```

Valid case:

```php
<?php
function f(XMLWriter $w) { $w->text($_POST['caption']); }
```

## Divergences

Native custos inspection, independently specified. No EA counterpart or EA fixture conformance applies.
