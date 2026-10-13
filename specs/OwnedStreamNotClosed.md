---
id: OwnedStreamNotClosed
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "", max: "" }
---

# OwnedStreamNotClosed

## Summary

Detect an owned successfully opened stream leaving its local scope without close, return or ownership transfer.

## Detection

- D1. Report an owned successfully opened stream leaving its local scope without close, return or ownership transfer.
- D2. Resolve builtin calls and named arguments before checking the contract. Project-local callees are considered only when their bodies or immutable summaries establish the required facts. The default is disabled; enable only when the stated intent matches the application.
- D3. For flow-dependent conditions, require a reachable source-to-use path with no intervening invalidation. Unknown calls, ambiguous aliases, recursion or analysis-budget exhaustion cannot establish definite facts.
- D4. The report is made on a `return` statement whose expression only lends the handle: every mention of the local in the returned expression is an argument of a resolved builtin stream call that reads, writes, positions or inspects the stream without closing or keeping it (`fgets`, `fread`, `fwrite`, `ftell`, `feof`, `stream_get_contents`, ...).

## Exceptions (no report)

- E1. Returned handles, transferred ownership, finally cleanup, automatic-lifetime intentional policy and unknown external retention are excluded.
- E2. Unresolved targets, incomplete required facts, suppressed findings and PHP versions lacking the referenced API or language feature produce no report. User declarations shadowing builtin names are not builtin contracts.
- E3. The return expression closes the handle itself (`return fclose($h);`, `return fgets($h) . fclose($h);`).
- E4. The handle may outlive the scope. Anywhere in the function, a use of the local other than a truthiness or equality test, a plain rebinding of the local, or an argument of the borrowing builtin calls of D4 counts as a hand-over: an element of an array literal, a constructor, method, dynamic or user-function argument, an argument of any other builtin (such as `register_shutdown_function`), a copy into another variable, property, static or array slot, a reference binding, a ternary or coalesce operand, an unpacked or by-reference argument, and a capture by a closure `use` clause or an arrow function body.
- E5. A value the shared flow analysis proves is not the failure value at the use is never reported, whatever form the proof takes: a test in an enclosing branch, loop condition or short-circuit operand, an early exit, a statement that leaves on the failing outcome (`$v !== false or exit;`, `$v or throw …;`, `$v ?: throw …;`), a failing branch that ends in a call that never returns (a callee declared `never`, or a function or a final or private method every path of which throws or exits), a test stored in a boolean and tested later while neither the boolean nor the tested variable has been written, and, on PHP 8.0 and later outside any `try` with a `catch`, a builtin call that returned after receiving the value for a parameter that only accepts arrays, objects or callables: such a call throws a `TypeError` for `false`. A `switch (true)` case is read like the condition of an `if`/`elseif` chain; in a `switch` on the variable itself with constant labels, the cases after a `false` or `null` label and the `default` hold a value that is neither, as does a `true` case. Positions PHP may skip prove nothing: the right side of `??=`, the arguments of a null-safe call and anything chained after `?->`. A null-safe call to a never-returning method does not end the path, and `f(...)` creates a closure without calling `f`. An arm of `match (true)` is read like a `switch (true)` case; an arm that throws, exits or calls a function that never returns takes no part in what holds after the `match`. `|` and `&` between two tests are read like `||` and `&&` for the outcome that fixes both operands (a false `|`, a true `&`), and a strict `in_array($v, [...], true)` over literals that include neither `false` nor `null` excludes both where it holds. `goto` jumps are followed, except in a scope that also contains a `try`, which the flow analysis leaves alone. Where a `goto` label is written before the use, and after the place where its scope first names the checked variable, a variable that can hold the same value, or code that may write any variable (`extract`, `compact`, a variable variable, `eval`, an included file), a jump can land between a guard and the use: only what the flow analysis establishes counts there, and a guard recognised by where it is written proves nothing. Two variables can hold the same value when one is written from the other by a plain assignment, an assignment by reference or `??=`, when a `foreach` takes one out of the other, or through a chain of those; in the value assigned, the receiver and the arguments of a call, the operands of an operator other than `??`, the variables a string embeds and an element key do not count, and a statement that only reads (`echo $prefix . $parts['host'];`) ties nothing. A parameter is named from the start of the function, and a checked expression that reads `$this` or no variable counts every label before it. In a scope the flow analysis does not follow (one that also contains a `try`, a generator, a scope that evaluates or includes code) a guarded use after such a label is therefore reported. A label written before all of that, such as a `retry:` label ahead of the `try` that produces and checks the value, changes nothing, whatever else the function reads beside the checked value; at file level every label written before the use counts. What was established of a `static` local holds until the next call, or code that runs without a call being written (an object converted to text, a magic property or offset access, the destructor of an object a variable held when it is given another value), that may enter the function again and assign the variable. A call of PHP's own counts as such when it may run code of the project: it is given a callback other than the name of a PHP function that only computes on text and numbers (`strlen`, `trim`, `intval` and the like), an object it may convert, or a class name it loads (`method_exists($name, 'm')`, `is_callable('A::b')`, `defined('A::B')`, `new ReflectionClass($name)`), a class to construct (`$statement->fetchObject('Row')`), or it runs what was registered earlier (an error, session or output-buffer handler, the callbacks of a cURL transfer or an XML parser, an iterator wrapped in another). This limits every guard form of this specification, not only the ones the flow analysis reads.

## Report

- Range: the complete violating call or expression identified by D1; for a declaration-state violation, the offending property access or assignment. Highlight the final unsafe use for source-to-use findings.
- Severity: warning.
- Message: `Close the owned stream before leaving its scope.`

## Fix

No automatic fix: adding finally may alter exception/return interactions and resource ownership.

A fix is offered only when the concrete prerequisites above are proven. Edits use exact source byte ranges, preserve comments and argument evaluation order, and introduce syntax supported by the configured PHP version. Never invent error-handling policy.

## Options

None. Enable or disable this rule through normal rule configuration.

## PHP versions

Apply only on PHP versions supporting the referenced APIs and syntax (within PHP 5.3–8.5). Builtin behavior and argument contracts follow the configured PHP version. Extension-specific calls require resolved extension API symbols; do not treat same-named application functions as extension calls.

## Examples

```php
<?php
function bad($path) { $h = fopen($path, 'rb'); if ($h === false) { return false; } <warning descr="Close the owned stream before leaving its scope.">return fgets($h);</warning> }
function lent($path) { $h = fopen($path, 'rb'); if ($h === false) { return 0; } <warning descr="Close the owned stream before leaving its scope.">return strlen(fgets($h)) + ftell($h);</warning> }
function closed($path) { $h = fopen($path, 'rb'); if ($h === false) { return false; } return fclose($h); }
function wrapped($path) { $h = fopen($path, 'rb'); if ($h === false) { return null; } return new Reader($h); }
function paired($path) { $h = fopen($path, 'rb'); if ($h === false) { return null; } return [$h, filesize($path)]; }
```

## Divergences

Native custos rule. No upstream inspection or fixture comparison applies. These conditions are independently specified and examples are original.
