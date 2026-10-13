---
id: ImageDecodeFailureUnchecked
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "", max: "" }
---

# ImageDecodeFailureUnchecked

## Summary

Decoding image bytes can fail and return false. Check the result before using it as an image.

## Detection

- D1. A local directly assigned resolved imagecreatefromstring is consumed as first argument of imagepng/imagejpeg/imagegif/imagesx/imagesy without a false-result guard. Require unknown input or known invalid literal bytes.
- D1a. A guard is a test of the same variable while it still holds that decode result. Accepted tests: plain truthiness (`$im`, `!$im`, `empty($im)`), strict or loose comparison with `false`, `instanceof`, `is_object`, `is_resource`, `is_bool`, and the decode assignment tested in place (`if (!$im = imagecreatefromstring($b))`). Tests combine through `!`, `&&`/`and` and `||`/`or` with their usual logic: a conjunction that holds proves each operand, a disjunction that fails refutes each operand.
- D1b. A guard protects a use when the use sits in the branch the test selects (if/elseif/else arm, including arms reached only after an earlier condition of the chain failed; ternary arm; right operand of a short-circuit operator), or when an earlier statement of the same or an enclosing statement list leaves the path on failure: an `if` without `elseif` whose failing arm always terminates (return, throw, exit, break, continue), `test or exit/throw`, `!test and exit/throw`, or `assert(test)`.
- D2. Resolve builtin functions, classes, constants and methods; user-defined lookalikes are excluded. Follow only bounded local proof in one lexical scope.
- D3. Emit one finding per violating operation, not per supporting evidence statement. Use existing analysis budgets; exhausted proof produces no report.

## Exceptions (no report)

- E1. Do not report unresolved or shadowed builtin names, unknown values, or incomplete receiver/type provenance. Unknown calls, aliases escaping, and intervening writes invalidate local proof. Similar code outside the stated detection contract is not a finding.
- E2. Both standard suppression forms suppress this native ID.
- E3. A guard on an earlier decode result does not cover a later reassignment, and a guard whose failing arm merely logs or calls an ordinary function does not protect the following code: both are still reported.
- E4. Guards expressed indirectly stay unrecognised and are reported: `match (true)` arms, and exceptions thrown and caught inside the same function.
- E5. A value the shared flow analysis proves is not the failure value at the use is never reported, whatever form the proof takes: a test in an enclosing branch, loop condition or short-circuit operand, an early exit, a statement that leaves on the failing outcome (`$v !== false or exit;`, `$v or throw …;`, `$v ?: throw …;`), a failing branch that ends in a call that never returns (a callee declared `never`, or a function or a final or private method every path of which throws or exits), a test stored in a boolean and tested later while neither the boolean nor the tested variable has been written, and, on PHP 8.0 and later outside any `try` with a `catch`, a builtin call that returned after receiving the value for a parameter that only accepts arrays, objects or callables: such a call throws a `TypeError` for `false`. A `switch (true)` case is read like the condition of an `if`/`elseif` chain; in a `switch` on the variable itself with constant labels, the cases after a `false` or `null` label and the `default` hold a value that is neither, as does a `true` case. Positions PHP may skip prove nothing: the right side of `??=`, the arguments of a null-safe call and anything chained after `?->`. A null-safe call to a never-returning method does not end the path, and `f(...)` creates a closure without calling `f`. An arm of `match (true)` is read like a `switch (true)` case; an arm that throws, exits or calls a function that never returns takes no part in what holds after the `match`. `|` and `&` between two tests are read like `||` and `&&` for the outcome that fixes both operands (a false `|`, a true `&`), and a strict `in_array($v, [...], true)` over literals that include neither `false` nor `null` excludes both where it holds. `goto` jumps are followed, except in a scope that also contains a `try`, which the flow analysis leaves alone. Where a `goto` label is written before the use, and after the place where its scope first names the checked variable, a variable that can hold the same value, or code that may write any variable (`extract`, `compact`, a variable variable, `eval`, an included file), a jump can land between a guard and the use: only what the flow analysis establishes counts there, and a guard recognised by where it is written proves nothing. Two variables can hold the same value when one is written from the other by a plain assignment, an assignment by reference or `??=`, when a `foreach` takes one out of the other, or through a chain of those; in the value assigned, the receiver and the arguments of a call, the operands of an operator other than `??`, the variables a string embeds and an element key do not count, and a statement that only reads (`echo $prefix . $parts['host'];`) ties nothing. A parameter is named from the start of the function, and a checked expression that reads `$this` or no variable counts every label before it. In a scope the flow analysis does not follow (one that also contains a `try`, a generator, a scope that evaluates or includes code) a guarded use after such a label is therefore reported. A label written before all of that, such as a `retry:` label ahead of the `try` that produces and checks the value, changes nothing, whatever else the function reads beside the checked value; at file level every label written before the use counts. What was established of a `static` local holds until the next call, or code that runs without a call being written (an object converted to text, a magic property or offset access, the destructor of an object a variable held when it is given another value), that may enter the function again and assign the variable. A call of PHP's own counts as such when it may run code of the project: it is given a callback other than the name of a PHP function that only computes on text and numbers (`strlen`, `trim`, `intval` and the like), an object it may convert, or a class name it loads (`method_exists($name, 'm')`, `is_callable('A::b')`, `defined('A::B')`, `new ReflectionClass($name)`), a class to construct (`$statement->fetchObject('Row')`), or it runs what was registered earlier (an error, session or output-buffer handler, the callbacks of a cURL transfer or an XML parser, an iterator wrapped in another). This limits every guard form of this specification, not only the ones the flow analysis reads.

## Report

- Range: the violating expression or call, unless D1 specifies a narrower range.
- Severity: warning.
- Enabled by default: true.
- Message: Check image decoding success before using the image.

## Fix

None. Choosing a repair requires intent or additional runtime information.

## Options

None.

## PHP versions

Available throughout supported PHP 5.3–8.5 where the referenced builtin/extension is available. Respect syntax/version-specific receiver contracts; unavailable APIs are not this rule’s concern.

## Examples

```php
<?php
$im=imagecreatefromstring($bytes); imagepng(<warning descr="Check image decoding success before using the image.">$im</warning>);
```

Valid case:

```php
<?php
$im=imagecreatefromstring($bytes); if($im===false){throw new RuntimeException("Invalid image");} imagepng($im);
```

```php
<?php
$im = imagecreatefromstring($bytes);
if (!$im) { return; }
echo $im instanceof GdImage ? imagesx($im) : 0;
```

## Divergences

Native custos inspection, independently specified; no EA counterpart or EA conformance requirement.
