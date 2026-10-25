---
id: PosixAccountLookupFailureDereferenced
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "", max: "" }
---

# PosixAccountLookupFailureDereferenced

## Summary

Identify array indexing on a direct or unchanged posix_getpwnam/posix_getpwuid result that a failed lookup (`false`) can still reach.

## Detection

- D1. Report array indexing on a direct or unchanged posix_getpwnam/posix_getpwuid result unless a guard proves the failure value `false` cannot reach the read (see D4).
- D2. Resolve builtin functions, classes, constants and methods, including imports and supported named arguments. Follow only bounded local facts within one lexical scope. Receiver and argument provenance must establish the specific API contract; variable names and comments never prove intent.
- D3. Report one finding per violating operation. A reachable path, unchanged value identity and applicable API behavior are required. Intervening mutation, unknown calls involving tracked values, escaped aliases or exhausted analysis budgets invalidate proof.
- D3a. When several assignments reach the read (a loop that reassigns the variable, or an assignment made on one path only), the read is reported if one of them is a result of this API that no guard on its own path covers. A guard written before the other assignment does not cover it; a test inside the loop, or after the conditional assignment, covers every assignment. A reaching value that is not a result of this API (a literal, another call, an unknown value) never causes a report by itself. It does not hide one either: an unguarded result of another call reaching the same read leaves a result of this API reported.

- D4. A guard is a test of the same unchanged variable, written after the lookup, whose outcome on the path to the read excludes `false`. Accepted tests, each usable positively or through `!`:
  - comparison with `false` using `===`, `!==`, `==` or `!=`, in either operand order;
  - truthiness of the variable itself, or of the assignment `$v = posix_getpwnam(...)` used as a condition (a successful lookup is a non-empty array);
  - `is_array($v)`, `isset($v[...])` with at least one key (plain `isset($v)` accepts `false`), and a false `empty($v)` or `empty($v[...])`;
  - `&&`/`and` and `||`/`or` combinations of the above, where a conjunction holds if either side proves it and a disjunction only if both do (and dually when the combination is false).

  A guard applies to the read when the read sits in the matching `if`/`elseif`/`else` branch or ternary arm, in the right operand of a short-circuit operator whose left operand is the guard, or after an earlier statement of an enclosing statement list that leaves it on failure: an `if` whose body always terminates without a `goto` (which could land later in the same list), or an expression statement `guard or throw ...` / `guard or exit` (respectively `failure and throw/exit`). A direct `posix_getpwuid(...)[...]` read has no variable to test and is always reported.

## Exceptions (no report)

- Array and list destructuring destinations create slots rather than reading them. Destructuring keys and ordinary array values still require reads.

- Guarded probes with isset, empty or null coalescing, unsets and plain nested slot creation do not read a failed or missing value. Index expressions and the fallback operand of null coalescing still require normal reads.

- E1. Exclude unresolved required symbols, user-defined lookalikes, incomplete syntax, unknown required values and unsupported versions. Cases outside D1 receive no finding.
- E2. Both `@custos-ignore PosixAccountLookupFailureDereferenced` and `@noinspection PosixAccountLookupFailureDereferenced` suppress this native rule.
- E3. A value the shared flow analysis proves is not the failure value at the use is never reported, whatever form the proof takes: a test in an enclosing branch, loop condition or short-circuit operand, an early exit, a statement that leaves on the failing outcome (`$v !== false or exit;`, `$v or throw …;`, `$v ?: throw …;`), a failing branch that ends in a call that never returns (a callee declared `never`, a function or a final or private method every path of which throws or exits, or, in a class that extends PHPUnit's `TestCase` when PHPUnit itself is not indexed, `fail()`, `markTestSkipped()` or `markTestIncomplete()` called on `$this`, `self`, `static` or `parent`), a test stored in a boolean and tested later while neither the boolean nor the tested variable has been written, and, on PHP 8.0 and later outside any `try` with a `catch`, a builtin call that returned after receiving the value for a parameter that only accepts arrays, objects or callables, or one PHP checks as a resource (the handle of `fread()` or `fwrite()`): such a call throws a `TypeError` for `false`. A `switch (true)` case is read like the condition of an `if`/`elseif` chain; in a `switch` on the variable itself with constant labels, the cases after a `false` or `null` label and the `default` hold a value that is neither, as does a `true` case. Positions PHP may skip prove nothing: the right side of `??=`, the arguments of a null-safe call and anything chained after `?->`. A null-safe call to a never-returning method does not end the path, and `f(...)` creates a closure without calling `f`. An arm of `match (true)` is read like a `switch (true)` case; an arm that throws, exits or calls a function that never returns takes no part in what holds after the `match`. `|` and `&` between two tests are read like `||` and `&&` for the outcome that fixes both operands (a false `|`, a true `&`), and a strict `in_array($v, [...], true)` over literals that include neither `false` nor `null` excludes both where it holds. `goto` jumps are followed, including a jump out of a `try` or `catch` block, which runs the `finally` blocks it leaves on its way. Where a `goto` label is written before the use, and after the place where its scope first names the checked variable, a variable that can hold the same value, or code that may write any variable (`extract`, `compact`, `get_defined_vars`, `parse_str` given one argument, a variable variable, `$GLOBALS`, `eval`, an included file), a jump can land between a guard and the use: only what the flow analysis establishes counts there, and a guard recognized by where it is written proves nothing. Two variables can hold the same value when one is written from the other by a plain assignment, an assignment by reference or `??=`, when a `foreach` takes one out of the other, or through a chain of those; in the value assigned, the receiver and the arguments of a call, the operands of an operator other than `??`, the variables a string embeds and an element key do not count, and a statement that only reads (`echo $prefix . $parts['host'];`) ties nothing. A parameter is named from the start of the function, and a checked expression that reads `$this` or no variable counts every label before it. In a scope the flow analysis does not follow (one that evaluates or includes code, which may write any variable) a guarded use after such a label is therefore reported. A generator is followed: while it is suspended at a `yield`, the code that consumes it may write only what an unknown call may write, and, when the generator returns by reference, the variable it yields. A label written before all of that, such as a `retry:` label ahead of the `try` that produces and checks the value, changes nothing, whatever else the function reads beside the checked value; at file level every label written before the use counts. What was established of a `static` local holds until the next call, or code that runs without a call being written (an object converted to text, a magic property or offset access, the destructor of an object a variable held when it is given another value), that may enter the function again and assign the variable. A call of PHP's own counts as such when it may run code of the project: it is given a callback other than the name of a PHP function that only computes on text and numbers and raises no warning whatever text it is given (a closed list: `strlen`, `trim`, `intval` and the like, not `hexdec`, `abs` or `chr`, which warn about text that is no number), an object it may convert, a class name it loads (`method_exists($name, 'm')`, `is_callable('A::b')`, `defined('A::B')`, `class_alias($name, 'B')`, `new ReflectionClass($name)`, `$arrayObject->setIteratorClass($name)`), a class to construct (`$statement->fetchObject('Row')`) or a filter that one of PHP's own `FILTER_…` constants does not name, which may be the one that calls back, or it runs what was registered earlier (an error, session or output-buffer handler, the callbacks of a cURL transfer or an XML parser, an iterator wrapped in another, a filter added to a stream). A warning that a PHP function raises by itself when called directly (`hex2bin` given an odd number of digits, a file that cannot be read) is not counted, although the error handler of the project receives it. This limits every guard form of this specification, not only the ones the flow analysis reads.
- E4. A property read is guarded like a variable when its earlier test and the read give the same value: the read runs no code (every class the receiver may be of declares the property without a hook, and has no `__get`; a class extending it that declares the property again with a hook is not considered), and between them no call that may run project code, no write, `unset`, increment, `foreach` target or reference of a property of that name through any receiver, and no write of the receiver variable can change it; a loop around the read and not the test counts its whole body.

## Report

- Range: the complete violating call or expression. Follow D1 if it specifies a narrower range.
- Severity: warning.
- Enabled by default: true.
- Message: `Reject account lookup failure before indexing.`

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
$account = posix_getpwnam($name); echo $account['uid'];
```

Valid case:

```php
<?php
$account = posix_getpwnam($name); if ($account !== false) { echo $account['uid']; }
```

Also valid, the failure never reaches the read:

```php
<?php
$owner = posix_getpwuid($uid) or exit('unknown owner');
echo $owner['name'];
$home = is_array($owner) ? $owner['dir'] : null;
```

### Joined producers

When several assignments reach a use, each producer is checked on its own: an
unguarded eligible producer is still reported when other paths use another API
or are guarded, and an irrelevant variant cannot hide a later eligible call. A
path whose value does not come from a known call establishes nothing. When the
producers cannot be tracked within the analysis budget, nothing is reported.

## Divergences

Native custos inspection, independently specified; no EA counterpart or EA conformance requirement. Conservative proof intentionally excludes unknown runtime state and intent.
