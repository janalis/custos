---
id: AssertionContainsRequiredSideEffect
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "", max: "" }
---

# AssertionContainsRequiredSideEffect

## Summary

Identify an assert call containing an assignment to a previously undefined local when a reachable subsequent read requires that local before any unconditional assignment. Restrict to straight-line scope; exclude parameters, global/static locals, references, any earlier local definition or prior read. Do not guess about other side effects.

## Detection

- D1. Report an assert call containing an assignment to a previously undefined local when a reachable subsequent read requires that local before any unconditional assignment. Restrict to straight-line scope; exclude parameters, global/static locals, references, any earlier local definition or prior read. Do not guess about other side effects.
- D1a. A later use counts as a required read only when it executes regardless of the assertion setting. A read placed inside another resolved builtin `assert()` call disappears together with the reported one and is not a required read. A variable passed to an in-out by-reference parameter (`sort`, `array_push`, `array_splice`, `settype`, the output list of `exec`, ...) is still a required read, because the callee consumes the previous value.
- D1b. The proof is abandoned, with no finding, when the local is rebound before any required read by a construct that does not consume its value: a `foreach` key or value target (unless the iterated expression itself reads the local), a `catch` variable, or a positional argument bound to a write-only by-reference parameter of a resolved builtin (the match array of `preg_match`/`preg_match_all`, the count of `str_replace`/`str_ireplace`/`preg_replace`/`preg_replace_callback`, the result of `parse_str`/`mb_parse_str`, the percentage of `similar_text`, the result code of `exec`/`system`/`passthru`, the image information of `getimagesize`/`getimagesizefromstring`, both outputs of `headers_sent`, `getmxrr` and `fsockopen`, and the callable name of `is_callable`). Named or unpacked arguments, and calls where another argument reads the local, do not qualify. Loops and handlers are not straight-line code, so these rebinding sites end the analysis instead of acting as an assignment after which later reads could be ignored.
- D2. Resolve builtin functions, classes, constants and methods, including imports and supported named arguments. Follow only bounded local facts within one lexical scope. Receiver and argument provenance must establish the specific API contract; variable names and comments never prove intent.
- D3. Report one finding per violating operation. A reachable path, unchanged value identity and applicable API behavior are required. Intervening mutation, unknown calls involving tracked values, escaped aliases or exhausted analysis budgets invalidate proof.

## Exceptions (no report)

- E1. Exclude unresolved required symbols, user-defined lookalikes, incomplete syntax, unknown required values and unsupported versions. Cases outside D1 receive no finding.
- E2. Both `@custos-ignore AssertionContainsRequiredSideEffect` and `@noinspection AssertionContainsRequiredSideEffect` suppress this native rule.

## Report

- Range: the complete violating call or expression. Follow D1 if it specifies a narrower range.
- Severity: warning.
- Enabled by default: true.
- Message: `Move required assignments outside assertions.`

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
assert(($result = 7) > 0); echo $result;
```

Valid case:

```php
<?php
$result = 7; assert($result > 0); echo $result;
```

Also valid, because nothing outside an assertion needs the asserted value:

```php
<?php
function digits(string $input): array {
    assert(($found = []) === []);
    preg_match('/[0-9]+/', $input, $found);
    return $found;
}
function half(int $total): bool {
    assert(($part = intdiv($total, 2)) >= 0);
    assert($part <= $total);
    return true;
}
```

## Divergences

Native custos inspection, independently specified; no EA counterpart or EA conformance requirement. Conservative proof intentionally excludes unknown runtime state and intent.
