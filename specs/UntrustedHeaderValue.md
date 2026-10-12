---
id: UntrustedHeaderValue
group: Security
kind: semantic
needs: [names, types, flow]
php: { min: "", max: "" }
---

# UntrustedHeaderValue

## Summary

Detect request-derived data reaching a response-header value without a dominating CR/LF rejection or safe construction summary.

## Detection

- D1. Report request-derived data reaching a response-header value without a dominating CR/LF rejection or safe construction summary.
- D2. Resolve builtin calls and named arguments before checking the contract. Project-local callees are considered only when their bodies or immutable summaries establish the required facts. The default is enabled, with positive proof required.
- D3. For flow-dependent conditions, require a reachable source-to-use path with no intervening invalidation. Unknown calls, ambiguous aliases, recursion or analysis-budget exhaustion cannot establish definite facts.
- D4. A CR/LF rejection is recognized for a variable when the tested expression still carries the single assignment that reaches the header value, and the test is either the condition of an earlier `if` in the same statement list as the use (or as a statement enclosing it) whose body always leaves (`exit`, `return`, `throw`, `break`, `continue`) and whose condition is necessarily true when a line break is present, or the condition of an enclosing `if` whose body (or `else`) is entered only when no line break is present.
- D5. The accepted line-break tests are a two-argument builtin `preg_match` with a literal, modifier-free pattern made of one character class holding exactly CR and LF (delimiter `/`, `#` or `~`, single- or double-quoted), and a two-argument builtin `strpbrk` whose literal character list contains both bytes. They may be used bare, negated with `!`, compared with `===`, `==`, `!==`, `!=` against `0`, `1` (`preg_match` only) or `false` in either operand order, written as `preg_match(...) > 0`, and combined through `||` (rejecting guard) or `&&` (accepting branch). The paired `str_contains` checks and the `preg_match(...) === 0` forms established by the shared flow analysis remain accepted.
- D6. The header value is examined structurally: a concatenation or interpolated string is safe when every part is; a ternary when both results are; a `match` when every arm result is (the subject and arm conditions only select, and a throwing arm yields nothing); a variable when its single reaching assignment is. The walk is bounded to 32 nested steps; deeper values keep the report. A variable is followed to its reaching assignment only when its scope cannot rebind it behind that assignment; a reference alias, a by-reference closure capture, `foreach` or argument, a `catch` binding or `static` declaration of the same name, `extract()`, an assignment through a variable variable, a `global` declaration of it or a `$GLOBALS` access that is not a different literal entry keeps the report.

## Exceptions (no report)

- E1. Constant values, validated finite choices, established CR/LF rejection and unknown flows are excluded.
- E2. Unresolved targets, incomplete required facts, suppressed findings and PHP versions lacking the referenced API or language feature produce no report. User declarations shadowing builtin names are not builtin contracts.

- E3. Known boolean or numeric scalar output and resolved builtin `md5`, `sha1`, `hash` or `hash_hmac` hexadecimal output cannot inject the target text syntax. Raw binary or unknown output modes retain source taint. These narrow text proofs do not establish URL or filesystem-path authorization.
- E4. Results that are numbers or booleans whatever bytes the input holds: resolved builtins returning an integer or boolean (`strlen`, `mb_strlen`, `count`, `crc32`, `ord`, the `strcmp` and `strpos` families, `substr_count`, `levenshtein`, `similar_text`, `in_array`, `array_key_exists`, `str_contains`, `str_starts_with`, `str_ends_with`, every `ctype_*` and `is_*` function), unary `+`/`-`, and binary `-`, `*`, `/`, `%`, `**`. Binary `+` qualifies only when one operand is provably not an array (a literal, an `(int)`, `(float)`, `(bool)` or `(string)` cast, or one of those builtin calls): `array + array` is a union that keeps request strings.
- E5. A guard proves nothing, and the report is kept, when the value is reassigned after it on any path, when it tests another value or a transformed copy, when its body can fall through, when the pattern has modifiers, a negated or wider class, or only one of the two bytes, or when `preg_match` receives an offset.
- E6. Request data that a `switch` compared with constant labels is the matched label inside an arm reached only through string labels that do not read as numbers (not empty and not starting, after whitespace, with a digit, a sign or a dot), and after the switch when every other way out replaced it or left the scope: `switch ($sort) { case 'name': case 'date': break; default: $sort = 'name'; }` validates `$sort`. A numeric, empty, boolean or null label, a label expression, a fall-through from another arm, a `default` arm and a subject that is not a plain variable prove nothing. A `switch (true)` reads each case expression like the condition of an `if`/`elseif` chain.

## Report

- Range: the complete violating call or expression identified by D1; for a declaration-state violation, the offending property access or assignment. Highlight the final unsafe use for source-to-use findings.
- Severity: error.
- Message: `Reject line breaks and validate this header value.`

## Fix

No automatic fix: header-specific validation must preserve application meaning.

A fix is offered only when the concrete prerequisites above are proven. Edits use exact source byte ranges, preserve comments and argument evaluation order, and introduce syntax supported by the configured PHP version. Never invent error-handling policy.

## Options

None. Enable or disable this rule through normal rule configuration.

## PHP versions

Apply only on PHP versions supporting the referenced APIs and syntax (within PHP 5.3–8.5). Builtin behavior and argument contracts follow the configured PHP version. Extension-specific calls require resolved extension API symbols; do not treat same-named application functions as extension calls.

## Examples

```php
<?php
<error descr="Reject line breaks and validate this header value.">header('X-Label: ' . $_GET['label'])</error>;

$label = $_GET['label'];
if (preg_match('/[\r\n]/', $label)) {
    exit;
}
header('X-Label: ' . $label); // not reported: line breaks were rejected

header('Content-Length: ' . strlen($_POST['body'])); // not reported: integer result

$direction = match ($_GET['sort']) { 'desc' => 'DESC', default => 'ASC' };
header('X-Sort: ' . $direction); // not reported: only literals can be selected
```

## Divergences

Native custos rule. No upstream inspection or fixture comparison applies. These conditions are independently specified and examples are original.
