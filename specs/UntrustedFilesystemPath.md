---
id: UntrustedFilesystemPath
group: Security
kind: semantic
needs: [names, types, flow]
php: { min: "", max: "" }
---

# UntrustedFilesystemPath

## Summary

Detect request-derived path components reaching file read/write/delete or include sinks without established canonical containment or finite filename selection.

## Detection

- D1. Report request-derived path components reaching file read/write/delete or include sinks without established canonical containment or finite filename selection.
- D2. Resolve builtin calls and named arguments before checking the contract. Project-local callees are considered only when their bodies or immutable summaries establish the required facts. The default is enabled, with positive proof required.
- D3. For flow-dependent conditions, require a reachable source-to-use path with no intervening invalidation. Unknown calls, ambiguous aliases, recursion or analysis-budget exhaustion cannot establish definite facts.

## Exceptions (no report)

- E1. Proven boundary-aware containment with explicit symlink policy, fixed paths, validated finite names and unknown flows are excluded.
- E1a. Canonical containment is proven for a variable only when all of the following hold. Anything weaker keeps the report.
  - The variable's single reaching value at the sink is a call to builtin `realpath()`, so symlinks and `..` are already resolved.
  - A prefix test on that same value (same reaching assignment) has succeeded on every path to the sink: the sink is in the body of an `if` whose condition being true implies the test, or an earlier sibling `if` whose body always leaves (return, throw, exit, …) runs when the test fails. Implication looks through parentheses, `!`, any operand of `&&` (true side) and any operand of `||` (false side); an `else` branch or a non-leaving body proves nothing.
  - The test is `str_starts_with($p, PREFIX)`, `strpos($p, PREFIX) === 0` without an offset (or `!== 0` on the rejecting side, operands in either order), or `strncmp($p, PREFIX, N) === 0` with an integer `N` not smaller than the byte length of a literal `PREFIX`. Loose comparison is not accepted: `strpos() == 0` is also true for `false`.
  - `PREFIX` is a string literal longer than one byte that starts and ends with `/`, or `$base . '/'` / `$base . DIRECTORY_SEPARATOR` where `$base` resolves to a builtin `realpath()` call on a path without request data and is proven not `false` at the test. An unchecked base may be `false`, which turns the prefix into `/` and accepts every absolute path. A prefix without the trailing separator also matches sibling directories.
  - Nothing in the scope can rebind or alter the variable (or the `$base` variable) behind the analysis: no closure capturing it by reference, no `catch` binding or `static` declaration of the same name, no reference alias, no `extract()` call, no `global` declaration of it and no `$GLOBALS` use; and, located after the `realpath()` assignment, no other assignment to it (compound, destructuring slot, offset write, variable variable), no `++`/`--`, no `foreach` key or value variable and no argument bound to a by-reference parameter of a resolved callee. A write that ends before the `realpath()` assignment is superseded by it. The check is positional, so a later reuse of the variable name keeps the report. A value that code holding a reference may have replaced (a closure capturing the variable by reference that is then called, a called function of the same file writing it through `global` or `$GLOBALS`) keeps the request data it held and gains the request data that code writes to it, so the report stays. After a `catch` clause bound to the variable it holds the exception and is no longer request-derived. Passing the variable to a function the analysis cannot resolve neither validates it nor removes its request data: a project helper such as `inside($p, $base)` is not a containment proof, whether or not its declaration is visible.
  - An explicit `=== false` rejection of the canonical value is not required: each accepted test already fails for `false`.
- E1b. A value produced by a `match` expression whose arm results all carry no request data is a finite server-chosen name, whatever the subject is. One request-derived arm result keeps the report, and so does a variable holding the selection once `extract()` of request data or a reference holder writing request data may have replaced it. The shared flow analysis provides this: a `match` carries the request data of its arm results only.
- E1c. A concatenation is excluded when every operand is clean or excluded by E1a/E1b.
- E2. Unresolved targets, incomplete required facts, suppressed findings and PHP versions lacking the referenced API or language feature produce no report. User declarations shadowing builtin names are not builtin contracts.
- E3. A value the shared flow analysis proves is not the failure value at the use is never reported, whatever form the proof takes: a test in an enclosing branch, loop condition or short-circuit operand, an early exit, a statement that leaves on the failing outcome (`$v !== false or exit;`, `$v or throw …;`, `$v ?: throw …;`), a failing branch that ends in a call that never returns (a callee declared `never`, or a function or a final or private method every path of which throws or exits), a test stored in a boolean and tested later while neither the boolean nor the tested variable has been written, and, on PHP 8.0 and later outside any `try` with a `catch`, a builtin call that returned after receiving the value for a parameter that only accepts arrays, objects or callables: such a call throws a `TypeError` for `false`. A `switch (true)` case is read like the condition of an `if`/`elseif` chain; in a `switch` on the variable itself with constant labels, the cases after a `false` or `null` label and the `default` hold a value that is neither, as does a `true` case. Positions PHP may skip prove nothing: the right side of `??=`, the arguments of a null-safe call and anything chained after `?->`. A null-safe call to a never-returning method does not end the path, and `f(...)` creates a closure without calling `f`. An arm of `match (true)` is read like a `switch (true)` case; an arm that throws, exits or calls a function that never returns takes no part in what holds after the `match`. `|` and `&` between two tests are read like `||` and `&&` for the outcome that fixes both operands (a false `|`, a true `&`), and a strict `in_array($v, [...], true)` over literals that include neither `false` nor `null` excludes both where it holds. `goto` jumps are followed, including a jump out of a `try` or `catch` block, which runs the `finally` blocks it leaves on its way. Where a `goto` label is written before the use, and after the place where its scope first names the checked variable, a variable that can hold the same value, or code that may write any variable (`extract`, `compact`, `get_defined_vars`, `parse_str` given one argument, a variable variable, `$GLOBALS`, `eval`, an included file), a jump can land between a guard and the use: only what the flow analysis establishes counts there, and a guard recognised by where it is written proves nothing. Two variables can hold the same value when one is written from the other by a plain assignment, an assignment by reference or `??=`, when a `foreach` takes one out of the other, or through a chain of those; in the value assigned, the receiver and the arguments of a call, the operands of an operator other than `??`, the variables a string embeds and an element key do not count, and a statement that only reads (`echo $prefix . $parts['host'];`) ties nothing. A parameter is named from the start of the function, and a checked expression that reads `$this` or no variable counts every label before it. In a scope the flow analysis does not follow (one that evaluates or includes code, which may write any variable) a guarded use after such a label is therefore reported. A generator is followed: while it is suspended at a `yield`, the code that consumes it may write only what an unknown call may write, and, when the generator returns by reference, the variable it yields. A label written before all of that, such as a `retry:` label ahead of the `try` that produces and checks the value, changes nothing, whatever else the function reads beside the checked value; at file level every label written before the use counts. What was established of a `static` local holds until the next call, or code that runs without a call being written (an object converted to text, a magic property or offset access, the destructor of an object a variable held when it is given another value), that may enter the function again and assign the variable. A call of PHP's own counts as such when it may run code of the project: it is given a callback other than the name of a PHP function that only computes on text and numbers and raises no warning whatever text it is given (a closed list: `strlen`, `trim`, `intval` and the like, not `hexdec`, `abs` or `chr`, which warn about text that is no number), an object it may convert, a class name it loads (`method_exists($name, 'm')`, `is_callable('A::b')`, `defined('A::B')`, `class_alias($name, 'B')`, `new ReflectionClass($name)`, `$arrayObject->setIteratorClass($name)`), a class to construct (`$statement->fetchObject('Row')`) or a filter that one of PHP's own `FILTER_…` constants does not name, which may be the one that calls back, or it runs what was registered earlier (an error, session or output-buffer handler, the callbacks of a cURL transfer or an XML parser, an iterator wrapped in another). A warning that a PHP function raises by itself when called directly (`hex2bin` given an odd number of digits, a file that cannot be read) is not counted, although the error handler of the project receives it. This limits every guard form of this specification, not only the ones the flow analysis reads.
- E4. Request data that a `switch` compared with constant labels is the matched label inside an arm reached only through string labels that do not read as numbers (not empty and not starting, after whitespace, with a digit, a sign or a dot), and after the switch when every other way out replaced it or left the scope: `switch ($sort) { case 'name': case 'date': break; default: $sort = 'name'; }` validates `$sort`. A numeric, empty, boolean or null label, a label expression, a fall-through from another arm, a `default` arm and any other subject prove nothing. The subject is a plain variable or an element of a request superglobal read under one constant key (`switch ($_GET['action'])`, not `$_FILES`); the element is the matched label only while nothing that may run code is evaluated: a call, an object creation, an assignment, an `unset`, or the use of a local variable, property, offset of a local or class constant ends it, as does a reference to the superglobal. A `switch (true)` reads each case expression like the condition of an `if`/`elseif` chain.

## Report

- Range: the complete violating call or expression identified by D1; for a declaration-state violation, the offending property access or assignment. Highlight the final unsafe use for source-to-use findings.
- Severity: error.
- Message: `Validate filesystem containment before using this path.`

## Fix

No automatic fix: canonicalization, existence and symlink races require deliberate caller policy.

A fix is offered only when the concrete prerequisites above are proven. Edits use exact source byte ranges, preserve comments and argument evaluation order, and introduce syntax supported by the configured PHP version. Never invent error-handling policy.

## Options

None. Enable or disable this rule through normal rule configuration.

## PHP versions

Apply only on PHP versions supporting the referenced APIs and syntax (within PHP 5.3–8.5). Builtin behavior and argument contracts follow the configured PHP version. Extension-specific calls require resolved extension API symbols; do not treat same-named application functions as extension calls.

## Examples

```php
<?php
<error descr="Validate filesystem containment before using this path.">readfile('/srv/downloads/' . $_GET['file'])</error>;
```

Canonical containment and finite selection are not reported:

```php
<?php
function download(): void
{
    $path = realpath('/srv/downloads/' . $_GET['file']);
    if ($path === false || !str_starts_with($path, '/srv/downloads/')) {
        exit;
    }
    readfile($path);
}

function page(): void
{
    $page = match ($_GET['page']) { 'terms' => 'terms.html', default => 'home.html' };
    readfile('/srv/pages/' . $page);
}
```

A prefix without its trailing separator still admits a sibling directory:

```php
<?php
function download(): void
{
    $path = realpath('/srv/downloads/' . $_GET['file']);
    if ($path === false || !str_starts_with($path, '/srv/downloads')) {
        exit;
    }
    <error descr="Validate filesystem containment before using this path.">readfile($path)</error>;
}
```

## Divergences

Native custos rule. No upstream inspection or fixture comparison applies. These conditions are independently specified and examples are original.
