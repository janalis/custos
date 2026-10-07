---
id: UsingInclusionOnceReturnValue
group: Probable bugs
kind: syntax
needs: []
php: { min: "", max: "" }
---

# UsingInclusionOnceReturnValue

## Summary
`include_once`/`require_once` return the file's value only the first time;
every later evaluation returns `true`. Code that uses that value works once
and then silently breaks. Use `include`/`require` when the result matters.

## Detection
- **D1** An inclusion expression whose keyword is `include_once` or
  `require_once`, in any letter case (`REQUIRE_ONCE`, `Include_Once`; see
  Divergences), and that has an operand.
- **D2** The inclusion is **not** the whole expression of a plain expression
  statement — its direct parent is anything else: an assignment right-hand
  side, a `return`, an `if`/`while` condition, a parenthesized expression, a
  binary operand, a call argument, an array element, `@`, another inclusion's
  operand, etc.

Note the low precedence of inclusion: in `require_once $a && require_once $b`
the operand of the first inclusion is `$a && require_once $b`, so the outer
inclusion spans up to the end of `$b` and the inner one is reported as well
(its parent is the `&&` expression).

## Exceptions (no report)
- **E1** `require_once 'x.php';` / `include_once('y.php');` as standalone
  statements.
- **E2** `include`/`require` (non-`_once`) anywhere.

## Report
- Range: the inclusion expression, from the keyword to the end of its operand
  (the operand may span lines and contain other inclusions). Surrounding
  parentheses and the `;` are not included.
- Severity: error.
- Message: `Only the first include_once/require_once returns the file's value; later ones return true.`

## Fix
- **F1** Replace the inclusion with `include {operand}` (for `include_once`)
  or `require {operand}` (for `require_once`): the new keyword, one space,
  then the operand's source text verbatim (whitespace/comments that were
  between the keyword and the operand are dropped). A parenthesized operand
  keeps its parentheses: `require_once($f)` → `require ($f)`.
- **F2** Nested reports are all fixed: the outer replacement keeps the inner
  inclusion's text, which is then replaced too
  (`require_once $a && require_once $b` → `require $a && require $b`).
  The new keyword is always lowercase: `INCLUDE_ONCE $f` becomes
  `include $f`.

## Options
None.

## PHP versions
No gating.

## Examples

```php
<?php
$routes = <error descr="Only the first include_once/require_once returns the file's value; later ones return true.">require_once __DIR__ . '/routes.php'</error>;
$ok = <error descr="Only the first include_once/require_once returns the file's value; later ones return true.">include_once $first ||
    <error descr="Only the first include_once/require_once returns the file's value; later ones return true.">include_once $second</error></error>;
while (<error descr="Only the first include_once/require_once returns the file's value; later ones return true.">include_once($plugin)</error>) {
    register((<error descr="Only the first include_once/require_once returns the file's value; later ones return true.">require_once $plugin</error>));
}
require_once __DIR__ . '/bootstrap.php';
include_once('helpers.php');
$cfg = (require 'config.php');
```

```php
<?php
$routes = require __DIR__ . '/routes.php';
$ok = include $first ||
    include $second;
while (include ($plugin)) {
    register((require $plugin));
}
require_once __DIR__ . '/bootstrap.php';
include_once('helpers.php');
$cfg = (require 'config.php');
```

## Divergences
- **Keyword case (custos diverges).** Upstream requires the keyword text to
  end with lowercase `_once`, so `$cfg = REQUIRE_ONCE $f;` is not reported
  although PHP keywords are case-insensitive and the value is just as
  unreliable. custos matches the keyword in any case (D1).
- Whether a parenthesized path is the inclusion's operand (`include_once($p)`
  → `include ($p)`) or part of the keyword syntax depends on the parser; the
  comparison is whitespace-collapsed, so `include ($p)` and `include($p)`
  differ. Upstream fixtures only use unparenthesized operands in fixed
  output. Recommendation: emit `include ($p)` (keyword, space, operand text).
- **Success tests (custos diverges).** `include_once`/`require_once`
  return `false` only when the file cannot be included, so a result that is
  only tested for success is reliable: custos does not report an inclusion
  (through parentheses and `@`) used as an `if`/`elseif`/`while`/`do-while`
  or ternary condition, as an operand of `!`, `&&`, `||`, `and`, `or`,
  `xor`, or compared with `false` (`===`, `!==`, `==`, `!=`) — e.g.
  `if (!include_once $lib)`, `(@include_once $f) !== false` (E3).
  Upstream reports them all.
- **No quick-fix (custos diverges).** Replacing `_once` with a plain
  `include`/`require` (upstream F1/F2) re-runs the file on every call, which
  redeclares the classes and functions it defines (fatal) and repeats its
  side effects. custos reports without a fix; whether to load the value
  another way is the author's decision.
