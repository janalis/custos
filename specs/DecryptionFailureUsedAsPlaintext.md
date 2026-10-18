---
id: DecryptionFailureUsedAsPlaintext
group: Security
kind: semantic
needs: [names, types, flow]
php: { min: "", max: "" }
---

# DecryptionFailureUsedAsPlaintext

## Summary

Decryption can fail and return false. Check success before treating the result as plaintext.

## Detection

- D1. Resolved openssl_decrypt result is passed to a known string sink file_put_contents data argument or strlen without a false-result guard. Track direct nesting or unmodified locals; arbitrary named consumers are excluded.
- D2. Resolve builtin functions, classes, constants and methods; user-defined lookalikes are excluded. Follow only bounded local proof in one lexical scope.
- D3. A false-result guard is a condition on the same local, still holding the same decryption call, whose outcome on the path to the sink rules out false. Accepted tests, also negated with `!`, parenthesised, or combined (any operand of a true `&&`/`and`, any operand of a false `||`/`or`): strict or loose comparison with `false` on either side, plain truthiness, `empty()`, a resolved builtin `is_string()`, and the assignment of the call written inside the condition. Accepted positions: the matching branch of an enclosing `if`/`elseif`/`else` (later arms count every earlier condition as failed), the matching arm of a ternary, the right operand of a short-circuit operator, and code following an earlier `if` without alternatives whose body always exits, in the same or an enclosing statement list.
- D4. Emit one finding per violating operation, not per supporting evidence statement. Use existing analysis budgets; exhausted proof produces no report.

## Exceptions (no report)

- E1. Do not report unresolved or shadowed builtin names, unknown values, or incomplete receiver/type provenance. Unknown calls, aliases escaping, and intervening writes invalidate local proof. Similar code outside the stated detection contract is not a finding.
- E2. A test of another variable, of an earlier or later value of the local, a `null` comparison, `isset()`, or a guard whose failing branch falls through to the sink is not a guard; the sink is still reported.
- E3. Both standard suppression forms suppress this native ID.
- E4. A value the shared flow analysis proves is not the failure value at the use is never reported, whatever form the proof takes: a test in an enclosing branch, loop condition or short-circuit operand, an early exit, a statement that leaves on the failing outcome (`$v !== false or exit;`, `$v or throw …;`, `$v ?: throw …;`), a failing branch that ends in a call that never returns (a callee declared `never`, or a function or a final or private method every path of which throws or exits), a test stored in a boolean and tested later while neither the boolean nor the tested variable has been written, and, on PHP 8.0 and later outside any `try` with a `catch`, a builtin call that returned after receiving the value for a parameter that only accepts arrays, objects or callables: such a call throws a `TypeError` for `false`. A `switch (true)` case is read like the condition of an `if`/`elseif` chain; in a `switch` on the variable itself with constant labels, the cases after a `false` or `null` label and the `default` hold a value that is neither, as does a `true` case. Positions PHP may skip prove nothing: the right side of `??=`, the arguments of a null-safe call and anything chained after `?->`. A null-safe call to a never-returning method does not end the path, and `f(...)` creates a closure without calling `f`. An arm of `match (true)` is read like a `switch (true)` case; an arm that throws, exits or calls a function that never returns takes no part in what holds after the `match`. `|` and `&` between two tests are read like `||` and `&&` for the outcome that fixes both operands (a false `|`, a true `&`), and a strict `in_array($v, [...], true)` over literals that include neither `false` nor `null` excludes both where it holds. `goto` jumps are followed, including a jump out of a `try` or `catch` block, which runs the `finally` blocks it leaves on its way. Where a `goto` label is written before the use, and after the place where its scope first names the checked variable, a variable that can hold the same value, or code that may write any variable (`extract`, `compact`, `get_defined_vars`, `parse_str` given one argument, a variable variable, `$GLOBALS`, `eval`, an included file), a jump can land between a guard and the use: only what the flow analysis establishes counts there, and a guard recognised by where it is written proves nothing. Two variables can hold the same value when one is written from the other by a plain assignment, an assignment by reference or `??=`, when a `foreach` takes one out of the other, or through a chain of those; in the value assigned, the receiver and the arguments of a call, the operands of an operator other than `??`, the variables a string embeds and an element key do not count, and a statement that only reads (`echo $prefix . $parts['host'];`) ties nothing. A parameter is named from the start of the function, and a checked expression that reads `$this` or no variable counts every label before it. In a scope the flow analysis does not follow (one that evaluates or includes code, which may write any variable) a guarded use after such a label is therefore reported. A generator is followed: while it is suspended at a `yield`, the code that consumes it may write only what an unknown call may write, and, when the generator returns by reference, the variable it yields. A label written before all of that, such as a `retry:` label ahead of the `try` that produces and checks the value, changes nothing, whatever else the function reads beside the checked value; at file level every label written before the use counts. What was established of a `static` local holds until the next call, or code that runs without a call being written (an object converted to text, a magic property or offset access, the destructor of an object a variable held when it is given another value), that may enter the function again and assign the variable. A call of PHP's own counts as such when it may run code of the project: it is given a callback other than the name of a PHP function that only computes on text and numbers and raises no warning whatever text it is given (a closed list: `strlen`, `trim`, `intval` and the like, not `hexdec`, `abs` or `chr`, which warn about text that is no number), an object it may convert, a class name it loads (`method_exists($name, 'm')`, `is_callable('A::b')`, `defined('A::B')`, `class_alias($name, 'B')`, `new ReflectionClass($name)`, `$arrayObject->setIteratorClass($name)`), a class to construct (`$statement->fetchObject('Row')`) or a filter that one of PHP's own `FILTER_…` constants does not name, which may be the one that calls back, or it runs what was registered earlier (an error, session or output-buffer handler, the callbacks of a cURL transfer or an XML parser, an iterator wrapped in another). A warning that a PHP function raises by itself when called directly (`hex2bin` given an odd number of digits, a file that cannot be read) is not counted, although the error handler of the project receives it. This limits every guard form of this specification, not only the ones the flow analysis reads.

## Report

- Range: the violating expression or call, unless D1 specifies a narrower range.
- Severity: warning.
- Enabled by default: true.
- Message: Check decryption success before using the plaintext.

## Fix

None. Choosing a repair requires intent or additional runtime information.

## Options

None.

## PHP versions

Available throughout supported PHP 5.3–8.5 where the referenced builtin/extension is available. Respect syntax/version-specific receiver contracts; unavailable APIs are not this rule’s concern.

## Examples

```php
<?php
$plain=openssl_decrypt($cipher,"aes-256-gcm",$key,OPENSSL_RAW_DATA,$iv,$tag); file_put_contents($path,<warning descr="Check decryption success before using the plaintext.">$plain</warning>);
```

Valid case:

```php
<?php
$p=openssl_decrypt($c,"aes-256-gcm",$key,OPENSSL_RAW_DATA,$iv,$tag);if($p!==false){file_put_contents($path,$p);}
```

Guards written in other forms are also accepted:

```php
<?php
if (($text = openssl_decrypt($blob, "aes-256-cbc", $secret, 0, $nonce)) === false) { throw new RuntimeException("bad blob"); }
echo strlen($text);
$size = !$text ? 0 : strlen($text);
$long = is_string($text) && strlen($text) > 16;
```

## Divergences

Native custos inspection, independently specified; no EA counterpart or EA conformance requirement.
