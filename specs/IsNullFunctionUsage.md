---
id: IsNullFunctionUsage
group: Code style
kind: syntax
needs: []
php: { min: "", max: "" }
---

# IsNullFunctionUsage

## Summary
`is_null($v)` is a function call doing what the `=== null` comparison does
natively. This opt-in rule rewrites `is_null()` calls (including negations and
comparisons of the result against `true`/`false`) into an identity comparison
with `null`.

## Detection
D1. Node: a plain function call with exactly one argument that resolves to
    the global function `is_null`: the name is compared case-insensitively
    (`IS_NULL`, `Is_Null` count) and may be written `\is_null`; a qualified
    namespaced call (`App\is_null`), a non-global `use function` import, or
    an unqualified call in a namespace that declares its own `is_null()`
    does not count. Let *A* be that argument.

D2. Determine the *target* expression and the *polarity* (checks "is null"
    = positive, or "is not null" = negative) from the call's **direct** parent:
    - D2a. Parent is a logical-not unary `!is_null(A)` → target is the unary
      expression; polarity negative.
    - D2b. Parent is a binary expression and the *other* operand is a boolean
      constant `true` or `false` (case-insensitive, optionally `\`-qualified):
      - operator `==` or `===`: target is the binary expression; polarity
        positive when the other operand is `true`, negative when `false`;
      - operator `!=` or `!==` (also `<>`): target is the binary expression;
        polarity negative when the other operand is `true`, positive when
        `false`;
      - any other operator (`&&`, `||`, `and`, `xor`, …): target is the call
        itself, polarity positive.
      The call may be either the left or the right operand.
    - D2c. Any other parent (including a binary expression whose other operand
      is not a boolean constant, parentheses, argument lists, statements):
      target is the call itself, polarity positive.

D3. Every call matching D1 is reported (on its D2 target).

## Exceptions (no report)
E1. `is_null()` with zero or two-plus arguments.
E2. Calls that resolve to a user function named `is_null` (see D1);
    method/static calls named `is_null`.
E3. A parenthesized negation `!(is_null($v))` is not treated as negation: the
    inner call is the target with positive polarity.

## Report
- Range: the D2 target — the call (`is_null` through `)`, including any leading
  namespace qualifier), or the whole `!is_null(...)`, or the whole binary
  comparison with the boolean (from its left operand's start to its right
  operand's end).
- Severity: info (fixtures tag it `weak_warning`).
- Message: `Replace with '{replacement}'.`

## Fix
F1. Replace the target with the replacement built as:
    - `{a}` = verbatim text of *A*, wrapped in parentheses `({A})` when *A* is
      an assignment (plain or compound), a ternary (including `?:`), or any
      binary expression (including `??`, `&&`, `.`, `+`, `instanceof`, …).
      Other forms (variables, calls, casts, unary ops, literals, member
      accesses) are not wrapped.
    - `{op}` = `===` for positive polarity, `!==` for negative.
    - regular comparison style (default): `{a} {op} null`;
      yoda style: `null {op} {a}`.
    Single spaces around `{op}`; nothing else is added (no outer parentheses).

## Options
| Option | Type | Default | Effect |
|---|---|---|---|
| (none) | | | Operand order follows the global comparison style (`regular` default / `yoda`). |

## PHP versions
None.

## Examples
Regular style:

```php
<?php

$a = <weak_warning descr="Replace with '$row === null'.">is_null($row)</weak_warning>;
$b = <weak_warning descr="Replace with '$row !== null'.">!is_null($row)</weak_warning>;
$c = <weak_warning descr="Replace with '$cfg->get() === null'.">is_null($cfg->get()) == TRUE</weak_warning>;
$d = <weak_warning descr="Replace with '$cfg->get() !== null'.">false === is_null($cfg->get())</weak_warning>;
$e = <weak_warning descr="Replace with '$row === null'.">is_null($row) != false</weak_warning>;
$f = <weak_warning descr="Replace with '$row !== null'.">\is_null($row) !== true</weak_warning>;
$g = $ok and <weak_warning descr="Replace with '$row === null'.">is_null($row)</weak_warning>;
$h = <weak_warning descr="Replace with '($left ?? $right) !== null'.">!is_null($left ?? $right)</weak_warning>;
$i = <weak_warning descr="Replace with '($n = next($it)) === null'.">is_null($n = next($it))</weak_warning>;
$j = <weak_warning descr="Replace with '(int) $raw === null'.">is_null((int) $raw)</weak_warning>;
$k = is_null($p, $q);
```

```php
<?php

$a = $row === null;
$b = $row !== null;
$c = $cfg->get() === null;
$d = $cfg->get() !== null;
$e = $row === null;
$f = $row !== null;
$g = $ok and $row === null;
$h = ($left ?? $right) !== null;
$i = ($n = next($it)) === null;
$j = (int) $raw === null;
$k = is_null($p, $q);
```

Yoda style:

```php
<?php

$x = <weak_warning descr="Replace with 'null !== $token'.">true !== is_null($token)</weak_warning>;
$y = <weak_warning descr="Replace with 'null === ($m ?: $n)'.">is_null($m ?: $n)</weak_warning>;
```

```php
<?php

$x = null !== $token;
$y = null === ($m ?: $n);
```

## Divergences
- **Letter case and name resolution (custos diverges).** Upstream checks
  only the last name segment, case-sensitively: `IS_NULL($v)` is missed, while
  namespaced calls like `App\is_null($v)` (a user function) are reported and
  rewritten. custos matches calls that resolve to the global `is_null` in any
  letter case and skips namespaced/shadowing user functions (D1).
- **Parentheses (custos diverges):** upstream adds no outer parentheses, so
  in a tighter-binding context (`'v=' . is_null($x)`, `is_null($x) == $y`)
  the rewrite changes meaning or does not parse. custos wraps the fix text
  (not the message) in parentheses when the target is an operand of a
  binary operator other than `&&`, `||`, `and`, `or`, `xor`, `??`, of a
  unary operator or of `instanceof`. *A* is also wrapped when it is a
  low-precedence keyword expression (`include`, `print`, `yield`, `throw`,
  arrow function): `is_null(include 'x.php')` → `(include 'x.php') === null`.
