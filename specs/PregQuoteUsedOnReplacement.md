---
id: PregQuoteUsedOnReplacement
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "", max: "" }
---

# PregQuoteUsedOnReplacement

## Summary

Pattern escaping can insert unwanted characters into replacement text. Keep literal replacements separate from regex pattern escaping.

## Detection

- D1. Resolved preg_replace replacement argument is direct resolved preg_quote call whose literal input contains regex punctuation escaped by preg_quote but not replacement syntax. Highlight preg_quote call; dynamic values and dollar/backslash replacement semantics excluded.
- D2. Resolve builtin functions, classes, constants and methods; user-defined lookalikes are excluded. Follow only bounded local proof in one lexical scope.
- D3. Emit one finding per violating operation, not per supporting evidence statement. Use existing analysis budgets; exhausted proof produces no report.

## Exceptions (no report)

- E1. Do not report unresolved or shadowed builtin names, unknown values, or incomplete receiver/type provenance. Unknown calls, aliases escaping, and intervening writes invalidate local proof. Similar code outside the stated detection contract is not a finding.
- E2. Both standard suppression forms suppress this native ID.

## Report

- Range: the violating expression or call, unless D1 specifies a narrower range.
- Severity: warning.
- Enabled by default: true.
- Message: Keep pattern escaping out of literal replacement text.

## Fix

F1. Replace a direct preg_quote invocation in the replacement argument with the original literal input expression. The decoded literal must contain escaped regex punctuation but neither a dollar sign nor a backslash; the raw argument is a literal, all omitted delimiter evaluation is side-effect-free, and the invocation contains no comments. Preserve the input literal source bytes. Do not fix dynamic input, comment-bearing calls, or potential replacement-group syntax.

Comments introduced by `#` are comments too: suppress a fix whenever its replacement would discard them, just as for line and block comments.

## Options

None.

## PHP versions

Available throughout supported PHP 5.3–8.5 where the referenced builtin/extension is available. Respect syntax/version-specific receiver contracts; unavailable APIs are not this rule’s concern.

## Examples

```php
<?php
$s=preg_replace("/item/",<warning descr="Keep pattern escaping out of literal replacement text.">preg_quote("a.b","/")</warning>,$text);
```

```php
<?php
$s=preg_replace("/item/","a.b",$text);
```

Valid case:

```php
<?php
$s=preg_replace("/item/","a.b",$text);
```

## Divergences

Native custos inspection, independently specified; no EA counterpart or EA conformance requirement.
