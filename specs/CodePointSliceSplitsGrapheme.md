---
id: CodePointSliceSplitsGrapheme
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "7.0", max: "" }
---

# CodePointSliceSplitsGrapheme

## Summary

A visible character can contain several Unicode code points. Slice display text at grapheme boundaries to preserve the complete character. This inspection is disabled by default because the intended policy or use can vary.

## Detection

- D1. Resolved mb_substr with explicit UTF-8 slices a known valid literal string at a boundary between base character and combining mark or within a known emoji sequence. Restrict initial proof to combining marks U+0300–036F and fixed literal start/length.
- D1d. The supported combining-mark proof excludes a mark immediately following a control character, including CR, LF and other Unicode control characters. Such a boundary is not evidence of splitting one grapheme. Unsupported segmentation contexts discard the proof.
- D2. Resolve builtin functions, classes, constants and methods; user-defined lookalikes are excluded. Follow only bounded local proof in one lexical scope.
- D3. Emit one finding per violating operation, not per supporting evidence statement. Use existing analysis budgets; exhausted proof produces no report.

## Exceptions (no report)

- E1. Do not report unresolved or shadowed builtin names, unknown values, or incomplete receiver/type provenance. Unknown calls, aliases escaping, and intervening writes invalidate local proof. Similar code outside the stated detection contract is not a finding.
- E2. Both standard suppression forms suppress this native ID.

## Report

- Range: the violating expression or call, unless D1 specifies a narrower range.
- Severity: warning.
- Enabled by default: false.
- Message: Slice display text at grapheme boundaries.

## Fix

None. Choosing a repair requires intent or additional runtime information.

## Options

None.

## PHP versions

Requires PHP 7.0 or later.

## Examples

```php
<?php
$s=<warning descr="Slice display text at grapheme boundaries.">mb_substr("e\u{0301}",0,1,"UTF-8")</warning>;
```

Valid case:

```php
<?php
$s=grapheme_substr("e\u{0301}",0,1);
```

## Divergences

Native custos inspection, independently specified; no EA counterpart or EA conformance requirement.
