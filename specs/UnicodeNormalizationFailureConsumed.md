---
id: UnicodeNormalizationFailureConsumed
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "5.3", max: "" }
---

# UnicodeNormalizationFailureConsumed

## Summary

Check normalization before consuming text.

## Detection

Follow each producer reaching a local read, including results written on an
earlier loop turn or before a backward jump. A check of an earlier result does
not check a later replacement; a check repeated before every consumption does.
Ignore producers belonging to other result contracts when selecting an
unchecked result.

- D1. Report a result of resolved Normalizer::normalize or normalizer_normalize reaching strlen or strtoupper without a dominating strict false guard. Follow direct expressions and unchanged local assignments; storage and testing alone are excluded.
- D2. Resolve builtin functions, extension classes, method receivers, constants and named arguments. Follow unchanged locals and aliases only within the same bounded lexical scope. Require a reachable path; merge branch facts only when all incoming paths establish them.
- D3. Emit one finding at the violating operation, not one per evidence statement. Rebinding, writes, by-reference escape, unresolved calls that can mutate relevant state, and exhausted analysis budgets remove proof.
- D4. A strict false guard is a condition that, on the path reaching the consumption, proves the value is not false. Accepted atoms: the value compared with `false` by `===` or `!==` (either operand order), and builtin `is_string()` applied to the value, alone or compared strictly with a boolean literal. The compared operand may be a plain assignment of the normalization result to the consumed variable written inside the condition. Atoms combine through `!`, through `&&`/`and` when the condition holds and through `||`/`or` when it fails.
- D5. A guard dominates the consumption when the consumption is in: the guarded branch of an `if`/`elseif`/`else` chain (a later branch sees every earlier condition failed); the matching operand of a ternary; the right operand of a short-circuit `&&`, `and`, `||` or `or`; an arm of `match (true)` whose single condition proves the value, or any arm after a condition that fails only when the value is not false; an arm of a `match` on the value itself placed after a literal `false` condition; or any statement after an earlier sibling `if` without alternatives whose body always leaves the path, or after an expression statement that short-circuits into `exit` or `throw` when the value is false. The guard must test the same reaching normalization result, after it was produced, with no rebinding or invalidation in between.

## Exceptions (no report)

- E1. Exclude shadowed builtins, unresolved receivers, incomplete syntax and unknown prerequisite facts. An extension function name alone does not prove an application function is that builtin.
- E2. A dominating valid guard or repair described in Detection prevents the finding. Both legacy noinspection and custos-ignore forms suppress this ID.
- E3. A value the shared flow analysis proves is not false at the use is never reported, whatever form the proof takes: a truthiness or `empty()` test, a loose comparison with `false` or `null`, a type predicate, or an early exit whose `if` also has an `else` branch that continues. The result is a string or false, so `if (!$s) { return; }`, `if ($s == false) { return; }` and `$s ? strlen($s) : 0` leave nothing to report. `assert()` can be compiled out and a short ternary on another value tests nothing: both stay reported. The same holds for a statement that leaves on the failing outcome (`$v !== false or exit;`, `$v or throw …;`, `$v ?: throw …;`), for a failing branch that ends in a call that never returns (a callee declared `never`, or a function or a final or private method every path of which throws or exits), for a test stored in a boolean and tested later while neither the boolean nor the tested variable has been written, and, on PHP 8.0 and later outside any `try` with a `catch`, for a builtin call that returned after receiving the value for a parameter that only accepts arrays, objects or callables: such a call throws a `TypeError` for `false`. A `switch (true)` case is read like the condition of an `if`/`elseif` chain; in a `switch` on the variable itself with constant labels, the cases after a `false` or `null` label and the `default` hold a value that is neither, as does a `true` case. Positions PHP may skip prove nothing: the right side of `??=`, the arguments of a null-safe call and anything chained after `?->`. A null-safe call to a never-returning method does not end the path, and `f(...)` creates a closure without calling `f`. An arm of `match (true)` is read like a `switch (true)` case; an arm that throws, exits or calls a function that never returns takes no part in what holds after the `match`. `|` and `&` between two tests are read like `||` and `&&` for the outcome that fixes both operands (a false `|`, a true `&`), and a strict `in_array($v, [...], true)` over literals that include neither `false` nor `null` excludes both where it holds. `goto` jumps are followed, except in a scope that also contains a `try`, which the flow analysis leaves alone. Where a `goto` label is written before the use, and after the place where its scope first names the checked variable, a variable that can hold the same value, or code that may write any variable (`extract`, `compact`, `get_defined_vars`, `parse_str` given one argument, a variable variable, `$GLOBALS`, `eval`, an included file), a jump can land between a guard and the use: only what the flow analysis establishes counts there, and a guard recognised by where it is written proves nothing. Two variables can hold the same value when one is written from the other by a plain assignment, an assignment by reference or `??=`, when a `foreach` takes one out of the other, or through a chain of those; in the value assigned, the receiver and the arguments of a call, the operands of an operator other than `??`, the variables a string embeds and an element key do not count, and a statement that only reads (`echo $prefix . $parts['host'];`) ties nothing. A parameter is named from the start of the function, and a checked expression that reads `$this` or no variable counts every label before it. In a scope the flow analysis does not follow (one that also contains a `try`, a generator, a scope that evaluates or includes code) a guarded use after such a label is therefore reported. A label written before all of that, such as a `retry:` label ahead of the `try` that produces and checks the value, changes nothing, whatever else the function reads beside the checked value; at file level every label written before the use counts. What was established of a `static` local holds until the next call, or code that runs without a call being written (an object converted to text, a magic property or offset access, the destructor of an object a variable held when it is given another value), that may enter the function again and assign the variable. A call of PHP's own counts as such when it may run code of the project: it is given a callback other than the name of a PHP function that only computes on text and numbers and raises no warning whatever text it is given (a closed list: `strlen`, `trim`, `intval` and the like, not `hexdec`, `abs` or `chr`, which warn about text that is no number), an object it may convert, a class name it loads (`method_exists($name, 'm')`, `is_callable('A::b')`, `defined('A::B')`, `class_alias($name, 'B')`, `new ReflectionClass($name)`, `$arrayObject->setIteratorClass($name)`), a class to construct (`$statement->fetchObject('Row')`) or a filter that one of PHP's own `FILTER_…` constants does not name, which may be the one that calls back, or it runs what was registered earlier (an error, session or output-buffer handler, the callbacks of a cURL transfer or an XML parser, an iterator wrapped in another). A warning that a PHP function raises by itself when called directly (`hex2bin` given an odd number of digits, a file that cannot be read) is not counted, although the error handler of the project receives it. This limits every guard form of this specification, not only the ones the flow analysis reads.

## Report

- Range: the violating expression or call described in D1; stateful absence checks highlight the final unsafe operation or loop.
- Severity: warning.
- Enabled by default: true.
- Message: `Check normalization before consuming text.`

## Fix

No automatic fix. The required repair depends on error handling, data selection, lifecycle ordering or application policy.

All edits use exact byte ranges and preserve comments and evaluation order. Offer no fix when its stated prerequisites cannot be proven or the replacement is unsupported by the configured PHP version.

## Options

None.

## PHP versions

Requires PHP 5.3 or later and the referenced extension API. Respect method-specific introduction versions within PHP 5.3–8.5; unsupported optional APIs cannot establish facts. Named arguments require PHP 8.0 or later.

## Examples

```php
<?php
$s = Normalizer::normalize($input); <warning descr="Check normalization before consuming text.">strlen($s)</warning>;
```

Valid case:

```php
<?php
$s = Normalizer::normalize($input); if ($s !== false) { echo strlen($s); }
```

Guards written as expressions are valid too:

```php
<?php
$s = normalizer_normalize($input);
echo $s === false ? 0 : strlen($s);
echo $s !== false && strlen($s) > 3 ? 'long' : 'short';
echo match (true) { $s === false => '', default => strtoupper($s) };
if (($t = normalizer_normalize($other)) === false) { return; }
echo strlen($t);
```

## Divergences

Native custos inspection, independently specified. No EA counterpart or EA fixture conformance applies.
