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

Analyzed units:

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
  must be an object-only type set. A variadic parameter (`?T ...$p`) is
  skipped: `$p` is always an array, only its elements are nullable. Then run
  the usage walk (below) for `p` with no declaration.

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
    Signatures of classes unavailable at the configured PHP version do not
    establish nullable member results. Drop unknown parts. A type that
    includes `mixed` (`@return Box|mixed|null`) is not reported: its null is
    no more specific than the `mixed` beside it. If any part is `null` or
    `void`, and no *null-tested*
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
usage whose parent is `A` (that one is skipped too). Usages inside `A`'s
right side are skipped as well (custos, see Divergences): they read the
value the variable held before `A`, such as a `foreach` value in
`foreach ($tree as $n) { $n = $this->find($n->children); }`. Then evaluate each
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
  - method call: reported (inside the declaring assignment `A` itself,
    `$n = $n->next();`, the usage was already skipped). Notably:
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
  Constructors are outside U11, even when the shared callee query can resolve
  their parameter contracts.
- Anything else (e.g. `v ?? x`, `v ?: x`, `return v`, plain read) → skip.
- **U12** (applies to every *report* of U7–U11) The report is dropped —
  the walk goes on without stopping — when the conditions that must have
  held to reach `v` prove `$n` is not null there, reading `$n` with its
  declared (native) type narrowed at `v`; a local whose type the declarations
  leave unknown is read as the nullable object D5 found (so
  `$row = $rows[$i] ?? null; if (Kind::A === $row?->kind) { $row->p; }` is
  silent). Such conditions guard the body of
  an enclosing `if`/`elseif`/`else` (an `elseif` or `else` body also implies
  every earlier condition of its chain false), `while`, or `for` (its body,
  and its loop expressions when neither the body nor an earlier loop
  expression assigns `$n`), an arm of a ternary or short ternary, the right
  operand of `&&`/`and`/`||`/`or`, a `match` arm (the earlier arms did not
  match), and the statements after an `if` whose body always leaves: by
  `return`, `throw`, `exit`, `continue`, `break`, a call that never returns,
  or an `if`/`else`, `try` or `switch` that leaves on every path.
  A condition proves `$n` non-null when true if it is `$n` or an access
  chain starting at `$n` (`$n->p`, `$n?->m()`, `$n['k']`) used as a truthy
  value, `$n instanceof T`, `isset(...)` of `$n` or such a chain, a builtin
  type predicate that excludes null (`is_object($n)`), `X !== null` or a
  loose comparison of `X` (`$n` or such a chain) with `null`, `''`, `0` or
  `true` that excludes null, `$n === E` where `E` is known never to be
  null, or `$n == E` where `E` is an object; it proves it when false if it
  is `empty(X)`, builtin `is_null($n)`, `X === null` or `X == null`; `!`,
  `&&` and `||` combine these. What a condition proved ends where `$n` may
  have changed since: an assignment to `$n` (any operator, destructuring
  included; a write to `$n->p` or `$n[…]` is not one) that can reach `v`
  without passing the condition again, a by-reference use, or, for a
  `static` local, a call that may run the function again.
- **U13** (custos, see Divergences) After a report, the walk *stops* when
  the reported usage throws on null and always runs before the code after
  it: it is a U8 method call, a U9 call of the variable, a U10 `clone` or a
  U11 argument (not a U7 element or U8 property read, which only warn on
  null); its statement is an expression statement, `return` or `echo`
  directly in the statement list that holds `A` (the body of `F` for a
  parameter); and on the
  way up to that statement no operator can skip it: it is not the right
  operand of `&&`, `||`, `and`, `or` or `??`, not a ternary branch, not
  inside a closure, and no nullsafe `?->` access lies above it.
  `$box->open(); $box->close();` reports the first call only: the second
  one only runs when the first did not throw.

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
  `match (true) { null !== $v?->m() => $v->m(), … }`, statements after
  `if (null === $v?->m()) { return; }`,
  `for (; $v; $v = $v->next()) { $v->p; }` and
  `match ($v) { null => 0, default => $v->p }`.
- **E10** Variadic parameters (`?T ...$v`): `$v[0]` indexes an array.
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
- **E11** Dereferences in code that control never reaches: a statement
  that follows, in its statement list (or in an enclosing one), a `return`,
  `throw`, `break`, `continue`, `exit` or another statement that never
  completes, unless a `goto` label lets a jump land in between.

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
            $item = $item->next();      // declaring statement of local $item: right side exempt
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

- **U13 — custos diverges from upstream.** Upstream reports every
  dereference of the same nullable value, even after one that always runs
  and would already have thrown.
- **E11 — custos diverges from upstream.** Upstream also reports
  dereferences in unreachable code, such as after a `break` in a `case`.
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
  on the same range upstream. custos reports each `->` once.
- Enclosing conditions (custos diverges from upstream). Upstream does not
  model control flow: a check it does not recognize as a null check
  (`is_object($v)`, `$v === $known`, `'v' === $v->name`, `$v?->p !== null`)
  merely skips, so a
  dereference inside the branch that this check guards is still reported.
  custos drops reports whose enclosing condition proves the variable
  non-null (U12). The rest of the walk is unchanged: dereferences outside
  that branch are still reported. The same holds after an early exit
  (`if (null === $v?->m()) { return; }`), in a `match (true)` arm, and in a
  loop whose condition is re-checked after a branch that reassigns the
  variable and then `continue`s; writing a property of the variable
  (`$v->p = 1`) does not undo the check. A `for` condition guards the loop
  body and, when the body never assigns the variable, the loop expressions;
  a `match ($v)` guards each arm after one that lists `null`, and an arm
  whose conditions are all known non-null values guards its own body.
- Variadic parameters (custos diverges from upstream). `?T ...$v` declares
  an array of nullable elements, never a nullable object, so `$v[0]` is not
  a null dereference; custos does not walk such parameters.
- **Named arguments (custos diverges, U11).** A named argument is matched
  to the parameter of that name, not to the parameter at its position:
  `Box::draw($on, tint: $v)` with
  `draw(?bool $on, Mode $mode = Mode::Idle, ?Tint $tint = null)` passes
  `$v` to `?Tint $tint`, which accepts null, not to `Mode $mode`.
- Whether compound assignments (`.=`, `??=`) and list destructuring count as
  "assignments" in D4/U6 is unverified upstream. In custos only plain and
  by-reference `=` with a variable on the left are declarations; a compound
  assignment to the variable is a non-stopping usage (skipped), except
  `??=` (below).
- **`??=` stops the walk (custos diverges).** `$b ??= $this->box;` leaves
  `$b` non-null when the right side is not nullable, exactly like
  `$b = $b ?? $this->box;` (which already stopped the walk), yet a later
  `$b->open()` was reported. In U6 custos treats `??=` to the variable like
  a plain `=`: it stops the walk unless the right side is nullable in the
  D5 sense (`$c ??= $this->next()` with `next(): ?Box` keeps walking).
  Other compound assignments stay non-stopping usages.
- **Locals assigned a tested call (custos diverges).** In D5, a right side
  that is a function or method call is not nullable when an identical call
  written before it was tested by the conditions guarding the assignment:
  `if (null !== $box->next()) { $n = $box->next(); $n->open(); }` and
  `if (lookup() === null) { return; } $n = lookup(); $n->open();` report
  nothing. A test that does not guard the assignment (an earlier `if` whose
  body has ended) still leaves the local nullable.
- **Reads inside the declaring assignment (custos diverges).** Upstream
  reports `$n->children` in `$n = $this->find($n->children);` once the
  declaration makes `$n` nullable, although that read happens before the
  assignment and sees the earlier value (a `foreach` or destructured value).
