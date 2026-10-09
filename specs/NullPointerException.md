---
id: NullPointerException
group: Probable bugs
kind: semantic
needs: [names, types, hierarchy, stubs]
php: { min: "", max: "" }
---

# NullPointerException

## Summary

A value that may be `null` (nullable object parameter, nullable local
variable, call declared to return `?T`/`void`) is dereferenced — property or
method access, array access, invocation, `clone`, or passed to a parameter
that rejects null — without a preceding null check. At runtime this is an
"call to a member function on null"-style error.

Disabled by default (experimental upstream).

## Detection

Analysed units:

- every **method** that is not abstract and whose file/class is not a test
  context (path ends with `Test.php`, `Spec.php`, `.phpt`, or contains
  `/Fixtures/`; class FQN ends with `Test` or contains `\Tests\`/`\Test\`);
- every **function and closure** (no test-context check).
Arrow functions have no `{}` body: only strategy B runs for them (and it finds
nothing of its own unless they contain method calls).

For each unit `F`, three strategies run in this order: A (parameters),
B (chained calls), C (local variables). Each of A and C keeps its own
"already reported" set of variable nodes so a node is reported at most once
per strategy.

Terminology:

- *object-only type set*: after removing `null`, the set is non-empty and
  every remaining type is a class/interface name (FQN), `self`, `static` or
  `object`. (`?int`, `mixed`, `\Foo|false`, untyped → not object-only.)
- *null check usage*: see U-rules below.

### Strategy A — nullable object parameters

- **D1** `F` has a `{}` body.
- **D2** For each parameter `$p`: its **declared** type (type hint only;
  docblocks are ignored) contains `null` (`?T`, `T|null`) **or** its default
  value is the `null` constant (any case). The declared types minus `null`
  must be an object-only type set. Then run the usage walk (below) for `p`
  with no declaration.

### Strategy B — chained calls on nullable results

- **D3** For every method call `X->m(...)` anywhere inside `F` (any depth,
  nested closures included, document/pre-order: an outer call is visited
  before the calls nested in it), using the plain `->` operator (not `?->`,
  not `::`):
  - if `X` is itself a call (function call, method call or static method
    call — not `new`, not a parenthesised expression), infer `X`'s return
    type (declared return type or `@return`, stubs for built-ins). Exclude
    null introduced solely by short-circuiting an earlier nullsafe access in
    this chain; keep null from the evaluated member's own result. A chain
    that certainly short-circuits has no evaluated result to dereference.
    Drop unknown parts. If any part is `null` or `void`, and no *null-tested*
    entry (below) has the same method/function name as `X` **and** is
    structurally identical to `X` (same text modulo whitespace) → report the
    `->` token of `X->m`.
  - afterwards (whether or not reported), record `X->m(...)` itself as a
    null-tested entry when:
    - its parent is a binary `==`, `!=`, `===`, `!==` whose other operand is
      the `null` constant; or
    - its parent is a binary `instanceof`, `&&` or `and` (either side); or
    - its parent is not a binary expression and it is used as a logical
      operand: possibly through parentheses, it is the condition of an
      `if`/`elseif`/`while`/`do-while`, the operand of `!`, or the condition
      of a full (non-short `?:`) ternary.
      (`||`/`or` parents do not record.)
  Entries are never removed: once a call is tested anywhere earlier in the
  unit, later identical chained calls are not reported.

### Strategy C — nullable local variables

- **D4** `F` has a `{}` body. Collect, in document order over the whole body
  (nested closures included), assignment expressions that are directly an
  expression statement (`$v = …;`), whose left side is a plain variable not
  named like a parameter of `F`, and whose right side is **not** a property
  fetch (`$this->p`, `$o->p`) and **not** a unary expression (casts, `!`,
  `-`, `clone`, `@`, …). Group them by variable name.
- **D5** For each name, take only the **first** collected assignment `A`. It
  is *nullable* when:
  - the inferred type of the right side contains `null` or `void`, and after
    removing those the set is object-only (so `$v = null;` alone is not
    nullable); and
  - type inference requirements (they decide several fixture cases):
    - parts of the inferred set that cannot be resolved are **dropped**; the
      remaining known parts decide (an unknown part never makes the whole
      set unknown/non-nullable on its own);
    - `x ?? y` has the types of `x` united with the types of `y`;
      `c ?: y` has the types of `c` united with the types of `y`;
      `c ? x : y` has the types of `x` united with those of `y` (removing
      `null`/`false` from the left operand of `??`/`?:` is acceptable, as
      the right operand is what supplies `null` in the cases that matter);
    - a variable read inside the right side of its **own** assignment
      (`$p = $p ?: null;`, `$p = $p ?? null;`) has the types the variable
      could hold *before* that assignment: the parameter's declared type
      (with `null` included when the parameter is `?T`, `T|null`, or has a
      `null` default) and/or earlier assignments in the unit. The assignment
      being evaluated must never feed its own type (no self-recursion that
      collapses to "unknown").
    So `$p = $p ?? null;` with `\Foo $p = null` infers `\Foo|null` →
    nullable.
  - docblock override: if `A` is a plain `=` assignment and the statement
    immediately preceding it is a doc comment with exactly one `@var` tag
    naming this same variable, `A` stays nullable only if that tag's types
    include `null`.
- **D6** If nullable, run the usage walk for the name with declaration `A`.

### Usage walk (shared by A and C)

Build the ordered usage list for name `n` from the variables named `n` in
`F`'s body whose innermost function-like is `F` itself (document order):

- if the variable's parent is an assignment: append the variables named `n`
  found **inside** its right side (any depth, including inside nested
  closures, excluding the right side itself when it is that variable), then
  the remaining variables named `n` anywhere in the assignment (typically the
  left side and a bare right-side variable) not already appended;
- otherwise append every variable named `n` found strictly inside the
  variable's parent (normally just the variable itself).
Duplicates are possible and harmless (reports are de-duplicated).

When a declaration `A` is given (strategy C), skip usages until reaching a
usage whose parent is `A` (that one is skipped too). Then evaluate each
remaining usage `v` (parent `P`, grandparent `G`) in order; *stop* ends the
walk for this name, *skip* goes to the next usage:

- **U1** `P` is binary `instanceof` → stop.
- **U2** `P` is binary `==`/`!=`/`===`/`!==`: other operand is `null` → stop;
  otherwise skip.
- **U3** `P` is `empty(...)` or `isset(...)`, or `v` is used as a logical
  operand (definition in D3: `if`/`elseif`/`while`/`do-while` condition,
  operand of `!`, full-ternary condition, operand of `&&`/`||`/`and`/`or`,
  all through parentheses) → stop.
- **U4** `P` is a `catch` clause (`catch (E $n)`) → stop.
- **U5** `v` is an argument (`P` is an argument list) and `G` is either:
  - a method call of **any** flavour — instance `->`, nullsafe `?->` or
    static `::` (`$this->assertNotNull($v)`, `self::assertNotNull($v)`,
    `Assertion::notNull($v)`) — named `assertNotNull`, `assertInstanceOf`,
    `notNull`, `isInstanceOf` or `isInstanceOfAny`; or
  - a method call of any flavour (again including static:
    `Assert::that($v)`, `$x->that($v)`) named `that`, where climbing from
    `G` through its directly enclosing method calls (each one having the
    previous as its object, any operator) reaches a call named `notNull`
    (`Assert::that($v)->notNull()`, `Assert::that($v)->x()->notNull()`); or
  - a plain function call `is_null(...)`
  → stop. Method names (`assertNotNull`, …, `that`, `notNull`) and
  `is_null` are matched case-insensitively, like PHP itself
  (`self::AssertNotNull($v)` stops the walk); see Divergences.
- **U6** `P` is an assignment: if `P` is `A` → skip. If `P`'s left side is a
  plain variable named `n` and `P` is not nullable in the D5 sense (same
  type rules, including the self-reference rule and the docblock override)
  → stop (variable overwritten with a non-null value). Otherwise skip.
  Consequence: `$p = $p ?: null;` and `$p = $p ?? null;` on a nullable
  object parameter/local are *nullable* re-assignments, so they do not stop
  the walk and a later `$p->prop` is reported. The occurrence of `$p` inside
  the right side (parent `?:`/`??`) is itself a skip (a short `?:`
  condition is not a logical operand, `??` is not a comparison).
- **U7** `P` is an array access and `v` is its base (`$n[...]`) → **report**
  `v`. (`$x[$n]`: `v` is the index → nothing.)
- **U8** `P` is a member access using plain `->` (not `?->`, not `::`) whose
  object is `v`:
  - property fetch: climb from `P` through enclosing property fetches and
    array accesses (`$n->a->b[0]->c`); if the topmost such node is the left
    operand of `??`, or is directly an argument of `isset(...)` → skip.
    (A method call in the chain stops the climb: `isset($n->a->m()->b)` is
    reported.)
  - method call: if a declaration `A` was given (strategy C only) and `G`
    is that very `A` (the declaring statement is `$n = $n->next();`) →
    skip. In every other situation the method call is reported, notably:
    strategy A (parameters — there is no `A`, so `$p = $p->next();` reports
    the right-side `$p`), and later, non-declaring re-assignments of a local
    (`$n = $n->next();` after the declaration reports the right-side `$n`
    unless the walk was already stopped).
  - otherwise **report** `v`.
- **U9** `P` is a plain function-call expression whose callee is `v`
  (`$n(...)`) → **report** `v`.
- **U10** `P` is `clone v` → **report** `v`.
- **U11** `v` is an argument of a function or method call `G` that resolves:
  take the callee parameter at `v`'s position (positional index among the
  call's arguments; none → skip), or, for a named argument (`tint: $v`),
  the parameter of that name (none → skip). If that parameter's declared type does not
  contain `null`, its default is not `null`, and its declared types form an
  object-only set → **report** `v`.
- Anything else (e.g. `v ?? x`, `v ?: x`, `return v`, plain read) → skip.
- **U12** (applies to every *report* of U7–U11) The report is dropped —
  the walk goes on without stopping — when `v` sits in a region that an
  enclosing condition `c` (between `v` and `F`) only enters when `$n` is not
  null, and `$n` is not the target of an assignment (any operator,
  destructuring included; a write to `$n->p` or `$n[…]` is not one) or a
  `foreach` key/value that completes between the end of `c` and `v`.
  An assignment does not count when it cannot reach `v` without passing
  `c` again: a statement list that holds it but not `v` then ends (after
  it) in `return`, `throw` or `exit`, or in `continue`/`break` whose
  nearest enclosing loop (no `switch` in between) holds `c`.
  Regions and the truth value of `c` they imply:
  the body of `if (c)` / `elseif (c)` / `while (c)` (true); an `elseif` body
  also implies every earlier `if`/`elseif` condition false; the `else` body
  implies every condition of its chain false; ternary `c ? v : …` (true),
  `c ? … : v` and `c ?: v` (false); the right operand of `c && v` /
  `c and v` (true) and of `c || v` / `c or v` (false); the body of the arm
  `c => v` of `match (true)` when it has the single condition `c` (true);
  the statements following, in the same statement list, an `if (c)` whose
  body always terminates (`return`, `throw`, `exit`, `continue`, `break`)
  (false).
  `c` *guarantees non-null when true* if, after removing parentheses:
  `!x` guarantees it when `x` is false; `a && b` when `a` or `b` does;
  otherwise `c` is `$n` itself or an access chain starting at `$n`
  (`$n->p`, `$n?->m()`, `$n['k']`, any depth) used as a truthy value;
  `$n instanceof T`; `isset(...)` with an argument that is `$n` or such a
  chain; `is_object($n)` (global function); `X !== null` / `X != null` with
  `X` being `$n` or such a chain; `X === E` (`X` being `$n` or such a
  chain: a null `$n` makes the chain null or throws) where `E`'s inferred
  type is known and contains neither `null`, `void` nor `mixed`; `X == E` where
  `E`'s type is additionally object-only (a loose comparison with `''`,
  `0`, `false` or `[]` also holds for `null`).
  `c` *guarantees non-null when false* if: `!x` when `x` guarantees it when
  true; `a || b` when `a` or `b` does; `empty(X)`, `is_null($n)` (global),
  `X === null` / `X == null` with `X` being `$n` or a chain starting at it.

## Exceptions (no report)

- **E1** Parameters without a type hint (docblocks never count), or nullable
  parameters whose type is scalar/mixed/union with non-class members.
- **E2** Usages after a null check: `null !== $v`, `$v instanceof T`,
  `isset($v)`, `!empty($v)`, `if ($v)`, `$v ? … : …`, `$a && $v`,
  `is_null($v)`, assertion calls (instance or static, including
  `Assert::that($v)->notNull()`, any letter case), `catch` re-definition.
- **E2b** Dereferences inside a branch whose enclosing condition proves
  non-null without being one of the checks above (U12):
  `if (is_object($v)) { $v->p; }`, `$v?->p !== null ? $v->p : 0`,
  `if ($v === $known) { $v->p; }`, `else` of `if (!is_object($v))`,
  `match (true) { null !== $v?->m() => $v->m(), … }`, and statements after
  `if (null === $v?->m()) { return; }`.
- **E2c** A nullable value passed by name to a parameter that accepts null
  (`f(tint: $v)` with `?Tint $tint`), whatever its position.
- **E3** `$v->prop ?? …`, `isset($v->prop)`, `isset($v->a['k'])`.
- **E4** `$v = $v->m();` only when that statement is the local variable's
  own declaring assignment (strategy C, first collected assignment). It is
  *not* an exception for parameters nor for later re-assignments (see U8);
  in practice such loops are silent only because a preceding null check
  (`while ($v !== null)`) stopped the walk.
- **E5** Local variables whose first statement-level assignment is `null`
  only, a property fetch, a unary expression, or non-object nullable; local
  variables annotated `@var T $v` without `null`.
- **E6** Usages before the declaring assignment (strategy C).
- **E7** Nullsafe `?->` accesses; `$v::CONST`; static calls.
- **E8** Chained calls on a call previously null-tested in the unit;
  chained calls whose base is a parenthesised expression or a variable.
- **E9** Abstract methods; methods in test contexts.

## Report

- Range: strategies A and C — the variable token `$name` (just the
  variable); strategy B — the `->` operator token between the nullable call
  and the next method name.
- Severity: warning.
- Message: `Possible null dereference.`

## Fix

None.

## Options

None.

## PHP versions

None as such; nullable types (`?T`) need PHP 7.1, union types PHP 8.0. The
EA fixture runs at PHP 7.1.

## Examples

```php
<?php
class Node
{
    public $value;
    public function next(): ?Node { return null; }
    public function take(Node $n) {}
    public function walk(?Node $head, Node $tail = null, ?int $count = null, Node $plain, $loose = null)
    {
        <warning descr="Possible null dereference.">$head</warning>->value = 1;
        <warning descr="Possible null dereference.">$head</warning>();
        if ($head === null) {
            return;
        }
        $head->value = 2;

        $x = clone <warning descr="Possible null dereference.">$tail</warning>;
        $this->take(<warning descr="Possible null dereference.">$tail</warning>);
        echo $tail->value ?? 'none';
        if (isset($tail->value['a'])) {}
        if ($tail) {
            $tail->value = 3;
        }
        $plain->value = 4;
        $loose->value = 5;
        $count++;
    }

    public function chain()
    {
        $this->next()<warning descr="Possible null dereference.">-></warning>next();
        if ($this->next() != null) {
            $this->next()->next();
        }
        $n = $this->next();
        <warning descr="Possible null dereference.">$n</warning>->next();
        <warning descr="Possible null dereference.">$n</warning>['k'];
        $n = new Node();
        $n->value = 0;

        /** @var Node $m */
        $m = $this->next();
        $m->value = 1;

        $cursor = $this->next();
        while ($cursor !== null) {
            $cursor = $cursor->next();
        }

        $probe = $this->next();
        <warning descr="Possible null dereference.">$probe</warning>->value = 1;
        $probe = <warning descr="Possible null dereference.">$probe</warning>->next();
    }

    public function refill(Node $left = null, ?Node $right, Node $up = null)
    {
        $left = $left ?: null;
        <warning descr="Possible null dereference.">$left</warning>->value = 1;

        $right = $right ?? null;
        <warning descr="Possible null dereference.">$right</warning>->next();

        $up = <warning descr="Possible null dereference.">$up</warning>->next();
    }

    /** @param Node[] $list */
    public function rewind(array $list)
    {
        foreach ($list as $item) {
            $item = $item->next();      // declaring statement of local $item: right side exempt (E4)
            <warning descr="Possible null dereference.">$item</warning>->value = 1;
        }
    }

    public function guarded(?Node $a, ?Node $b, ?Node $c)
    {
        static::assertNotNull($a);
        Verify::that($b)->isObject()->notNull();
        $this->that($c)->notNull();

        return [$a->value, $b->value, $c->value];
    }
}

function visit(?Node $item) {
    \Assert\Assertion::notNull($item);
    return $item->value;
}
```

## Divergences

- Case-insensitive assertion names (custos diverges from upstream).
  Upstream compares the assertion method names (U5) and `is_null`
  case-sensitively, so `self::AssertNotNull($v)` or `IS_NULL($v)` — which
  PHP calls just the same — does not stop the walk and the following
  dereferences are reported. custos matches them case-insensitively.
- An earlier version of this spec listed `$v = $v->next();` as a blanket
  exception. Upstream only exempts it when it is the local's declaring
  statement (U8/E4); parameters and later re-assignments are reported.
- Strategy B scans nested closures both from the enclosing unit and from the
  closure itself, so a chained call inside a closure could be reported twice
  on the same range upstream. Recommendation: report each `->` once.
- Enclosing conditions (custos diverges from upstream). Upstream does not
  model control flow: a check it does not recognise as a null check
  (`is_object($v)`, `$v === $known`, `'v' === $v->name`, `$v?->p !== null`)
  merely skips, so a
  dereference inside the branch that this check guards is still reported.
  custos drops reports whose enclosing condition proves the variable
  non-null (U12). The rest of the walk is unchanged: dereferences outside
  that branch are still reported. The same holds after an early exit
  (`if (null === $v?->m()) { return; }`), in a `match (true)` arm, and in a
  loop whose condition is re-checked after a branch that reassigns the
  variable and then `continue`s; writing a property of the variable
  (`$v->p = 1`) does not undo the check.
- **Named arguments (custos diverges, U11).** A named argument is matched
  to the parameter of that name, not to the parameter at its position:
  `Box::draw($on, tint: $v)` with
  `draw(?bool $on, Mode $mode = Mode::Idle, ?Tint $tint = null)` passes
  `$v` to `?Tint $tint`, which accepts null, not to `Mode $mode`.
- Whether compound assignments (`.=`, `??=`) and list destructuring count as
  "assignments" in D4/U6 is unverified upstream; recommendation: only plain
  and by-reference `=` with a variable on the left are declarations; a
  compound assignment to the variable is a non-stopping usage (skip).
- **`??=` stops the walk (custos diverges).** `$b ??= $this->box;` leaves
  `$b` non-null when the right side is not nullable, exactly like
  `$b = $b ?? $this->box;` (which already stopped the walk), yet a later
  `$b->open()` was reported. In U6 custos treats `??=` to the variable like
  a plain `=`: it stops the walk unless the right side is nullable in the
  D5 sense (`$c ??= $this->next()` with `next(): ?Box` keeps walking).
  Other compound assignments stay non-stopping usages.
