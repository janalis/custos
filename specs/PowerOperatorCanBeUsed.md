---
id: PowerOperatorCanBeUsed
group: Language level migration
kind: semantic
needs: [names, stubs]
php: { min: "5.6", max: "" }
---

# PowerOperatorCanBeUsed

## Summary
Since PHP 5.6 exponentiation has its own operator. `pow($a, $b)` can be
written `$a ** $b`, which is shorter and avoids a function call.

## Detection
- **D1** A function call whose name (last segment, case-insensitive: `POW`
  matches) is `pow`
  and which resolves to the global function `\pow` under PHP's runtime
  rules: `\pow(...)` and unqualified `pow(...)` in the global namespace match;
  an unqualified call inside a namespace matches only when that namespace
  declares no `pow` function and no `use function` import brings in a
  non-global one; a qualified `Foo\pow(...)` never matches. Method calls and
  static calls named `pow` are not function calls.
- **D2** The call has exactly two arguments, base `B` and exponent `E`.
- **D3** Build the replacement:
  - `b` = `(` + B text + `)` when `B` is a binary expression (any binary
    operator, including `??`, `.`, comparisons, logical operators,
    `instanceof`) or a ternary (full or short `?:`); otherwise B's text.
  - `e` = same rule applied to `E`.
  - `R0` = `b ** e` (single spaces around `**`).
  - If the call's direct parent is a binary expression (any operator), `R` =
    `(` + `R0` + `)`; otherwise `R` = `R0`.
- **D4** Report the call.

## Exceptions (no report)
- **E1** Language level below 5.6.
- **E2** Argument count other than two (including a single spread argument
  `pow(...$pair)`).

## Report
- Range: the whole call (`pow` through its closing parenthesis).
- Severity: warning.
- Message: `Use '{R}' (exponentiation operator) instead.`

## Fix
- **F1** Replace the call with `R` verbatim. Argument texts are copied as
  written (comments/whitespace inside an argument are kept).
  - `pow($x, 3)` → `$x ** 3`
  - `pow($x - 1, $n)` → `($x - 1) ** $n`
  - `pow($x, $n ?: 2)` → `$x ** ($n ?: 2)`
  - `10 * pow($x, 2)` → `10 * ($x ** 2)`

## Options
None.

## PHP versions
Reported only at language level ≥ 5.6. The upstream fixture declares no level
and runs at the test default (below 7.1 but at least 5.6); the conformance
harness default level must therefore be ≥ 5.6 for this rule.

## Examples

```php
<?php
$area   = <warning descr="Use '$side ** 2' (exponentiation operator) instead.">pow($side, 2)</warning>;
$growth = <warning descr="Use '($rate * 0.01 + 1) ** $years' (exponentiation operator) instead.">pow($rate * 0.01 + 1, $years)</warning>;
$scaled = <warning descr="Use '2 ** ($bits - 1)' (exponentiation operator) instead.">\pow(2, $bits - 1)</warning>;
$total  = $offset - <warning descr="Use '($k ** $m)' (exponentiation operator) instead.">pow($k, $m)</warning>;
$safe   = <warning descr="Use '($q ?? 0) ** ($flag ? 2 : 3)' (exponentiation operator) instead.">pow($q ?? 0, $flag ? 2 : 3)</warning>;
$keep   = pow($side);
$meth   = $calc->pow($side, 2);
```

```php
<?php
$area   = $side ** 2;
$growth = ($rate * 0.01 + 1) ** $years;
$scaled = 2 ** ($bits - 1);
$total  = $offset - ($k ** $m);
$safe   = ($q ?? 0) ** ($flag ? 2 : 3);
$keep   = pow($side);
$meth   = $calc->pow($side, 2);
```

## Divergences
- Upstream does not wrap unary-operator or assignment arguments:
  `pow(-2, 2)` becomes `-2 ** 2`, which PHP evaluates as `-(2 ** 2)` = -4
  instead of 4; `pow($a = 2, 3)` becomes `$a = 2 ** 3`. Recommendation:
  also parenthesise a base that is a unary expression (negation, cast, `!`,
  `@`, `++`/`--` prefix) or an assignment, and an exponent that is an
  assignment. Not covered by upstream fixtures, so conformance is unaffected.
- Upstream builds the text with placeholder substitution; an argument whose
  text happens to contain the placeholder token could be garbled. Our
  implementation must simply concatenate.
- **Callee resolved — custos diverges from upstream** (D1). Upstream checks
  only the last name segment, so a user function `App\pow()` (qualified, or
  unqualified inside `namespace App` where it is declared) is reported and
  replaced by `**`, which silently drops the user's implementation. custos
  reports only calls that reach the built-in `pow`.
- **Function-name case (custos diverges):** upstream matches `pow`
  case-sensitively, missing `POW($a, 2)` and `\Pow($a, 2)`. custos matches
  any case (D1).
