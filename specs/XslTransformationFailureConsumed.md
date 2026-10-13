---
id: XslTransformationFailureConsumed
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "", max: "" }
---

# XslTransformationFailureConsumed

## Summary

Identify transformToXML result on a resolved XSLTProcessor passed directly or unchanged to file_put_contents or echo without rejecting false.

## Detection

- D1. Report transformToXML result on a resolved XSLTProcessor passed directly or unchanged to file_put_contents or echo without rejecting false.
- D2. Resolve builtin functions, classes, constants and methods, including imports and supported named arguments. Follow only bounded local facts within one lexical scope. Receiver and argument provenance must establish the specific API contract; variable names and comments never prove intent.
- D4. A value counts as having rejected false when the use is dominated by a test of the same unchanged value that cannot pass for false:
  - a strict comparison with `false` (`!== false` branch, or a terminating `=== false` guard);
  - a truthiness test (`if ($out)`, `if (!$out) { return; }`, `empty($out)`), since false and null are both falsy;
  - a loose comparison with `false` (`!= false` branch, terminating `== false` guard);
  - a resolved builtin `is_string()` test.
  The test may be the condition of an enclosing `if`/`elseif`/`else` branch or ternary arm, the left operand of a short-circuit `&&`/`||`/`and`/`or` whose right operand holds the use, or an earlier sibling `if` without `else`/`elseif` whose body always terminates. Negation and parentheses are followed. A conjunction that held, or a disjunction that failed, proves what either operand proves; a conjunction that failed, or a disjunction that held, proves only what both operands prove (`if ($out || is_string($out))`, `if ($out === false && !is_string($out)) { return; }`). An assignment of the call inside the condition is the tested value.
- D3. Report one finding per violating operation. A reachable path, unchanged value identity and applicable API behavior are required. Intervening mutation, unknown calls involving tracked values, escaped aliases or exhausted analysis budgets invalidate proof.

## Exceptions (no report)

- E1. Exclude unresolved required symbols, user-defined lookalikes, incomplete syntax, unknown required values and unsupported versions. Cases outside D1 receive no finding.
- E3. A guard on a different value, on an earlier transformation result that was since reassigned, or on a test that false can pass (`isset`, `!== null`, `is_string($out) || $other`) does not suppress the finding.
- E4. A value the shared flow analysis proves is not false at the use is never reported, whatever form the proof takes: a truthiness or `empty()` test, a loose comparison with `false` or `null`, a type predicate, or an early exit whose `if` also has an `else` branch that continues. In particular `if (!$out) { return; } else { ... }` guards the statements after it. The same holds for a statement that leaves on the failing outcome (`$v !== false or exit;`, `$v or throw …;`, `$v ?: throw …;`), for a failing branch that ends in a call that never returns (a callee declared `never`, or a function or a final or private method every path of which throws or exits), for a test stored in a boolean and tested later while neither the boolean nor the tested variable has been written, and, on PHP 8.0 and later outside any `try` with a `catch`, for a builtin call that returned after receiving the value for a parameter that only accepts arrays, objects or callables: such a call throws a `TypeError` for `false`. A `switch (true)` case is read like the condition of an `if`/`elseif` chain; in a `switch` on the variable itself with constant labels, the cases after a `false` or `null` label and the `default` hold a value that is neither, as does a `true` case. Positions PHP may skip prove nothing: the right side of `??=`, the arguments of a null-safe call and anything chained after `?->`. A null-safe call to a never-returning method does not end the path, and `f(...)` creates a closure without calling `f`. An arm of `match (true)` is read like a `switch (true)` case; an arm that throws, exits or calls a function that never returns takes no part in what holds after the `match`. `|` and `&` between two tests are read like `||` and `&&` for the outcome that fixes both operands (a false `|`, a true `&`), and a strict `in_array($v, [...], true)` over literals that include neither `false` nor `null` excludes both where it holds. `goto` jumps are followed, except in a scope that also contains a `try`, which the flow analysis leaves alone. Where a `goto` label is written before the use, and after the place where its scope first names the checked variable, a variable that can hold the same value, or code that may write any variable (`extract`, `compact`, a variable variable, `eval`, an included file), a jump can land between a guard and the use: only what the flow analysis establishes counts there, and a guard recognised by where it is written proves nothing. Two variables can hold the same value when one is written from the other by a plain assignment, an assignment by reference or `??=`, when a `foreach` takes one out of the other, or through a chain of those; in the value assigned, the receiver and the arguments of a call, the operands of an operator other than `??`, the variables a string embeds and an element key do not count, and a statement that only reads (`echo $prefix . $parts['host'];`) ties nothing. A parameter is named from the start of the function, and a checked expression that reads `$this` or no variable counts every label before it. In a scope the flow analysis does not follow (one that also contains a `try`, a generator, a scope that evaluates or includes code) a guarded use after such a label is therefore reported. A label written before all of that, such as a `retry:` label ahead of the `try` that produces and checks the value, changes nothing, whatever else the function reads beside the checked value; at file level every label written before the use counts. What was established of a `static` local holds until the next call, or code that runs without a call being written (an object converted to text, a magic property or offset access, the destructor of an object a variable held when it is given another value), that may enter the function again and assign the variable. A call of PHP's own counts as such when it may run code of the project: it is given a callback other than the name of a PHP function that only computes on text and numbers (`strlen`, `trim`, `intval` and the like), an object it may convert, or a class name it loads (`method_exists($name, 'm')`, `is_callable('A::b')`, `defined('A::B')`, `new ReflectionClass($name)`), a class to construct (`$statement->fetchObject('Row')`), or it runs what was registered earlier (an error, session or output-buffer handler, the callbacks of a cURL transfer or an XML parser, an iterator wrapped in another). This limits every guard form of this specification, not only the ones the flow analysis reads.
- E2. Both `@custos-ignore XslTransformationFailureConsumed` and `@noinspection XslTransformationFailureConsumed` suppress this native rule.

## Report

- Range: the complete violating call or expression. Follow D1 if it specifies a narrower range.
- Severity: warning.
- Enabled by default: true.
- Message: `Reject transformation failure before using the XML.`

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
$processor = new XSLTProcessor(); file_put_contents($output, $processor->transformToXML($document));
```

Valid case:

```php
<?php
$processor = new XSLTProcessor(); $out = $processor->transformToXML($document); if ($out !== false) { file_put_contents($output, $out); }
```

```php
<?php
$processor = new XSLTProcessor(); $out = $processor->transformToXML($document); if (!$out) { throw new RuntimeException('transformation failed'); } echo $out;
```

## Divergences

Native custos inspection, independently specified; no EA counterpart or EA conformance requirement. Conservative proof intentionally excludes unknown runtime state and intent.
