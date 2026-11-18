---
id: UnvalidatedRedirectTarget
group: Security
kind: flow
needs: [names, types, flow]
php: { min: "", max: "" }
---

# UnvalidatedRedirectTarget

## Summary

Detect request-derived data reaching a Location header without a proven permitted-local-target or finite destination policy.

## Detection

- D1. Report request-derived data reaching a Location header without a proven permitted-local-target or finite destination policy.
- D1a. The first argument of `header()` is read as the ordered operands of its concatenation chain and of any interpolated string (double-quoted or heredoc) inside it, ignoring grouping parentheses. The leading operands whose text is statically known form the written prefix; the call is a redirect when that prefix starts with `Location:` in any letter case. Every operand after the prefix is checked, so request data followed by a suffix (`'Location: ' . $_GET['return'] . '?saved=1'`) or embedded in an interpolated string is reported exactly like a trailing operand.
- D1b. A `match` expression, written in place or reached through a variable with one agreeing definition, contributes only the request data of its arm results. The subject chooses an arm but is never part of the produced value. An arm that yields request data is still reported. Selections nested more than eight levels deep are treated as request-derived. A variable is followed to its reaching assignment only when nothing in its scope can change it behind its plain assignments, anywhere and in any statement order: `++` or `--` on it or on one of its elements; a `catch`, `foreach` key or value binding it; a reference to it (a by-reference assignment on either side, a by-reference `foreach` over it, a by-reference array item, a by-reference capture by a nested closure, or an argument, other than the inspected argument itself, that a resolved callee binds by reference or that a callee custos cannot resolve receives); a write naming its target at run time (a variable variable, `extract()`, one-argument `parse_str()` or `mb_parse_str()`, `include`/`require`, `eval`); or a `$GLOBALS` access other than a literal entry naming another variable. A `global` or `static` declaration of it, or a by-reference capture by the closure that is the scope, removes the proof too. In top-level code, where locals are globals, a `global` declaration naming it in any function of the file, or a `$GLOBALS` access anywhere in the file, does as well. The same holds for every local whose value an assignment stores into it, whatever the assigned expression computes, directly or through further assignments. Otherwise the report is kept.
- D2. Resolve builtin calls and named arguments before checking the contract. Project-local callees are considered only when their bodies or immutable summaries establish the required facts. The default is enabled, with positive proof required.
- D3. For flow-dependent conditions, require a reachable source-to-use path with no intervening invalidation. Unknown calls, ambiguous aliases, recursion or code too complex to analyze cannot establish definite facts.

## Exceptions (no report)

- E1. Literal destinations, configured allowlisted origin/path validation and unknown taint flows are excluded.
- E1a. A `match` over request data whose arms all produce values free of request data is a finite destination policy.
- E1b. Fixed local path with encoded query values: when the written prefix after `Location:` starts with a single `/` (the next character is neither `/` nor `\`) and already contains `?`, an operand produced by the builtin `urlencode()`, `rawurlencode()` or `http_build_query()` (directly, or through a variable with one agreeing definition) is not reported. Such a value can only be query data on the current origin. Any other request-derived operand in the same header is still reported, as is an encoded value without that prefix, in a path segment, or after a `//` or `/\` prefix.
- E1c. A backtick command as the header argument is shell output, not written header text; it has no written prefix and is not reported by this rule.
- E2. Unresolved targets, incomplete required facts, suppressed findings and PHP versions lacking the referenced API or language feature produce no report. User declarations shadowing builtin names are not builtin contracts.
- E3. Request data that a `switch` compared with constant labels is the matched label inside an arm reached only through string labels that do not read as numbers (not empty and not starting, after whitespace, with a digit, a sign or a dot), and after the switch when every other way out replaced it or left the scope: `switch ($sort) { case 'name': case 'date': break; default: $sort = 'name'; }` validates `$sort`. A numeric, empty, boolean or null label, a label expression, a fall-through from another arm, a `default` arm and any other subject prove nothing. The subject is a plain variable or an element of a request superglobal read under one constant key (`switch ($_GET['action'])`, not `$_FILES`); the element is the matched label only while nothing that may run code is evaluated: a call, an object creation, an assignment, an `unset`, or the use of a local variable, property, offset of a local or class constant ends it, as does a reference to the superglobal. A `switch (true)` reads each case expression like the condition of an `if`/`elseif` chain.

## Report

- Range: the complete violating call or expression identified by D1; for a declaration-state violation, the offending property access or assignment. Highlight the final unsafe use for source-to-use findings.
- Severity: error.
- Message: `Validate the redirect target against the permitted destination policy.`

## Fix

No automatic fix: the set of permitted destinations must be supplied by the application.

## Options

None.

## PHP versions

Apply only on PHP versions supporting the referenced APIs and syntax (within PHP 5.3–8.5). Builtin behavior and argument contracts follow the configured PHP version. Extension-specific calls require resolved extension API symbols; do not treat same-named application functions as extension calls.

## Examples

```php
<?php
<error descr="Validate the redirect target against the permitted destination policy.">header('Location: ' . $_GET['next'])</error>;
<error descr="Validate the redirect target against the permitted destination policy.">header("Location: {$_GET['next']}#top")</error>;

// Not reported: finite selection, and encoded query data on a fixed local path.
$page = match ($_GET['tab']) { 'billing' => '/account/billing', default => '/account' };
header('Location: ' . $page);
header('Location: /find?term=' . rawurlencode($_GET['term']));
```

## Divergences

Native custos inspection, independently specified; no EA counterpart or EA conformance requirement.
