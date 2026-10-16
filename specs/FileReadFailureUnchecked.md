---
id: FileReadFailureUnchecked
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "", max: "" }
---

# FileReadFailureUnchecked

## Summary

Detect file_get_contents or stream_get_contents result consumed as a string without excluding false.

## Detection

Bounded, pure constant conditions select only their executable flow successor.
The same applies to ternary arms. User constants, calls and unsupported scalar
coercions remain unknown; labels inside skipped branches can still be entered
by a legal jump. Lexical reachability queries retain their existing contract.

Follow each producer reaching a local read, including results written on an
earlier loop turn or before a backward jump. A check of an earlier result does
not check a later replacement; a check repeated before every consumption does.
Ignore producers belonging to other result contracts when selecting an
unchecked result.

- D1. Report file_get_contents or stream_get_contents result consumed as a string without excluding false.
- D1a. A result is consumed as a string when it is an operand of a concatenation, or an argument bound to a parameter declared exactly `string` by the callee the call resolves to, when that declaration is the one that runs: a function declared once; a static method named by its class, `self` or `parent`; a method of a receiver whose class is known, when the method is final or private, the class is final, the receiver is a variable holding an object made by `new` in the same scope, or no class of the analysed project that extends the receiver's class replaces the method. When such classes do replace it, the parameter still counts if every replacement declares it `string` under the same name, at the same position, or wherever it stands when the argument is passed by name. Each class involved must be declared once in the analysed paths (two copies of a library are two declarations, either of which may be loaded), and the method must be visible where the call is written. An untyped, nullable or union-typed parameter, a `string` parameter whose default value is `null` (which makes it nullable), a call whose callee cannot be resolved, a variable receiver the flow analysis does not follow, and any declaration of the project on a PHP target below 7.0, where `string` names a class, consume nothing. Nor does an assertion or expectation of PHPUnit: a method whose name starts with `assert` or `expect` and that the call finds in PHPUnit's own `Assert` or `TestCase` class (reached through `$this`, a typed object, `self`, `static`, `parent` or the class name), or an `assert…` function of the `PHPUnit\Framework` namespace. Such a call is the test's check of the value: it fails, or throws under strict types, when handed the failure value, save for a degenerate comparison with empty text. The exclusion reads only which class declares the method, so it does not depend on how many copies of PHPUnit the analysed paths hold. A method the project declares on its own test classes, a helper as much as an `assert…` method of its own, remains an ordinary callee.
- D2. Resolve builtin calls and named arguments before checking the contract. Project-local callees are considered only when their bodies or immutable summaries establish the required facts. The default is enabled, with positive proof required.
- D3. For flow-dependent conditions, require a reachable source-to-use path with no intervening invalidation. Unknown calls, ambiguous aliases, recursion or analysis-budget exhaustion cannot establish definite facts.

## Exceptions (no report)

- E1. Strict false guard, deliberate documented false coercion, termination, and result reassignment invalidate the report.
- E2. Unresolved targets, incomplete required facts, suppressed findings and PHP versions lacking the referenced API or language feature produce no report. User declarations shadowing builtin names are not builtin contracts.
- E3. A guard on the same read result that excludes false counts as handling, whether it selects the branch holding the use (`if`, `elseif`/`else`, ternary, the right operand of `&&`/`||`) or sits in an earlier sibling `if` without alternatives whose body leaves the path. Accepted conditions, possibly negated or combined with `&&`/`||` where the combination still proves the fact: truthiness, `empty()`, `is_string()`, strict comparison with `false`, and loose comparison with a constant that `false` equals (`false`, `null`, `0`, `''`, `'0'`). The guarded variable must hold only that earlier read, or be assigned it inside the condition.
- E4. A guard body leaves the path when it terminates syntactically or contains, as a direct statement, a call whose resolved return type is `never`. Calls with unknown or returning targets do not qualify.
- E5. `$text = read() or <alternative>` (also `||` around the parenthesized assignment) handles failure when the alternative is `exit`/`die`, a `throw` expression or a call resolved as `never`; uses inside the alternative itself stay reported.
- E6. A value the shared flow analysis proves is not the failure value at the use is never reported, whatever form the proof takes: a test in an enclosing branch, loop condition or short-circuit operand, an early exit, a statement that leaves on the failing outcome (`$v !== false or exit;`, `$v or throw …;`, `$v ?: throw …;`), a failing branch that ends in a call that never returns (a callee declared `never`, or a function or a final or private method every path of which throws or exits), a test stored in a boolean and tested later while neither the boolean nor the tested variable has been written, and, on PHP 8.0 and later outside any `try` with a `catch`, a builtin call that returned after receiving the value for a parameter that only accepts arrays, objects or callables: such a call throws a `TypeError` for `false`. A `switch (true)` case is read like the condition of an `if`/`elseif` chain; in a `switch` on the variable itself with constant labels, the cases after a `false` or `null` label and the `default` hold a value that is neither, as does a `true` case. Positions PHP may skip prove nothing: the right side of `??=`, the arguments of a null-safe call and anything chained after `?->`. A null-safe call to a never-returning method does not end the path, and `f(...)` creates a closure without calling `f`. An arm of `match (true)` is read like a `switch (true)` case; an arm that throws, exits or calls a function that never returns takes no part in what holds after the `match`. `|` and `&` between two tests are read like `||` and `&&` for the outcome that fixes both operands (a false `|`, a true `&`), and a strict `in_array($v, [...], true)` over literals that include neither `false` nor `null` excludes both where it holds. `goto` jumps are followed, except in a scope that also contains a `try`, which the flow analysis leaves alone. Where a `goto` label is written before the use, and after the place where its scope first names the checked variable, a variable that can hold the same value, or code that may write any variable (`extract`, `compact`, `get_defined_vars`, `parse_str` given one argument, a variable variable, `$GLOBALS`, `eval`, an included file), a jump can land between a guard and the use: only what the flow analysis establishes counts there, and a guard recognised by where it is written proves nothing. Two variables can hold the same value when one is written from the other by a plain assignment, an assignment by reference or `??=`, when a `foreach` takes one out of the other, or through a chain of those; in the value assigned, the receiver and the arguments of a call, the operands of an operator other than `??`, the variables a string embeds and an element key do not count, and a statement that only reads (`echo $prefix . $parts['host'];`) ties nothing. A parameter is named from the start of the function, and a checked expression that reads `$this` or no variable counts every label before it. In a scope the flow analysis does not follow (one that also contains a `try`, a generator, a scope that evaluates or includes code) a guarded use after such a label is therefore reported. A label written before all of that, such as a `retry:` label ahead of the `try` that produces and checks the value, changes nothing, whatever else the function reads beside the checked value; at file level every label written before the use counts. What was established of a `static` local holds until the next call, or code that runs without a call being written (an object converted to text, a magic property or offset access, the destructor of an object a variable held when it is given another value), that may enter the function again and assign the variable. A call of PHP's own counts as such when it may run code of the project: it is given a callback other than the name of a PHP function that only computes on text and numbers and raises no warning whatever text it is given (a closed list: `strlen`, `trim`, `intval` and the like, not `hexdec`, `abs` or `chr`, which warn about text that is no number), an object it may convert, a class name it loads (`method_exists($name, 'm')`, `is_callable('A::b')`, `defined('A::B')`, `class_alias($name, 'B')`, `new ReflectionClass($name)`, `$arrayObject->setIteratorClass($name)`), a class to construct (`$statement->fetchObject('Row')`) or a filter that one of PHP's own `FILTER_…` constants does not name, which may be the one that calls back, or it runs what was registered earlier (an error, session or output-buffer handler, the callbacks of a cURL transfer or an XML parser, an iterator wrapped in another). A warning that a PHP function raises by itself when called directly (`hex2bin` given an odd number of digits, a file that cannot be read) is not counted, although the error handler of the project receives it. This limits every guard form of this specification, not only the ones the flow analysis reads.

A temporary array argument cannot rebind variables copied into its entries. An unconditional unset releases a former reference binding for subsequent local assignments within that statement list; conditional releases, a new reference binding and jumps that can skip the release do not establish that independence. Unresolved whole-variable arguments may be rebound and therefore do not establish a retained failure producer. Exhausting the bounded local write or reference-lifetime work allowance abandons the corresponding proof.

## Report

- Range: the complete violating call or expression identified by D1; for a declaration-state violation, the offending property access or assignment. Highlight the final unsafe use for source-to-use findings.
- Severity: warning.
- Message: `Handle file reading failure before using its result.`

## Fix

No automatic fix: read failure handling belongs to the caller.

A fix is offered only when the concrete prerequisites above are proven. Edits use exact source byte ranges, preserve comments and argument evaluation order, and introduce syntax supported by the configured PHP version. Never invent error-handling policy.

## Options

None. Enable or disable this rule through normal rule configuration.

## PHP versions

Apply only on PHP versions supporting the referenced APIs and syntax (within PHP 5.3–8.5). Builtin behavior and argument contracts follow the configured PHP version. Extension-specific calls require resolved extension API symbols; do not treat same-named application functions as extension calls.

## Examples

```php
<?php
$text = <warning descr="Handle file reading failure before using its result.">trim(file_get_contents($path))</warning>;
```

Not reported:

```php
<?php
function abort(): never { throw new RuntimeException('unreadable'); }

$text = file_get_contents($path) or die('unreadable');
$size = strlen($text);

$other = file_get_contents($backup);
if ($other == false) {
    abort();
}
$size += strlen($other);
```

### Assertion assumptions

Only plain variable tests, literals, boolean operators, strict comparisons,
instanceof, isset/empty and native type predicates supply assertion guards.
A direct native flock result can prove that operation succeeded.
Conditions with property/array accesses, evaluated code, conversions or other
effects supply none; an earlier test may precede a rewrite.

A direct native `assert()` call is a guard only when the resolved project
setting `flow.assertions` is true. Its default is false, since PHP can disable
assertions. On PHP 7 and later, writes inside assertion arguments may be skipped:
a sanitizer or cast written there supplies no definite safety proof in the
default mode. Dynamic callable invocations evaluate arguments normally and do
not supply an assertion guard. With the setting enabled, failed assertions are
assumed to terminate. A later description expression that can rewrite the tested
value prevents using the condition as a guard. Before PHP 8, evaluated string
assertions invalidate local proofs; their PHP code is not interpreted.

## Divergences

Native custos rule. No upstream inspection or fixture comparison applies. These conditions are independently specified and examples are original.
