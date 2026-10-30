---
id: PregReplaceFailureUnchecked
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "", max: "" }
---

# PregReplaceFailureUnchecked

## Summary

Detect a nullable result of preg_replace or preg_replace_callback flowing into a non-null string parameter or dereference without a successful null guard.

## Detection

Follow each producer reaching a local read, including results written on an
earlier loop turn or before a backward jump. A check of an earlier result does
not check a later replacement; a check repeated before every consumption does.
Ignore producers belonging to other result contracts when selecting an
unchecked result.

A sentinel or length guard applies only to the same reaching value. Reassignment or reevaluating an identical-looking call invalidates the earlier guard.

- D1. Report a nullable result of preg_replace or preg_replace_callback flowing into a non-null string parameter or dereference without a successful null guard.
- D1a. A non-null string parameter is a parameter declared exactly `string` by the callee the call resolves to (for a function or method PHP itself declares, documented as exactly `string`), when that declaration is the one that runs: a function declared once, or declared several times with the same parameters; a static method named by its class, `self` or `parent`; a method of a receiver whose class is known, when the method is final or private, the class is final, the receiver is a variable holding an object made by `new` in the same scope, or no class of the analyzed project that extends the receiver's class replaces the method. When such classes do replace it, the parameter still counts if every replacement declares it `string` under the same name, at the same position, or wherever it stands when the argument is passed by name. Each class involved must be declared once in the analyzed paths (two copies of a library are two declarations, either of which may be loaded), except the class the method is looked up on: its copies count when each declares the method itself, which must then take the parameter as exactly `string` in every copy, and the method must be visible where the call is written (otherwise `__call` or `__callStatic` receives it). An untyped, nullable or union-typed parameter, a `string` parameter whose default value is `null` (which makes it nullable), a call whose callee cannot be resolved (a receiver of unknown class, a computed name), a variable receiver the flow analysis does not follow, and any declaration of the project on a PHP target below 7.0, where `string` names a class, consume nothing. The finding is placed on the call that receives the value. When the declaration that runs is exactly known (not merely agreed on by its replacements) and no enclosing `try` has a `catch`, the same result read again after that call is not reported a second time, since the call does not return for null. An assertion or expectation of PHPUnit is not such a use: a method whose name starts with `assert` or `expect` and that the call finds in PHPUnit's own `Assert` or `TestCase` class (reached through `$this`, a typed object, `self`, `static`, `parent` or the class name), or an `assert…` function of the `PHPUnit\Framework` namespace. Such a call is the test's check of the value: handed null for a `string` parameter it throws, which fails the test. The exclusion reads only which class declares the method, so it does not depend on how many copies of PHPUnit the analyzed paths hold. Such an assertion still uses the result when it would pass with null: compared by `assertEquals`, `assertEqualsCanonicalizing` or `assertEqualsIgnoringCase` with a literal that equals null loosely (false, null, the empty text, zero; not `"0"`, which PHP compares with null as text), or given to `assertEmpty`. A string needle rejects null, so the failure there is reported by the test itself. A static call of webmozart/assert's `Webmozart\Assert\Assert`, matched by the class name written, is such a check too, and uses the result the same way when it would pass with null: `eq` with a literal that equals null loosely, `isEmpty`, and any `nullOr…` assertion, which accepts null. Assertions of other libraries are ordinary calls. A method the project declares on its own test classes, a helper as much as an `assert…` method of its own, remains an ordinary callee.
- D2. Resolve builtin calls and named arguments before checking the contract. Project-local callees are considered only when their bodies or immutable summaries establish the required facts. The default is enabled, with positive proof required.
- D3. For flow-dependent conditions, require a reachable source-to-use path with no intervening invalidation. Unknown calls, ambiguous aliases, recursion or analysis-budget exhaustion cannot establish definite facts.

## Exceptions (no report)

- E1. Explicit null fallback, guard, termination, and known non-PCRE callees are excluded. The guard does not depend on the declared type of the subject; a result typed `string|array|null` is guarded the same way as `?string`. A use is guarded when a null test on the same reaching value dominates it:
  - the test is `$v !== null` or `isset($v)` or `is_string($v)` holding true, or `$v === null` or `is_null($v)` holding false, with either operand order, optional parentheses, `!` negation, and the tested value optionally assigned in place (`null !== $v = preg_replace(...)`);
  - a conjunction (`&&`, `and`) holding true proves what either operand proves; a disjunction (`||`, `or`) holding false proves what either operand proves;
  - the test dominates the use as the condition of an enclosing `if`/`elseif` body, as a failed earlier condition of the same `if`/`elseif` chain for its later arms, as the condition of an enclosing ternary arm, as the left operand of the short-circuit operator whose right operand contains the use, or as the condition of an earlier `if` in an enclosing statement list whose body always terminates.
- E3. Tests of another variable, and guards followed by a reassignment or a new call prove nothing; the use is still reported. A loose comparison with `null` (`$v == null` leaving the path, `$v != null` selecting the branch) and general truthiness (`if ($v)`) do exclude null and are accepted like the strict forms.
- E2. Unresolved targets, incomplete required facts, suppressed findings and PHP versions lacking the referenced API or language feature produce no report. User declarations shadowing builtin names are not builtin contracts.
- E4. A value the shared flow analysis proves is not the failure value at the use is never reported, whatever form the proof takes: a test in an enclosing branch, loop condition or short-circuit operand, an early exit, a statement that leaves on the failing outcome (`$v !== false or exit;`, `$v or throw …;`, `$v ?: throw …;`), a failing branch that ends in a call that never returns (a callee declared `never`, a function or a final or private method every path of which throws or exits, or, in a class that extends PHPUnit's `TestCase` when PHPUnit itself is not indexed, `fail()`, `markTestSkipped()` or `markTestIncomplete()` called on `$this`, `self`, `static` or `parent`), a test stored in a boolean and tested later while neither the boolean nor the tested variable has been written, and, on PHP 8.0 and later outside any `try` with a `catch`, a builtin call that returned after receiving the value for a parameter that only accepts arrays, objects (any class, such as `GdImage`) or callables, or one PHP checks as a resource (the handle of `fread()` or `fwrite()`): such a call throws a `TypeError` for `false`. A `switch (true)` case is read like the condition of an `if`/`elseif` chain; in a `switch` on the variable itself with constant labels, the cases after a `false` or `null` label and the `default` hold a value that is neither, as does a `true` case. Positions PHP may skip prove nothing: the right side of `??=`, the arguments of a null-safe call and anything chained after `?->`. A null-safe call to a never-returning method does not end the path, and `f(...)` creates a closure without calling `f`. An arm of `match (true)` is read like a `switch (true)` case; an arm that throws, exits or calls a function that never returns takes no part in what holds after the `match`. `|` and `&` between two tests are read like `||` and `&&` for the outcome that fixes both operands (a false `|`, a true `&`), and a strict `in_array($v, [...], true)` over literals that include neither `false` nor `null` excludes both where it holds. `goto` jumps are followed, including a jump out of a `try` or `catch` block, which runs the `finally` blocks it leaves on its way. Where a `goto` label is written before the use, and after the place where its scope first names the checked variable, a variable that can hold the same value, or code that may write any variable (`extract`, `compact`, `get_defined_vars`, `parse_str` given one argument, a variable variable, `$GLOBALS`, `eval`, an included file), a jump can land between a guard and the use: only what the flow analysis establishes counts there, and a guard recognized by where it is written proves nothing. Two variables can hold the same value when one is written from the other by a plain assignment, an assignment by reference or `??=`, when a `foreach` takes one out of the other, or through a chain of those; in the value assigned, the receiver and the arguments of a call, the operands of an operator other than `??`, the variables a string embeds and an element key do not count, and a statement that only reads (`echo $prefix . $parts['host'];`) ties nothing. A parameter is named from the start of the function, and a checked expression that reads `$this` or no variable counts every label before it. In a scope the flow analysis does not follow (one that evaluates or includes code, which may write any variable) a guarded use after such a label is therefore reported. A generator is followed: while it is suspended at a `yield`, the code that consumes it may write only what an unknown call may write, and, when the generator returns by reference, the variable it yields. A label written before all of that, such as a `retry:` label ahead of the `try` that produces and checks the value, changes nothing, whatever else the function reads beside the checked value; at file level every label written before the use counts. What was established of a `static` local holds until the next call, or code that runs without a call being written (an object converted to text, a magic property or offset access, the destructor of an object a variable held when it is given another value), that may enter the function again and assign the variable. A call of PHP's own counts as such when it may run code of the project: it is given a callback other than the name of a PHP function that only computes on text and numbers and raises no warning whatever text it is given (a closed list: `strlen`, `trim`, `intval` and the like, not `hexdec`, `abs` or `chr`, which warn about text that is no number), an object it may convert, a class name it loads (`method_exists($name, 'm')`, `is_callable('A::b')`, `defined('A::B')`, `class_alias($name, 'B')`, `new ReflectionClass($name)`, `$arrayObject->setIteratorClass($name)`), a class to construct (`$statement->fetchObject('Row')`) or a filter that one of PHP's own `FILTER_…` constants does not name, which may be the one that calls back, or it runs what was registered earlier (an error, session or output-buffer handler, the callbacks of a cURL transfer or an XML parser, an iterator wrapped in another, a filter added to a stream). A warning that a PHP function raises by itself when called directly (`hex2bin` given an odd number of digits, a file that cannot be read) is not counted, although the error handler of the project receives it. This limits every guard form of this specification, not only the ones the flow analysis reads.

A resolved native string test also excludes null when its outcome cannot hold after null becomes an empty string: a true prefix, suffix or containment test with a nonempty literal needle; a position or substring search (`strpos`, `stripos`, `strrpos`, `strripos`, `strstr`, `stristr`, `strrchr`) for a nonempty literal needle whose result is truthy or not false; or a strict comparison of a string offset with a literal string. Haystack and needle may be passed by position or by name; an argument given twice proves nothing. The tested binding must remain unchanged on every route to the use, which may also be reached through a merge with a value of unknown type. Empty needles, shadowed functions, rewrites with a value that may be null and jumps that bypass the test prove nothing; a rewrite with a string (a literal, `substr()`) needs no test. A test whose outcome the condition stores in another variable (`if ($negated = str_starts_with($line, '!'))`) counts as the test itself.

## Report

- Range: the complete violating call or expression identified by D1; for a declaration-state violation, the offending property access or assignment. Highlight the final unsafe use for source-to-use findings.
- Severity: warning.
- Message: `Handle regex replacement failure before using its result.`

## Fix

No automatic fix: the correct handling of regex failure is unknown.

## Options

None. Enable or disable this rule through normal rule configuration.

## PHP versions

Apply only on PHP versions supporting the referenced APIs and syntax (within PHP 5.3–8.5). Builtin behavior and argument contracts follow the configured PHP version. Extension-specific calls require resolved extension API symbols; do not treat same-named application functions as extension calls.

## Examples

```php
<?php
$text = <warning descr="Handle regex replacement failure before using its result.">strtoupper(preg_replace('/z/u', '', $input))</warning>;
```

```php
<?php
// Guarded: no report, whatever the declared type of $input.
$clean = preg_replace('/\s+/u', ' ', $input);
echo is_string($clean) ? strtoupper($clean) : '';
```

### Joined producers

When several assignments reach a use, each producer is checked on its own: an
unguarded eligible producer is still reported when other paths use another API
or are guarded, and an irrelevant variant cannot hide a later eligible call. A
path whose value does not come from a known call establishes nothing. When the
producers cannot be tracked within the analysis budget, nothing is reported.

## Divergences

Native custos rule. No upstream inspection or fixture comparison applies. These conditions are independently specified and examples are original.
