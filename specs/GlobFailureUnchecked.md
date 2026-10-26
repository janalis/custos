---
id: GlobFailureUnchecked
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "", max: "" }
---

# GlobFailureUnchecked

## Summary

Detect glob result directly or through a proven alias consumed by foreach or array-only operations without excluding false.

## Detection

Follow each producer reaching a local read, including results written on an
earlier loop turn or before a backward jump. A check of an earlier result does
not check a later replacement; a check repeated before every consumption does.
Ignore producers belonging to other result contracts when selecting an
unchecked result.

- D1. Report glob result directly or through a proven alias consumed by foreach or array-only operations without excluding false.
- D1a. Array-only operations are calls to these builtins with the glob result as their array argument, positionally or by name: `count`, `sizeof`, `array_map`, `array_filter`, `array_values`, `array_keys`, `array_reverse`, `array_slice`, `array_unique`, `array_flip`, `array_sum`, `array_product`, `array_column`, `array_chunk`, `array_key_first`, `array_key_last`, `array_is_list`, `array_rand`, `array_key_exists`, `in_array`, `array_search`, `implode`, `join`, `array_combine`, `array_fill_keys`, and the first array of `array_merge`, `array_diff` and `array_intersect`. They throw a `TypeError` for false since PHP 8.0 and warn with a meaningless result before. Each unguarded use is reported on its own.
- D2. Resolve builtin calls and named arguments before checking the contract. Project-local callees are considered only when their bodies or immutable summaries establish the required facts. The default is enabled, with positive proof required.
- D3. For flow-dependent conditions, require a reachable source-to-use path with no intervening invalidation. Unknown calls, ambiguous aliases, recursion or analysis-budget exhaustion cannot establish definite facts.

## Exceptions (no report)

- E1. Strict failure guards, explicit fallback and unknown wrapper contracts are excluded.
- E1a. Because glob yields an array or false, any condition on the same reaching value that separates the two outcomes is a guard, not only a strict comparison. Recognized conditions: the bare variable (truthiness), `empty()`, `isset()` of one of its offsets, `is_array()`, `is_iterable()`, `is_countable()`, `is_bool()`, strict or loose comparison of any of these with a `true`/`false` literal, their negation with `!`, a conjunct of `&&`/`and` that holds, and a disjunct of `||`/`or` that fails. `isset()` of the variable itself only excludes null and is not a guard.
- E1b. A guard applies when it dominates the use: the use is in the branch of an `if`/`elseif`/`else` or ternary the condition selects, in the right operand of a short-circuit operator whose left operand is the condition, or after an earlier statement of the same statement list that is either an `if` whose body always leaves (return, throw, exit, break, continue) or a `condition or exit`/`condition and throw` style expression statement. The guarded variable must still hold the same glob result at the use; a reassignment between guard and use voids the guard.
- E1c. Not treated as guards: `assert()`, a comparison stored in another variable, loop conditions and a preceding array-only call on the same value. `implode`/`join` called with the array as their only argument are not inspected.
- E2. Unresolved targets, incomplete required facts, suppressed findings and PHP versions lacking the referenced API or language feature produce no report. User declarations shadowing builtin names are not builtin contracts.
- E3. A value the shared flow analysis proves is not the failure value at the use is never reported, whatever form the proof takes: a test in an enclosing branch, loop condition or short-circuit operand, an early exit, a statement that leaves on the failing outcome (`$v !== false or exit;`, `$v or throw …;`, `$v ?: throw …;`), a failing branch that ends in a call that never returns (a callee declared `never`, a function or a final or private method every path of which throws or exits, or, in a class that extends PHPUnit's `TestCase` when PHPUnit itself is not indexed, `fail()`, `markTestSkipped()` or `markTestIncomplete()` called on `$this`, `self`, `static` or `parent`), a statement calling an assertion method of PHPUnit or `webmozart/assert` whose documented assertion excludes the failure value (`assertNotFalse()`, `assertIsResource()`, `assertIsArray()`), which throws when it fails (an assertion another function or method documents is no guard), a test stored in a boolean and tested later while neither the boolean nor the tested variable has been written, and, on PHP 8.0 and later outside any `try` with a `catch`, a builtin call that returned after receiving the value for a parameter that only accepts arrays, objects (any class, such as `GdImage`) or callables, or one PHP checks as a resource (the handle of `fread()` or `fwrite()`): such a call throws a `TypeError` for `false`. A `switch (true)` case is read like the condition of an `if`/`elseif` chain; in a `switch` on the variable itself with constant labels, the cases after a `false` or `null` label and the `default` hold a value that is neither, as does a `true` case. Positions PHP may skip prove nothing: the right side of `??=`, the arguments of a null-safe call and anything chained after `?->`. A null-safe call to a never-returning method does not end the path, and `f(...)` creates a closure without calling `f`. An arm of `match (true)` is read like a `switch (true)` case; an arm that throws, exits or calls a function that never returns takes no part in what holds after the `match`. `|` and `&` between two tests are read like `||` and `&&` for the outcome that fixes both operands (a false `|`, a true `&`), and a strict `in_array($v, [...], true)` over literals that include neither `false` nor `null` excludes both where it holds. `goto` jumps are followed, including a jump out of a `try` or `catch` block, which runs the `finally` blocks it leaves on its way. Where a `goto` label is written before the use, and after the place where its scope first names the checked variable, a variable that can hold the same value, or code that may write any variable (`extract`, `compact`, `get_defined_vars`, `parse_str` given one argument, a variable variable, `$GLOBALS`, `eval`, an included file), a jump can land between a guard and the use: only what the flow analysis establishes counts there, and a guard recognized by where it is written proves nothing. Two variables can hold the same value when one is written from the other by a plain assignment, an assignment by reference or `??=`, when a `foreach` takes one out of the other, or through a chain of those; in the value assigned, the receiver and the arguments of a call, the operands of an operator other than `??`, the variables a string embeds and an element key do not count, and a statement that only reads (`echo $prefix . $parts['host'];`) ties nothing. A parameter is named from the start of the function, and a checked expression that reads `$this` or no variable counts every label before it. In a scope the flow analysis does not follow (one that evaluates or includes code, which may write any variable) a guarded use after such a label is therefore reported. A generator is followed: while it is suspended at a `yield`, the code that consumes it may write only what an unknown call may write, and, when the generator returns by reference, the variable it yields. A label written before all of that, such as a `retry:` label ahead of the `try` that produces and checks the value, changes nothing, whatever else the function reads beside the checked value; at file level every label written before the use counts. What was established of a `static` local holds until the next call, or code that runs without a call being written (an object converted to text, a magic property or offset access, the destructor of an object a variable held when it is given another value), that may enter the function again and assign the variable. A call of PHP's own counts as such when it may run code of the project: it is given a callback other than the name of a PHP function that only computes on text and numbers and raises no warning whatever text it is given (a closed list: `strlen`, `trim`, `intval` and the like, not `hexdec`, `abs` or `chr`, which warn about text that is no number), an object it may convert, a class name it loads (`method_exists($name, 'm')`, `is_callable('A::b')`, `defined('A::B')`, `class_alias($name, 'B')`, `new ReflectionClass($name)`, `$arrayObject->setIteratorClass($name)`), a class to construct (`$statement->fetchObject('Row')`) or a filter that one of PHP's own `FILTER_…` constants does not name, which may be the one that calls back, or it runs what was registered earlier (an error, session or output-buffer handler, the callbacks of a cURL transfer or an XML parser, an iterator wrapped in another, a filter added to a stream). A warning that a PHP function raises by itself when called directly (`hex2bin` given an odd number of digits, a file that cannot be read) is not counted, although the error handler of the project receives it. This limits every guard form of this specification, not only the ones the flow analysis reads.

## Report

- Range: the complete violating call or expression identified by D1; for a declaration-state violation, the offending property access or assignment. Highlight the final unsafe use for source-to-use findings.
- Severity: warning.
- Message: `Handle glob failure before iterating its result.` for foreach; `Handle glob failure before using its result as an array.` for an array-only call (range: the argument).

## Fix

No automatic fix: failure and an empty match set must remain distinguishable unless caller policy establishes otherwise.

## Options

None. Enable or disable this rule through normal rule configuration.

## PHP versions

Apply only on PHP versions supporting the referenced APIs and syntax (within PHP 5.3–8.5). Builtin behavior and argument contracts follow the configured PHP version. Extension-specific calls require resolved extension API symbols; do not treat same-named application functions as extension calls.

## Examples

```php
<?php
function bad($pattern) { foreach (<warning descr="Handle glob failure before iterating its result.">glob($pattern)</warning> as $path) { echo $path; } }
function names($pattern) {
    $paths = glob($pattern);
    return array_map('basename', <warning descr="Handle glob failure before using its result as an array.">$paths</warning>);
}
function guarded($pattern) {
    $paths = glob($pattern);
    if (!$paths) {
        return 0;
    }
    return count($paths);
}
```

### Joined producers

When several assignments reach a use, each producer is checked on its own: an
unguarded eligible producer is still reported when other paths use another API
or are guarded, and an irrelevant variant cannot hide a later eligible call. A
path whose value does not come from a known call establishes nothing. When the
producers cannot be tracked within the analysis budget, nothing is reported.

## Divergences

Native custos rule. No upstream inspection or fixture comparison applies. These conditions are independently specified and examples are original.
