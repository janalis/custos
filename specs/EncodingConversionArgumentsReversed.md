---
id: EncodingConversionArgumentsReversed
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "", max: "" }
---

# EncodingConversionArgumentsReversed

## Summary

Converting text away from the output consumer's required encoding can make it invalid input. Convert to the encoding required by the consumer. This inspection is disabled by default because the intended policy or use can vary.

## Detection

- D1. Resolved mb_convert_encoding has literal source/target encodings differing; converted result enters json_encode or response echo with definite Content-Type charset matching the input encoding. Require literal input valid only in claimed source encoding or proven upstream encode contract; variable names do not establish encoding.
- D1a. The initial proof requires a valid non-ASCII UTF-8 literal converted to ISO-8859-1 and passed to json_encode, where the resulting bytes are not valid UTF-8. ASCII-only text is excluded.
- D1d. Compute the supported literal conversion and prove the resulting Latin-1 bytes are invalid UTF-8. Exclude converted bytes that already form valid UTF-8, even when the source literal contains non-ASCII code points; not every Latin-1 byte sequence is invalid UTF-8.
- D2. Resolve builtin functions, classes, constants and methods; user-defined lookalikes are excluded. Follow only bounded local proof in one lexical scope.
- D3. Emit one finding per violating operation, not per supporting evidence statement. Use existing analysis budgets; exhausted proof produces no report.

## Exceptions (no report)

- E1. Do not report unresolved or shadowed builtin names, unknown values, or incomplete receiver/type provenance. Unknown calls, aliases escaping, and intervening writes invalidate local proof. Similar code outside the stated detection contract is not a finding.
- E2. Both standard suppression forms suppress this native ID.

## Report

- Range: the violating expression or call, unless D1 specifies a narrower range.
- Severity: warning.
- Enabled by default: false.
- Message: Convert to the encoding required by the output consumer.

## Fix

None. Choosing a repair requires intent or additional runtime information.

## Options

None.

## PHP versions

Available throughout supported PHP 5.3–8.5 where the referenced builtin/extension is available. Respect syntax/version-specific receiver contracts; unavailable APIs are not this rule’s concern.

## Examples

```php
<?php
$s=<warning descr="Convert to the encoding required by the output consumer.">mb_convert_encoding("été","ISO-8859-1","UTF-8")</warning>;echo json_encode($s);
```

Valid case:

```php
<?php
$s=mb_convert_encoding($latin,"UTF-8","ISO-8859-1");echo json_encode($s);
```

## Divergences

Native custos inspection, independently specified; no EA counterpart or EA conformance requirement.
