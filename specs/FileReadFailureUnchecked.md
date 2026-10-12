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

- D1. Report file_get_contents or stream_get_contents result consumed as a string without excluding false.
- D2. Resolve builtin calls and named arguments before checking the contract. Project-local callees are considered only when their bodies or immutable summaries establish the required facts. The default is enabled, with positive proof required.
- D3. For flow-dependent conditions, require a reachable source-to-use path with no intervening invalidation. Unknown calls, ambiguous aliases, recursion or analysis-budget exhaustion cannot establish definite facts.

## Exceptions (no report)

- E1. Strict false guard, deliberate documented false coercion, termination, and result reassignment invalidate the report.
- E2. Unresolved targets, incomplete required facts, suppressed findings and PHP versions lacking the referenced API or language feature produce no report. User declarations shadowing builtin names are not builtin contracts.
- E3. A guard on the same read result that excludes false counts as handling, whether it selects the branch holding the use (`if`, `elseif`/`else`, ternary, the right operand of `&&`/`||`) or sits in an earlier sibling `if` without alternatives whose body leaves the path. Accepted conditions, possibly negated or combined with `&&`/`||` where the combination still proves the fact: truthiness, `empty()`, `is_string()`, strict comparison with `false`, and loose comparison with a constant that `false` equals (`false`, `null`, `0`, `''`, `'0'`). The guarded variable must hold only that earlier read, or be assigned it inside the condition.
- E4. A guard body leaves the path when it terminates syntactically or contains, as a direct statement, a call whose resolved return type is `never`. Calls with unknown or returning targets do not qualify.
- E5. `$text = read() or <alternative>` (also `||` around the parenthesized assignment) handles failure when the alternative is `exit`/`die`, a `throw` expression or a call resolved as `never`; uses inside the alternative itself stay reported.
- E6. A value the shared flow analysis proves is not the failure value at the use is never reported, whatever form the proof takes: a test in an enclosing branch, loop condition or short-circuit operand, an early exit, a statement that leaves on the failing outcome (`$v !== false or exit;`, `$v or throw …;`, `$v ?: throw …;`), a failing branch that ends in a call that never returns (a callee declared `never`, or a function or a final or private method every path of which throws or exits), a test stored in a boolean and tested later while neither the boolean nor the tested variable has been written, and, on PHP 8.0 and later outside any `try` with a `catch`, a builtin call that returned after receiving the value for a parameter that only accepts arrays, objects or callables: such a call throws a `TypeError` for `false`. Positions PHP may skip prove nothing: the right side of `??=`, the arguments of a null-safe call and anything chained after `?->`. A null-safe call to a never-returning method does not end the path, and `f(...)` creates a closure without calling `f`.

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

## Divergences

Native custos rule. No upstream inspection or fixture comparison applies. These conditions are independently specified and examples are original.
