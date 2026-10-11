---
id: SimpleXmlChildValueContainsBareAmpersand
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "", max: "" }
---

# SimpleXmlChildValueContainsBareAmpersand

## Summary

Identify SimpleXMLElement::addChild value with known ampersand not part of a valid entity/reference. Exclude named entities whose declarations may come from the document DTD, unknown values and name-only calls.

## Detection

- D1. Report SimpleXMLElement::addChild value with known ampersand not part of a valid entity/reference. Exclude named entities whose declarations may come from the document DTD, unknown values and name-only calls.
- D1a. A run from `&` to the next `;` counts as a possible named entity only when the text between them is an XML Name: a letter, `_` or `:` first, then letters, digits, `-`, `.`, `_` or `:` (plus the non-ASCII name characters of XML 1.0). Any other run, such as `&r=2;`, `&1a;`, `&;` or one containing whitespace, cannot be declared by a DTD and is reported unless it is a valid character reference (`&#38;`, `&#x26;`).
- D2. Resolve builtin functions, classes, constants and methods, including imports and supported named arguments. Follow only bounded local facts within one lexical scope. Receiver and argument provenance must establish the specific API contract; variable names and comments never prove intent.
- D3. Report one finding per violating operation. A reachable path, unchanged value identity and applicable API behavior are required. Intervening mutation, unknown calls involving tracked values, escaped aliases or exhausted analysis budgets invalidate proof.

## Exceptions (no report)

- E1. Exclude unresolved required symbols, user-defined lookalikes, incomplete syntax, unknown required values and unsupported versions. Cases outside D1 receive no finding.
- E2. Both `@custos-ignore SimpleXmlChildValueContainsBareAmpersand` and `@noinspection SimpleXmlChildValueContainsBareAmpersand` suppress this native rule.

## Report

- Range: the complete violating call or expression. Follow D1 if it specifies a narrower range.
- Severity: warning.
- Enabled by default: true.
- Message: `Escape bare ampersands in XML child values.`

## Fix

None. A repair requires a semantic decision or application error-handling policy.

## Options

None.

## PHP versions

Available in PHP 5.3–8.5 where the referenced API exists. Do not apply extension contracts when the resolved API is unavailable. Gate syntax and version-specific behavior to the configured target.

## Examples

Finding (enable the rule explicitly when disabled by default):

```php
<?php
$xml = new SimpleXMLElement('<root/>'); $xml->addChild('label', 'C & D');
```

Valid case:

```php
<?php
$xml = new SimpleXMLElement('<root/>'); $xml->addChild('label', 'C &amp; D');
```

Finding, the run before `;` is not an XML Name so no DTD can declare it:

```php
<?php
$xml = new SimpleXMLElement('<root/>'); $xml->addChild('query', 'q=1&r=2;');
```

Valid case, the entity may be declared by the document:

```php
<?php
$xml = new SimpleXMLElement('<root/>'); $xml->addChild('label', '&company.name;');
```

## Divergences

Native custos inspection, independently specified; no EA counterpart or EA conformance requirement. Conservative proof intentionally excludes unknown runtime state and intent.
