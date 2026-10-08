---
id: SuspiciousBinaryOperation
group: Probable bugs
kind: semantic
needs: [names, index, types, stubs]
php: { min: "", max: "" }
---

# SuspiciousBinaryOperation

## Summary
Flags binary operations that are almost certainly not what the author meant:
`instanceof` against a trait, identical operands, `==` used as a statement,
`>=` typed instead of `=>` in an array, a comparison placed inside a call's
parentheses, negated comparisons on nullable values, `??` after a cast or
`!`, concatenation with an array literal, hard-coded booleans in `&&`/`||`,
and mixed `&&`/`||`/assignment precedence.

## Detection
Every binary expression is examined on its own (nested binary expressions get
their own pass). For one binary expression the checks below are tried **in
this order**, and the first one that reports stops the others for that
expression. Checks 9 and 10 run only when their option is enabled.

Common definitions:
- *strip(x)*: remove any number of wrapping parentheses from `x`.
- *Equivalent* (`≡`): same node kind and, for simple variables, same name;
  otherwise the same token sequence ignoring whitespace/comments, or identical
  source text. (`rand() ≡ rand()` holds: no purity check.)
- *Comparison operators*: `==`, `!=`, `<>`, `===`, `!==`, `<`, `>`, `<=`,
  `>=`, `<=>`.
- `&&`/`||` below mean exactly the symbolic operators; `and`/`or` are
  separate operators unless stated.

1. **D1 instanceof a trait.** Operator `instanceof`; the right operand is a
   class name (not a variable/expression) whose text is not `self` or
   `static`; the name resolves (current namespace + imports, then project
   index) to a **trait**. Report the whole binary expression.
2. **D2 `==` as a statement.** Operator `==` and the binary expression is
   the whole expression of an expression statement (`$a == $b;`). Report the
   `==` token.
3. **D3 `>=` in an array.** Operator `>=`, the left operand is directly a
   string literal, and the expression is a plain (non-keyed) element of an
   array literal (`['k' >= 1]`, `array('k' >= 1)`; not a key, not the value of
   a keyed element). Report the `>=` token.
4. **D4 negated comparison on a nullable value.** Operator `<` or `<=`;
   going up from the expression through any number of enclosing parentheses,
   the first non-parenthesis ancestor is a unary `!`; the left operand's type
   is fully resolved (no unknown part) and contains `null` or a boolean
   (`bool`, `true`, `false`). Report the whole `!` expression (from `!` to the
   last closing parenthesis). Suggested replacement `R` = `{left} >= {right}`
   for `<`, `{left} > {right}` for `<=` (operand source texts verbatim, single
   spaces).
5. **D5 identical operands.** Operator in `== != <> === !== > >= < <=
   instanceof`; `strip(left) ≡ strip(right)`. Report the whole binary
   expression (including parentheses that are inside it).
6. **D6 comparison misplaced inside a call.** Operator in `== != === !== >
   >= < <=` (not `<>`, not `<=>`), and:
   - the expression is directly an argument of a function or method call
     `C`, and it is the **last** argument;
   - `C` is used as a logical operand: going up through parentheses, the
     parent is the condition of `if`/`elseif`/`while`/`do-while`, the operand
     of `!`, an operand of `&&`/`||`/`and`/`or`, or the condition of a full
     (non-short) ternary;
   - `C` resolves to a function/method declaration with at least as many
     declared parameters as `C` has arguments;
   - the declared type set of the parameter at the last argument's position
     (unknown parts dropped) does not contain `bool` (an untyped parameter
     qualifies);
   - every type of the right operand (unknown parts dropped) is among the
     return types of the callee (unknown parts dropped). When nothing is
     known about the right operand (empty type set after dropping unknown
     parts), the check does not apply.
   Report the operator token.
7. **D7 `??` after a cast or `!`.** Operator `??`; `strip(left)` is a unary
   expression whose operator is `!` or a cast (`(int)`, `(integer)`, `(bool)`,
   `(boolean)`, `(float)`, `(double)`, `(real)`, `(string)`, `(binary)`,
   `(array)`, `(object)`, `(unset)`). Report `strip(left)` (without the
   wrapping parentheses); message includes its source text.
8. **D8 concatenation with an array literal.** Operator `.` and the left or
   the right operand is directly an array literal (`[...]` or `array(...)`,
   no parenthesis stripping). Report the `.` token.
9. **D9 hard-coded boolean operand** (option `VERIFY_CONSTANTS_IN_CONDITIONS`).
   Operator `&&` or `and`, or `||` or `or`. Look at the left operand, then the
   right operand (direct operands only, no parenthesis stripping); report only
   the first one that matches:
   - for `&&`/`and`: `false` or `null` → "decides the result"; `true` →
     "useless";
   - for `||`/`or`: `true` → "decides the result"; `false` or `null` →
     "useless".
   Constants are matched case-insensitively and may be `\`-qualified. Report
   the constant operand.
10. **D10 unclear precedence** (option `VERIFY_UNCLEAR_OPERATIONS_PRIORITIES`).
    - **D10a** Operator `&&` or `||` and the direct parent is a binary
      expression whose operator is the *other* one of `&&`/`||` (`$a || $b &&
      $c`, `$a && $b || $c`). Report this (inner) expression.
      Likewise for the keyword forms: operator `and`, `or` or `xor` whose
      direct parent is a binary expression with a *different* one of
      `and`/`or`/`xor` (`$a and $b or $c`, `$a or $b and $c`,
      `$a xor $b or $c`). Mixing a keyword form with `&&`/`||` is not
      reported (`$x = f() or die()`-style idioms stay quiet).
    - **D10b** Operator `&&` or `||` and the direct parent is a plain `=`
      assignment (this expression is its right-hand side) that is **not** the
      whole expression of an expression statement (`if ($x = f() && $y)`,
      `return $x = $a || $b;`, `g($x = $a && $b)`). Report this expression.
      For chains like `f() && $b && $c = g()` only the outermost `&&` is
      reported (inner ones have a same-operator parent).
    - **D10c** A comparison operator, and the direct parent is a plain `=`
      assignment whose own parent is an `if` statement (the assignment is the
      `if` condition itself; `elseif`/`while` do not count). Report the whole
      **assignment**.
    - **D10d** Otherwise, operator `<`, `>`, `<=` or `>=` (not `<=>`) and the
      left operand is directly a unary `!` expression (`! $a > $b`,
      `!($a) > $b`). Report the whole binary expression.
- **Name case.** Wherever this rule compares two expressions for
  equivalence, the names PHP resolves case-insensitively — function and
  method names, class names in calls, `new`, `instanceof` and `::`
  accesses, and keywords — compare case-insensitively (`Cache::$map['k']`
  matches `cache::$map['k']`, `$o->Name()` matches `$o->name()`); variable,
  property and constant names stay case-sensitive.

## Exceptions (no report)
- **E1** `instanceof self`, `instanceof static`, `instanceof $var`, unresolved
  names, classes and interfaces.
- **E2** `==` that is part of a larger expression; `===` statements.
- **E3** `>=` whose left operand is not a string literal, or that is used as
  an array key or keyed value.
- **E4** `!($v < 5)` when `$v`'s type is unknown/partially unknown or has no
  null/bool part; `!$v < 5` (the `!` is the operand, not the parent).
- **E5** Different operands after paren stripping.
- **E6** Call whose last parameter accepts `bool`; comparison that is not the
  last argument; right operand of unknown type (`f($x === $untyped)`); call not in a logical context; unresolved callee; more
  arguments than declared parameters (variadics).
- **E7** `??` whose left side is any other expression.
- **E8** `.` with array-typed variables (only literals count).
- **E9** Already-parenthesized groups (`$a || ($b && $c)`, `(!$a) > $b`);
  `$z = $a && $b;` as a statement; `!$a <=> $b`.

## Report
Severity: error for every check (upstream shows the "useless" variant of D9
with a strike-through style, but its severity is still error).

| Check | Range | Message |
|---|---|---|
| D1 | whole binary expression | `A trait is never an instanceof target; this is always false.` |
| D2 | `==` token | `Comparison result is discarded; did you mean '='?` |
| D3 | `>=` token | `Did you mean '=>' for an array key?` |
| D4 | whole `!(...)` expression | `Null or false operands make this negated comparison misleading; use '{R}'.` |
| D5 | whole binary expression | `Both operands are the same.` |
| D6 | operator token | `This comparison probably belongs outside the call parentheses.` |
| D7 | stripped left operand | `'{text}' is never null, so '??' is useless; add parentheses.` |
| D8 | `.` token | `Concatenating an array literal makes no sense.` |
| D9 decides | the constant | `This constant decides the whole condition.` |
| D9 useless | the constant | `This constant has no effect in the condition.` |
| D10 | as described in D10a–D10d | `Operator precedence is unclear here; add parentheses.` |

## Fix
- **F1 (D4)** Replace the whole `!(...)` expression with `R`
  (`!(($cnt <= 3))` → `$cnt > 3`).
- **F2 (D6)** Replace the whole call `C` with
  `{C'} {op} {right}` where `C'` is the call's source text in which the
  comparison's text is replaced by the left operand's text, `{op}` is the
  operator text and `{right}` the right operand text, separated by single
  spaces (`strlen($s >= 2)` → `strlen($s) >= 2`).
- **F3 (D10a, D10b)** Wrap the reported expression in parentheses, text kept
  verbatim (`$a && $b || $c` → `($a && $b) || $c`).
- **F4 (D10c)** Replace the assignment with its own text in which the
  right-hand side's text is wrapped in parentheses
  (`$r = $a !== $b` → `$r = ($a !== $b)`).
- **F5 (D10d)** Replace the expression with its text in which the left `!`
  operand is wrapped in parentheses, and any whitespace directly after that
  `!` removed (`! $a > $b` → `(!$a) > $b`; `!($a) > $b` → `(!($a)) > $b`).
  The whitespace removal reproduces upstream's formatter output, which is
  visible in the compared result.
- D1, D2, D3, D5, D7, D8, D9 have no fix.

## Options
| Option | Type | Default | Effect |
|---|---|---|---|
| `VERIFY_CONSTANTS_IN_CONDITIONS` | bool | `true` | Enables D9. |
| `VERIFY_UNCLEAR_OPERATIONS_PRIORITIES` | bool | `true` | Enables D10. |

## PHP versions
No gating. Upstream fixtures run at the test-default level (below 7.1) but use
`??` and scalar type hints; parse them regardless of the configured level.

## Examples

```php
<?php
trait Loggable {}
interface Shape {}

function checks($obj, $p, $q, array $tags, ?int $limit) {
    $t1 = <error descr="A trait is never an instanceof target; this is always false.">$obj instanceof Loggable</error>;
    $t2 = $obj instanceof Shape;
    $s1 = <error descr="Both operands are the same.">$p->id === ($p->id)</error>;
    $s2 = <error descr="Both operands are the same.">count($tags) > count($tags)</error>;
    $p <error descr="Comparison result is discarded; did you mean '='?">==</error> $q;
    $map = ['enabled' <error descr="Did you mean '=>' for an array key?">>=</error> true, $p >= 'x'];
    $out = [
        <error descr="'(bool)$p' is never null, so '??' is useless; add parentheses.">(bool)$p</error> ?? 0,
        (<error descr="'!$q' is never null, so '??' is useless; add parentheses.">!$q</error>) ?? 1,
        $p ?? $q,
    ];
    $msg = 'tags: ' <error descr="Concatenating an array literal makes no sense.">.</error> ['a'];
    $ok = [
        $p and <error descr="This constant decides the whole condition.">FALSE</error>,
        $p || <error descr="This constant decides the whole condition.">true</error>,
        $p && <error descr="This constant has no effect in the condition.">\true</error>,
        $p or <error descr="This constant has no effect in the condition.">null</error>,
    ];
    if (<error descr="Null or false operands make this negated comparison misleading; use '$limit >= 10'.">!($limit < 10)</error>) {}
}

class Quota {
    public function fits(string $who): bool {
        if (strlen($who <error descr="This comparison probably belongs outside the call parentheses.">>=</error> 3)) {}
        if ($this->allow($who === 'root')) {}
        return true;
    }
    private function allow(bool $flag) { return $flag; }
}

if (<error descr="Operator precedence is unclear here; add parentheses.">$found = $left != $right</error>) {}
if ($x || <error descr="Operator precedence is unclear here; add parentheses.">$y && $z</error>) {}
if ($row = <error descr="Operator precedence is unclear here; add parentheses.">fetch() || $fallback</error>) {}
if (<error descr="Operator precedence is unclear here; add parentheses.">!  $x >= $limit</error>) {}
if (($x && $y) || $z) {}
$both = $x && $y;
if ((!$x) < $y) {}
```

```php
<?php
trait Loggable {}
interface Shape {}

function checks($obj, $p, $q, array $tags, ?int $limit) {
    $t1 = $obj instanceof Loggable;
    $t2 = $obj instanceof Shape;
    $s1 = $p->id === ($p->id);
    $s2 = count($tags) > count($tags);
    $p == $q;
    $map = ['enabled' >= true, $p >= 'x'];
    $out = [
        (bool)$p ?? 0,
        (!$q) ?? 1,
        $p ?? $q,
    ];
    $msg = 'tags: ' . ['a'];
    $ok = [
        $p and FALSE,
        $p || true,
        $p && \true,
        $p or null,
    ];
    if ($limit >= 10) {}
}

class Quota {
    public function fits(string $who): bool {
        if (strlen($who) >= 3) {}
        if ($this->allow($who === 'root')) {}
        return true;
    }
    private function allow(bool $flag) { return $flag; }
}

if ($found = ($left != $right)) {}
if ($x || ($y && $z)) {}
if ($row = (fetch() || $fallback)) {}
if ((!$x) >= $limit) {}
if (($x && $y) || $z) {}
$both = $x && $y;
if ((!$x) < $y) {}
```

## Divergences
- `?int $limit` resolves to `int|null`, fully known, so D4 fires; a parameter
  without a type is unknown and is skipped. This mirrors upstream.
- F4/F5 use plain text substitution upstream: *every* occurrence of the
  replaced text inside the reported node is wrapped. If the target text also
  occurs elsewhere in the node (e.g. `if ($v = $v == 1)` contains `$v == 1`
  only once, fine; `!$a > !$a` would wrap both), the result differs from a
  structural edit. Recommendation: do a structural edit (wrap only the
  intended node); identical for all upstream fixtures.
- **D10a keyword operators (custos diverges).** Upstream only looks at
  `&&` mixed with `||`, so the same ambiguity spelled with keywords
  (`$a and $b or $c`, `$a xor $b or $c`) goes unreported although `and`,
  `xor` and `or` also have three distinct precedence levels. custos reports
  an `and`/`or`/`xor` expression nested directly in a different one of
  those keywords and offers the same parenthesising fix. Keyword/symbol
  mixes (`$a && $b or $c`) stay silent.
- **D6 — custos diverges from upstream.** Upstream treats an empty
  right-operand type set as contained in any return type, so a comparison
  against an untyped variable or unresolved expression inside a logical call
  (`if (check($a === $b))` with `$b` untyped) is reported and the fix moves
  the comparison out of the call, although nothing suggests the callee's
  return value is what `$b` should be compared to. custos treats an empty
  set as "no evidence" and stays silent; a right operand with a known type
  still needs every type to be among the callee's return types.
- F5 whitespace removal: upstream's formatter may also normalise other
  spacing in the generated node; only the space after `!` is observable in
  fixtures. Keep the rest verbatim.
- **Name case (custos diverges).** Upstream compares the expressions
  textually, so operands that differ only in the case of a function, method
  or class name (`Stats::$n` vs `stats::$n`), which PHP treats as the same,
  are not recognised as equivalent. custos folds the case of those names
  (Detection, "Name case").
- **Operands that may vary (custos diverges from D5).** Two identical
  operands are not "the same" when evaluating them twice may give different
  values: method, static and nullsafe calls, `new`, `++`/`--`,
  assignments, calls to user-defined or unresolved functions, and
  built-ins with randomness, clocks or cursors (`mt_rand() == mt_rand()`,
  MediaWiki's `wfRandom() == wfRandom()`, Pimcore's
  `Asset::getById($id) === Asset::getById($id)`). Deterministic built-ins
  (`count($tags) > count($tags)`) stay reported.
