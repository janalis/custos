---
id: UntrustedXmlExternalEntityExpansion
group: Security
kind: semantic
needs: [names, types, flow]
php: { min: "", max: "" }
---

# UntrustedXmlExternalEntityExpansion

## Summary

Expanding external entities in request-controlled XML can expose external resources. Disable that capability for untrusted documents.

## Detection

- D1. Resolved DOMDocument::loadXML receives direct request-superglobal data and options definitely contain both LIBXML_NOENT and LIBXML_DTDLOAD, without a supported LIBXML_NO_XXE option. LIBXML_NONET alone does not suppress: it blocks network resources but permits other external resource access. Report explicit external entity access capability. The no-XXE bit suppresses only for PHP 8.4 or later; the same numeric bit on PHP 8.3 or earlier does not establish protection.
- D1a. Input must be a directly written array-field access on $_GET, $_POST, $_REQUEST or $_COOKIE. Stored local values and other superglobals are excluded.
- D1b. Inspect at most 64 preceding statements in the same straight-line lexical prefix. Each must be an ordinary, non-reference simple-local assignment whose value is a resolved new DOMDocument expression with no constructor arguments. The assignment target must not be the same request-superglobal variable used by the input access. Any other preceding operation discards the untouched-input proof, including superglobal writes, sanitization calls and unknown operations. Constructors with any arguments are excluded, including otherwise harmless literal arguments, to keep the request-source proof conservative.
- D2. Resolve builtin functions, classes, constants and methods; user-defined lookalikes are excluded. Follow only bounded local proof in one lexical scope.
- D3. Emit one finding per violating operation, not per supporting evidence statement. Use existing analysis budgets; exhausted proof produces no report.

## Exceptions (no report)

- E1. Do not report unresolved or shadowed builtin names, unknown values, or incomplete receiver/type provenance. Unknown calls, aliases escaping, and intervening writes invalidate local proof. Similar code outside the stated detection contract is not a finding.
- E2. Both standard suppression forms suppress this native ID.

## Report

- Range: the violating expression or call, unless D1 specifies a narrower range.
- Severity: warning.
- Enabled by default: true.
- Message: Disable external entity expansion for request-controlled XML.

## Fix

None. Choosing a repair requires intent or additional runtime information.

## Options

None.

## PHP versions

Available throughout supported PHP 5.3–8.5 when DOM/libxml is available. Treat LIBXML_NO_XXE as protection only on PHP 8.4 or later; a numeric no-XXE bit on an earlier target does not suppress the finding. LIBXML_NONET alone does not establish protection against local external resources.

## Examples

```php
<?php
$d=new DOMDocument(); <warning descr="Disable external entity expansion for request-controlled XML.">$d->loadXML($_POST["xml"],LIBXML_NOENT|LIBXML_DTDLOAD)</warning>;
```

Valid case:

```php
<?php
$d=new DOMDocument(); $d->loadXML($_POST["xml"],LIBXML_NONET);
```

## Divergences

Native custos inspection, independently specified; no EA counterpart or EA conformance requirement.
