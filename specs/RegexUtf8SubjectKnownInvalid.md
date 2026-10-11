---
id: RegexUtf8SubjectKnownInvalid
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "", max: "" }
---

# RegexUtf8SubjectKnownInvalid

## Summary

Identify preg_match, preg_match_all, preg_replace or preg_split with known u modifier and a constant subject whose bytes fail strict UTF-8 validation; exclude byte-oriented patterns.

## Detection

- D1. Report preg_match, preg_match_all, preg_replace or preg_split with known u modifier and a constant subject whose bytes fail strict UTF-8 validation; exclude byte-oriented patterns.
- D1a. The u modifier is known only when the constant pattern has a usable delimiter layout: after optional leading whitespace, the delimiter is an ASCII character other than a letter, a digit, a backslash or NUL; the body ends at the first unescaped closing delimiter; for the bracket-style delimiters `(`, `[`, `{` and `<` the closing delimiter is the matching `)`, `]`, `}` or `>`, counted with nesting; everything after it consists only of modifier letters (plus spaces and newlines), one of which is `u`.
- D2. Resolve builtin functions, classes, constants and methods, including imports and supported named arguments. Follow only bounded local facts within one lexical scope. Receiver and argument provenance must establish the specific API contract; variable names and comments never prove intent.
- D3. Report one finding per violating operation. A reachable path, unchanged value identity and applicable API behavior are required. Intervening mutation, unknown calls involving tracked values, escaped aliases or exhausted analysis budgets invalidate proof.

## Exceptions (no report)

- E1. Exclude unresolved required symbols, user-defined lookalikes, incomplete syntax, unknown required values and unsupported versions. Cases outside D1 receive no finding.
- E1a. A `u` that belongs to the pattern body is not a modifier: `'[a[u]]'`, `'{a{1}u}'` and `'(a(u))'` carry no modifiers and match bytewise. A pattern with no closing delimiter, an invalid delimiter or an unknown character in its modifier list does not compile, so its subject is never validated and nothing is reported.
- E2. Both `@custos-ignore RegexUtf8SubjectKnownInvalid` and `@noinspection RegexUtf8SubjectKnownInvalid` suppress this native rule.

## Report

- Range: the complete violating call or expression. Follow D1 if it specifies a narrower range.
- Severity: error.
- Enabled by default: true.
- Message: `Supply valid UTF-8 to a Unicode pattern.`

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
preg_match('~.+~u', "\xFF");
```

Valid case:

```php
<?php
preg_match('~.+~', "\xFF");
```

Bracket-style delimiters (only the first call has the u modifier):

```php
<?php
preg_match('(a(b))u', "ab\xFF"); // finding
preg_match('(a(u))', "au\xFF");  // no modifiers: no finding
```

## Divergences

Native custos inspection, independently specified; no EA counterpart or EA conformance requirement. Conservative proof intentionally excludes unknown runtime state and intent.
