---
id: LibxmlErrorBufferNeverCleared
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "", max: "" }
---

# LibxmlErrorBufferNeverCleared

## Summary

Repeated XML parsing can accumulate internally collected errors. Retrieve and clear the error buffer during the parsing loop. This inspection is disabled by default because the intended policy or use can vary.

## Detection

- D1. A resolved libxml_use_internal_errors(true) precedes a loop containing a resolved DOMDocument::loadXML or simplexml_load_string; neither loop nor guaranteed cleanup calls libxml_clear_errors. Require definite enabled state and no unknown control escape.
- D1d. Inspect loop initialization, condition, iterable and increment expressions as well as the body. A header that clears libxml errors or disables internal collection excludes the finding; unknown calls in the header discard collection-state proof. Exclude statically empty iterables and statically false conditions because the parse loop never runs.
- D1f. Exclude parse operations with nested calls in their argument expressions, and DOMDocument construction with calls in constructor arguments. Those calls can clear errors or change error collection before parsing; unsupported argument side effects discard the accumulation proof.
- D2. Resolve builtin functions, classes, constants and methods; user-defined lookalikes are excluded. Follow only bounded local proof in one lexical scope.
- D3. Emit one finding per violating operation, not per supporting evidence statement. Use existing analysis budgets; exhausted proof produces no report.

## Exceptions (no report)

- E1. Do not report unresolved or shadowed builtin names, unknown values, or incomplete receiver/type provenance. Unknown calls, aliases escaping, and intervening writes invalidate local proof. Similar code outside the stated detection contract is not a finding.
- E2. Both standard suppression forms suppress this native ID.

## Report

- Range: the violating expression or call, unless D1 specifies a narrower range.
- Severity: warning.
- Enabled by default: false.
- Message: Clear collected XML errors during repeated parsing.

## Fix

None. Choosing a repair requires intent or additional runtime information.

## Options

None.

## PHP versions

Available throughout supported PHP 5.3–8.5 where the referenced builtin/extension is available. Respect syntax/version-specific receiver contracts; unavailable APIs are not this rule’s concern.

## Examples

```php
<?php
libxml_use_internal_errors(true); $d=new DOMDocument(); <warning descr="Clear collected XML errors during repeated parsing.">foreach($docs as $xml){$d->loadXML($xml);}</warning>
```

Valid case:

```php
<?php
libxml_use_internal_errors(true); $d=new DOMDocument(); foreach($docs as $xml){$d->loadXML($xml);libxml_clear_errors();}
```

## Divergences

Native custos inspection, independently specified; no EA counterpart or EA conformance requirement.
