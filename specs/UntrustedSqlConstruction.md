---
id: UntrustedSqlConstruction
group: Security
kind: semantic
needs: [names, types, flow]
php: { min: "", max: "" }
---

# UntrustedSqlConstruction

## Summary

Detect request-derived data reaching a resolved SQL execution sink as SQL syntax without parameter binding or proven context-appropriate validation.

## Detection

- D1. Report request-derived data reaching a resolved SQL execution sink as SQL syntax without parameter binding or proven context-appropriate validation.
- D2. Resolve builtin calls and named arguments before checking the contract. Project-local callees are considered only when their bodies or immutable summaries establish the required facts. The default is enabled, with positive proof required.
- D3. For flow-dependent conditions, require a reachable source-to-use path with no intervening invalidation. Unknown calls, ambiguous aliases, recursion or analysis-budget exhaustion cannot establish definite facts.
- D4. The sinks are the builtin `PDO::query`, `PDO::exec` and `PDO::prepare` (first argument), called on a receiver whose single resolved class is `PDO` or a subclass that does not override the method, plus project wrappers whose summary forwards an argument to one of them. The three methods exist on every supported PHP version, so they are sinks on every target version even when the stub signature in use was introduced later.
- D5. The statement text is injectable only when a part that contributes characters to it is. Concatenation, interpolation, `??`, both ternary branches (the condition for `?:`), every `match` arm result, array literal items, a variable with a single reaching definition, the separator and pieces of `implode`/`join`, the fill value of `array_fill` and the repeated text of `str_repeat` are followed part by part. A `match` subject and a ternary condition only select a result and contribute nothing themselves. Any other shape keeps the flow verdict. A variable is followed to its reaching assignment only when its scope cannot rebind it behind that assignment; a reference alias, a by-reference closure capture, `foreach` or argument, a `catch` binding or `static` declaration of the same name, `extract()`, an assignment through a variable variable, a `global` declaration of it or a `$GLOBALS` access that is not a different literal entry keeps the report.

Project method wrappers are followed only when a single visible declaration is fixed by a final receiver class, a final method, or a private method accessible at the call. Open overridable methods, ambiguous declarations, conditional class bodies and trait bodies do not establish a wrapper sink. Static calls must name a static method. These contracts compose through nested project wrappers and refresh when class or method dispatch metadata changes.

A wrapper dependency that reads only array values preserves the numeric safety of an array whose values are all numbers, booleans or null. Array keys, whole-array formatting and mixed key/value dependencies remain untrusted; numeric conversion of extracted keys can establish safety again. Return and reference-output dependencies preserve the same distinction.

## Exceptions (no report)

- E1. Value parameters, strict finite allowlists, constant queries, known-safe identifier selection and unknown flows are excluded.
- E2. Unresolved targets, incomplete required facts, suppressed findings and PHP versions lacking the referenced API or language feature produce no report. User declarations shadowing builtin names are not builtin contracts.

- E3. Known boolean or numeric scalar output and resolved builtin `md5`, `sha1`, `hash` or `hash_hmac` hexadecimal output cannot inject the target text syntax. Raw binary or unknown output modes retain source taint. These narrow text proofs do not establish URL or filesystem-path authorization.
- E4. Numeric results are never reported: unary `+`/`-`, the arithmetic operators `-`, `*`, `/`, `%`, `**` (they yield a number or throw, whatever the operands hold), and resolved counting, measuring, comparing and testing builtins such as `count`, `sizeof`, `strlen`, `mb_strlen`, `substr_count`, `strcmp`, `strpos`, `ord`, `crc32`, `in_array`, `is_numeric` or `ctype_digit`, which return only a number or a boolean. `+` is numeric only when one operand is itself provably a number (numeric literal, cast, arithmetic, or a number-returning builtin); two arrays form a union that keeps request text, so that stays reported. A placeholder list such as `implode(',', array_fill(0, count($ids), '?'))` is therefore constant text.
- E5. Request data that a `switch` compared with constant labels is the matched label inside an arm reached only through string labels that do not read as numbers (not empty and not starting, after whitespace, with a digit, a sign or a dot), and after the switch when every other way out replaced it or left the scope: `switch ($sort) { case 'name': case 'date': break; default: $sort = 'name'; }` validates `$sort`. A numeric, empty, boolean or null label, a label expression, a fall-through from another arm, a `default` arm and any other subject prove nothing. The subject is a plain variable or an element of a request superglobal read under one constant key (`switch ($_GET['action'])`, not `$_FILES`); the element is the matched label only while nothing that may run code is evaluated: a call, an object creation, an assignment, an `unset`, or the use of a local variable, property, offset of a local or class constant ends it, as does a reference to the superglobal. A `switch (true)` reads each case expression like the condition of an `if`/`elseif` chain.
- E6. A value quoted by `PDO::quote` is enclosed and escaped for the connection: no report. A method of the same name on another class proves nothing.

## Report

- Range: the complete violating call or expression identified by D1; for a declaration-state violation, the offending property access or assignment. Highlight the final unsafe use for source-to-use findings.
- Severity: error.
- Message: `Bind untrusted values as SQL parameters.`

## Fix

No automatic fix: prepared-query rewriting requires preserving all binding and statement semantics.

## Options

None. Enable or disable this rule through normal rule configuration.

## PHP versions

Apply only on PHP versions supporting the referenced APIs and syntax (within PHP 5.3–8.5). The PDO sinks of D4 apply on all of them. Builtin behavior and argument contracts follow the configured PHP version. Extension-specific calls require resolved extension API symbols; do not treat same-named application functions as extension calls.

## Examples

```php
<?php
function bad(PDO $pdo) { <error descr="Bind untrusted values as SQL parameters.">$pdo->query('SELECT * FROM people WHERE id = ' . $_GET['id'])</error>; }
```

Not reported: only placeholders, a number and a constant keyword reach the text.

```php
<?php
function good(PDO $pdo) {
    $marks = implode(',', array_fill(0, count($_GET['ids']), '?'));
    $direction = match ($_GET['sort']) { 'desc' => 'DESC', default => 'ASC' };
    $stmt = $pdo->prepare("SELECT * FROM people WHERE id IN ($marks) ORDER BY id $direction LIMIT " . ($_GET['page'] * 10));
    $stmt->execute($_GET['ids']);
}
```

## Divergences

Native custos rule. No upstream inspection or fixture comparison applies. These conditions are independently specified and examples are original.
