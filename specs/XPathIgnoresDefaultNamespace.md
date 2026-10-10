---
id: XPathIgnoresDefaultNamespace
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "", max: "" }
---

# XPathIgnoresDefaultNamespace

## Summary

Unprefixed XPath names do not select elements in a nonempty default namespace. Register and use a prefix for that namespace.

## Detection

- D1. A local DOMDocument loads a literal XML root with a nonempty default xmlns and no unnamespaced matching descendant. A DOMXPath created for that document queries a literal //NAME path corresponding to a namespaced element, without registered prefix in the query.
- D1b. The XPath query must immediately follow the DOMXPath construction, so no intervening document mutation can change namespace evidence.
- D2. Resolve builtin functions, classes, constants and methods; user-defined lookalikes are excluded. Follow only bounded local proof in one lexical scope.
- D3. Emit one finding per violating operation, not per supporting evidence statement. Use existing analysis budgets; exhausted proof produces no report.

## Exceptions (no report)

- E1. Unknown XML, complex XPath, namespace resets, or intervening document mutations have no report.
- E2. Both standard suppression forms suppress this native ID.

## Report

- Range: the violating expression or call, unless D1 specifies a narrower range.
- Severity: warning.
- Enabled by default: true.
- Message: Register a namespace prefix for the XPath query.

## Fix

None. Choosing a repair requires intent or additional runtime information.

## Options

None.

## PHP versions

Available throughout supported PHP 5.3–8.5 where the referenced builtin/extension is available. Respect syntax/version-specific receiver contracts; unavailable APIs are not this rule’s concern.

## Examples

```php
<?php
$d=new DOMDocument();$d->loadXML('<root xmlns="urn:demo"><item/></root>');$x=new DOMXPath($d);<warning descr="Register a namespace prefix for the XPath query.">$x->query("//item")</warning>;
```

Valid case:

```php
<?php
$d=new DOMDocument();$d->loadXML('<root xmlns="urn:demo"><item/></root>');$x=new DOMXPath($d);$x->registerNamespace("d","urn:demo");$x->query("//d:item");
```

## Divergences

Native custos inspection, independently specified; no EA counterpart or EA conformance requirement.
