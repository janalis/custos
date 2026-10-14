---
id: HashFileFailureUsedAsDigest
group: Security
kind: semantic
needs: [names, types, flow]
php: { min: "", max: "" }
---

# HashFileFailureUsedAsDigest

## Summary

Identify unchecked hash_file result consumed as file contents or as an expected digest in strict comparison; exclude result propagation and guarded false rejection.

## Detection

- D1. Report unchecked hash_file result consumed as file contents or as an expected digest in strict comparison; exclude result propagation and guarded false rejection.
- D2. Resolve builtin functions, classes, constants and methods, including imports and supported named arguments. Follow only bounded local facts within one lexical scope. Receiver and argument provenance must establish the specific API contract; variable names and comments never prove intent.
- D3. Report one finding per violating operation. A reachable path, unchanged value identity and applicable API behavior are required. Intervening mutation, unknown calls involving tracked values, escaped aliases or exhausted analysis budgets invalidate proof.
- D4. A use is guarded when false cannot reach it for the same reaching value (no reassignment or invalidation between guard and use). Recognised guards:
  - an earlier statement of any enclosing block that is an `if` without `elseif`/`else` whose body always leaves the path (return, throw, exit, break, continue) and whose condition holds whenever the value is false; unrelated statements may sit between the guard and the use;
  - an enclosing `if` body or ternary arm selected by a condition that excludes false, or the `else` body/arm when the condition holds for false;
  - the right operand of `&&`/`and` whose left operand excludes false, or of `||`/`or` whose left operand holds for false.
- D5. A condition excludes false when it is `$v !== false`, `$v != false`, plain truthiness of `$v`, or `!empty($v)`; it holds for false when it is `$v === false`, `$v == false`, `!$v` or `empty($v)` (operands in either order, parentheses ignored). `!` swaps the two. A conjunction excludes false when either operand does and holds for false only when both do; a disjunction holds for false when either operand does and excludes false only when both do. Truthiness counts because the digest sink never needs an empty string distinguished from failure to be safe: every falsy value is rejected along with false.

## Exceptions (no report)

- E1. Exclude unresolved required symbols, user-defined lookalikes, incomplete syntax, unknown required values and unsupported versions. Cases outside D1 receive no finding.
- E2. Both `@custos-ignore HashFileFailureUsedAsDigest` and `@noinspection HashFileFailureUsedAsDigest` suppress this native rule.
- E3. A value the shared flow analysis proves is not the failure value at the use is never reported, whatever form the proof takes: a test in an enclosing branch, loop condition or short-circuit operand, an early exit, a statement that leaves on the failing outcome (`$v !== false or exit;`, `$v or throw …;`, `$v ?: throw …;`), a failing branch that ends in a call that never returns (a callee declared `never`, or a function or a final or private method every path of which throws or exits), a test stored in a boolean and tested later while neither the boolean nor the tested variable has been written, and, on PHP 8.0 and later outside any `try` with a `catch`, a builtin call that returned after receiving the value for a parameter that only accepts arrays, objects or callables: such a call throws a `TypeError` for `false`. A `switch (true)` case is read like the condition of an `if`/`elseif` chain; in a `switch` on the variable itself with constant labels, the cases after a `false` or `null` label and the `default` hold a value that is neither, as does a `true` case. Positions PHP may skip prove nothing: the right side of `??=`, the arguments of a null-safe call and anything chained after `?->`. A null-safe call to a never-returning method does not end the path, and `f(...)` creates a closure without calling `f`. An arm of `match (true)` is read like a `switch (true)` case; an arm that throws, exits or calls a function that never returns takes no part in what holds after the `match`. `|` and `&` between two tests are read like `||` and `&&` for the outcome that fixes both operands (a false `|`, a true `&`), and a strict `in_array($v, [...], true)` over literals that include neither `false` nor `null` excludes both where it holds. `goto` jumps are followed, except in a scope that also contains a `try`, which the flow analysis leaves alone. Where a `goto` label is written before the use, and after the place where its scope first names the checked variable, a variable that can hold the same value, or code that may write any variable (`extract`, `compact`, `get_defined_vars`, `parse_str` given one argument, a variable variable, `$GLOBALS`, `eval`, an included file), a jump can land between a guard and the use: only what the flow analysis establishes counts there, and a guard recognised by where it is written proves nothing. Two variables can hold the same value when one is written from the other by a plain assignment, an assignment by reference or `??=`, when a `foreach` takes one out of the other, or through a chain of those; in the value assigned, the receiver and the arguments of a call, the operands of an operator other than `??`, the variables a string embeds and an element key do not count, and a statement that only reads (`echo $prefix . $parts['host'];`) ties nothing. A parameter is named from the start of the function, and a checked expression that reads `$this` or no variable counts every label before it. In a scope the flow analysis does not follow (one that also contains a `try`, a generator, a scope that evaluates or includes code) a guarded use after such a label is therefore reported. A label written before all of that, such as a `retry:` label ahead of the `try` that produces and checks the value, changes nothing, whatever else the function reads beside the checked value; at file level every label written before the use counts. What was established of a `static` local holds until the next call, or code that runs without a call being written (an object converted to text, a magic property or offset access, the destructor of an object a variable held when it is given another value), that may enter the function again and assign the variable. A call of PHP's own counts as such when it may run code of the project: it is given a callback other than the name of a PHP function that only computes on text and numbers and raises no warning whatever text it is given (a closed list: `strlen`, `trim`, `intval` and the like, not `hexdec`, `abs` or `chr`, which warn about text that is no number), an object it may convert, a class name it loads (`method_exists($name, 'm')`, `is_callable('A::b')`, `defined('A::B')`, `class_alias($name, 'B')`, `new ReflectionClass($name)`, `$arrayObject->setIteratorClass($name)`), a class to construct (`$statement->fetchObject('Row')`) or a filter that one of PHP's own `FILTER_…` constants does not name, which may be the one that calls back, or it runs what was registered earlier (an error, session or output-buffer handler, the callbacks of a cURL transfer or an XML parser, an iterator wrapped in another). A warning that a PHP function raises by itself when called directly (`hex2bin` given an odd number of digits, a file that cannot be read) is not counted, although the error handler of the project receives it. This limits every guard form of this specification, not only the ones the flow analysis reads.

## Report

- Range: the complete violating call or expression. Follow D1 if it specifies a narrower range.
- Severity: warning.
- Enabled by default: true.
- Message: `Reject file hashing failure before using the digest.`

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
file_put_contents($manifest, hash_file('sha256', $path));
```

Valid case:

```php
<?php
$h = hash_file('sha256', $path); if ($h !== false) { file_put_contents($manifest, $h); }
```

Valid cases (early exit followed by unrelated work, truthiness, short-circuit):

```php
<?php
function store(string $path, string $manifest): bool
{
    $sum = hash_file('sha256', $path);
    if (!$sum || $manifest === '') {
        return false;
    }
    $label = basename($path);
    return file_put_contents($manifest, $sum) !== false;
}

function matches(string $path): bool
{
    $sum = hash_file('sha256', $path);
    return $sum !== false && $sum === 'abc';
}
```

## Divergences

Native custos inspection, independently specified; no EA counterpart or EA conformance requirement. Conservative proof intentionally excludes unknown runtime state and intent.
