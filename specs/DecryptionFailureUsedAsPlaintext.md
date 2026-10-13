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
- E4. A value the shared flow analysis proves is not the failure value at the use is never reported, whatever form the proof takes: a test in an enclosing branch, loop condition or short-circuit operand, an early exit, a statement that leaves on the failing outcome (`$v !== false or exit;`, `$v or throw …;`, `$v ?: throw …;`), a failing branch that ends in a call that never returns (a callee declared `never`, or a function or a final or private method every path of which throws or exits), a test stored in a boolean and tested later while neither the boolean nor the tested variable has been written, and, on PHP 8.0 and later outside any `try` with a `catch`, a builtin call that returned after receiving the value for a parameter that only accepts arrays, objects or callables: such a call throws a `TypeError` for `false`. A `switch (true)` case is read like the condition of an `if`/`elseif` chain; in a `switch` on the variable itself with constant labels, the cases after a `false` or `null` label and the `default` hold a value that is neither, as does a `true` case. Positions PHP may skip prove nothing: the right side of `??=`, the arguments of a null-safe call and anything chained after `?->`. A null-safe call to a never-returning method does not end the path, and `f(...)` creates a closure without calling `f`.

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
