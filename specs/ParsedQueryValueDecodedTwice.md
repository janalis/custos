---
id: ParsedQueryValueDecodedTwice
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "", max: "" }
---

# ParsedQueryValueDecodedTwice

## Summary

parse_str already URL-decodes query values. Avoid a second decoding that changes encoded data into a different value.

## Detection

- D1. Resolved parse_str output literal-key value flows directly into urldecode/rawurldecode without transformation. Require known literal query with a percent-encoded percent sequence yielding a second decodable escape or plus, proving meaning changes.
- D1a. Support the decoding call only as a standalone expression statement, the value of a plain assignment whose parent is an expression statement, a direct return value, or the sole expression of an echo statement. Exclude other wrappers and compound expressions: another expression can replace or mutate the parsed output before decoding occurs.
- D2. Resolve builtin functions, classes, constants and methods; user-defined lookalikes are excluded. Follow only bounded local proof in one lexical scope.
- D3. Emit one finding per violating operation, not per supporting evidence statement. Use existing analysis budgets; exhausted proof produces no report.

## Exceptions (no report)

- E1. Do not report unresolved or shadowed builtin names, unknown values, or incomplete receiver/type provenance. Unknown calls, aliases escaping, and intervening writes invalidate local proof. Similar code outside the stated detection contract is not a finding.
- E2. Both standard suppression forms suppress this native ID.

## Report

- Range: the violating expression or call, unless D1 specifies a narrower range.
- Severity: warning.
- Enabled by default: true.
- Message: Use the query value without decoding it a second time.

## Fix

None. Choosing a repair requires intent or additional runtime information.

## Options

None.

## PHP versions

Available throughout supported PHP 5.3–8.5 where the referenced builtin/extension is available. Respect syntax/version-specific receiver contracts; unavailable APIs are not this rule’s concern.

## Examples

```php
<?php
parse_str("next=a%252Fb",$q);$v=<warning descr="Use the query value without decoding it a second time.">urldecode($q["next"])</warning>;
```

Valid case:

```php
<?php
parse_str("next=a%252Fb",$q);$v=$q["next"];
```

## Divergences

Native custos inspection, independently specified; no EA counterpart or EA conformance requirement.
