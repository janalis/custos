---
id: ParseStrKeyNormalizationMismatch
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "", max: "" }
---

# ParseStrKeyNormalizationMismatch

## Summary

parse_str normalizes dots and spaces in parameter names to underscores. Read the normalized key from the parsed result.

## Detection

- D1. Resolved parse_str literal query and explicit local output array contains a key with dot/space normalized to underscore. A subsequent unguarded result access uses original key, which is absent after parsing. Ignore bracket syntax and duplicate normalized keys in initial proof.
- D1d. Require an actual value read of the absent dimension. Exclude assignment targets, unset, by-reference binding or arguments, and calls whose argument contract does not prove a read. Any mutation of the result array in the same containing statement discards the access proof. Query keys with leading whitespace after URL decoding are unsupported and excluded.
- D1e. All argument uses are excluded because they can involve reference passing. Also exclude increment/decrement, unset and write contexts. Any same-statement mutation or call receiving the result array discards the actual-read proof.
- D2. Resolve builtin functions, classes, constants and methods; user-defined lookalikes are excluded. Follow only bounded local proof in one lexical scope.
- D3. Emit one finding per violating operation, not per supporting evidence statement. Use existing analysis budgets; exhausted proof produces no report.

## Exceptions (no report)

- E1. Do not report unresolved or shadowed builtin names, unknown values, or incomplete receiver/type provenance. Unknown calls, aliases escaping, and intervening writes invalidate local proof. Similar code outside the stated detection contract is not a finding.
- E2. Both standard suppression forms suppress this native ID.

## Report

- Range: the violating expression or call, unless D1 specifies a narrower range.
- Severity: warning.
- Enabled by default: true.
- Message: Read the normalized query parameter key.

## Fix

F1. Replace the directly written literal array dimension with a single-quoted normalized query key, escaping single quotes and backslashes. The literal query must prove that normalized key exists uniquely: no duplicate normalized keys, bracket notation, or NUL bytes. Preserve all surrounding access bytes; unavailable for variable/computed dimensions or ambiguous input.

## Options

None.

## PHP versions

Available throughout supported PHP 5.3–8.5 where the referenced builtin/extension is available. Respect syntax/version-specific receiver contracts; unavailable APIs are not this rule’s concern.

See the [PHP query parsing contract](https://www.php.net/manual/en/function.parse-str.php) for the relevant builtin behavior.

## Examples

```php
<?php
parse_str("user.name=Ada",$q);echo <warning descr="Read the normalized query parameter key.">$q["user.name"]</warning>;
```

```php
<?php
parse_str("user.name=Ada",$q);echo $q['user_name'];
```

Valid case:

```php
<?php
parse_str("user.name=Ada",$q);echo $q["user_name"];
```

## Divergences

Native custos inspection, independently specified; no EA counterpart or EA conformance requirement.
