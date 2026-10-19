---
id: CurlReadCallbackExceedsRequestedSize
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "", max: "" }
---

# CurlReadCallbackExceedsRequestedSize

## Summary

Identify READFUNCTION callback whose every reachable nonempty return has length provably greater than its third parameter. Restrict initial analysis to str_repeat of one byte and requested size plus positive integer; exclude error/pause sentinels.

## Detection

- D1. Report READFUNCTION callback whose every reachable nonempty return has length provably greater than its third parameter. Restrict initial analysis to str_repeat of one byte and requested size plus positive integer; exclude error/pause sentinels.
- D2. Resolve builtin functions, classes, constants and methods, including imports and supported named arguments. Follow only bounded local facts within one lexical scope. Receiver and argument provenance must establish the specific API contract; variable names and comments never prove intent.
- D3. Report one finding per violating operation. A reachable path, unchanged value identity and applicable API behavior are required. Intervening mutation, unknown calls involving tracked values or escaped aliases invalidate proof. When the proof is inconclusive, nothing is reported.
- D4. The third parameter must provably keep the value cURL passed. The whole callback body is searched, not only the code before a return, because a loop can run a later statement first. The proof is dropped when the parameter is: an assignment target (plain, compound, destructuring), a `++`/`--` operand, a `foreach` key or value target, named by `global`, `static`, `unset` or a `catch` clause, bound by reference (`$a = &$limit`, `[&$limit]`, closure `use (&$limit)`), or passed itself as an argument of any call or `new` (the receiving parameter may be by-reference). It is also dropped when the body contains a construct able to write a local by name: a variable-variable, `extract`, `parse_str`, `mb_parse_str`, `import_request_variables`, `include`/`require` or `eval`. Plain reads, by-value closure captures and arguments that are expressions built from the parameter (`f($limit + 1)`) keep the proof.

## Exceptions (no report)

- E1. Exclude unresolved required symbols, user-defined lookalikes, incomplete syntax, unknown required values and unsupported versions. Cases outside D1 receive no finding.
- E2. Both `@custos-ignore CurlReadCallbackExceedsRequestedSize` and `@noinspection CurlReadCallbackExceedsRequestedSize` suppress this native rule.
- E3. A callback that may rebind its requested-size parameter (D4) is not reported, even when the rebinding cannot actually lower the value.

## Report

- Range: the complete violating call or expression. Follow D1 if it specifies a narrower range.
- Severity: warning.
- Enabled by default: true.
- Message: `Return no more than the requested byte count.`

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
curl_setopt($ch, CURLOPT_READFUNCTION, fn($ch, $h, $limit) => str_repeat('x', $limit + 2));
```

Valid case:

```php
<?php
curl_setopt($ch, CURLOPT_READFUNCTION, fn($ch, $h, $limit) => str_repeat('x', $limit));
```

Valid case (the size is clamped through a by-reference parameter):

```php
<?php
function clamp(&$n) { $n = max(0, $n - 16); }
curl_setopt($ch, CURLOPT_READFUNCTION, function ($ch, $h, $limit) { clamp($limit); return str_repeat('x', $limit + 1); });
```

## Divergences

Native custos inspection, independently specified; no EA counterpart or EA conformance requirement. Conservative proof intentionally excludes unknown runtime state and intent.
