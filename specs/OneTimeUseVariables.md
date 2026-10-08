---
id: OneTimeUseVariables
group: Control flow
kind: semantic
needs: [flow]
php: { min: "", max: "" }
---

# OneTimeUseVariables

## Summary
A local variable that is assigned in one statement and consumed exactly once in
the very next `return`, `throw` or destructuring statement adds a name without
adding meaning. The value can be used directly at the consumption site.

## Detection
Three *consumer* statement shapes are inspected (each gated by an option):

- **D1 (return)** — option `ANALYZE_RETURN_STATEMENTS`. A `return <arg>;`
  statement. Skip it when the nearest enclosing *named* function or method is
  declared to return by reference (`function &name()`, a `&` token right before
  the name, whitespace allowed between). Closures (which have no name) are not
  subject to this check, and a `return` at file level has no enclosing function
  so the check does not apply.
- **D2 (throw)** — option `ANALYZE_THROW_STATEMENTS`. A `throw <arg>;`
  statement (the `throw` is the whole expression of the statement).
- **D3 (destructuring)** — option `ANALYZE_ARRAY_DESTRUCTURING`. An expression
  statement whose expression is a destructuring assignment written with
  `list(...) = <arg>` or `[...] = <arg>`, and that assignment is the direct
  expression of the statement (`list($p, $q) = $pair;`, `[$p, $q] = $pair;`).
  Destructuring inside `foreach` or nested in another expression is not
  inspected.

- **D4 (subject variable)** — after removing any parentheses around `<arg>`,
  the subject variable `V` is:
  - `<arg>` itself if it is a plain variable `$name`;
  - or, if `<arg>` is a property fetch or method call whose object part is a
    plain variable and whose operator is `->` (also `?->`): that object
    variable. Examples: `$v->field`, `$v->run()`, `$v->$dyn`, `$v->{'k'}`.
  - Anything else yields no subject (`$v::$p`, `$v::m()`, `$v->a->b`,
    `$v[0]`, `"$v"`, `-$v`, `(string) $v`, `f($v)`, literals, arrays, …) and
    nothing is reported.
  Variable-variables (`$$n`) have no name and are ignored.

- **D5 (preceding assignment)** — take the consumer statement's previous
  sibling statement in the same statement list, skipping whitespace and
  ordinary comments (`//`, `#`, `/* */`). If that sibling is a doc comment
  (`/** */`) or there is no previous sibling, stop. The previous statement
  must be an expression statement whose *first/top* expression is a plain
  assignment using the `=` operator (by-reference `= &` counts too; compound
  assignments such as `+=`, `.=`, `??=` do not; a destructuring assignment
  does not). Its left-hand side must be a plain variable whose name equals
  `V`'s name (case-sensitive). Its value, with surrounding parentheses removed,
  is `VAL` (must exist).

- **D6 (by-reference bindings)** — let `S` be the nearest function-like scope
  (function, method, closure, arrow function) enclosing the consumer. If `S`
  has a parameter named `V` declared by reference (`&$v`), stop. If `S` is a
  closure whose `use (...)` list imports `V` by reference (`& $v`, `&$v`),
  stop. (By-value imports do not stop.)

- **D7 (length)** — when `ALLOW_LONG_STATEMENTS` is false (default), stop if
  the source text of the assignment expression (from the variable through the
  end of the value, without the `;`) is longer than 80 characters.

- **D8 (usage count)** — let `S'` be the nearest function-like scope enclosing
  the *assignment*, or the file's top-level code when there is none (code
  inside `namespace` blocks included). Walk every access to a variable named
  `V` in `S'`'s body (or in the top-level code) that is reachable from the
  scope's entry (no statement on the way from the access up to the scope is
  preceded in its statement list by a statement that always terminates —
  `return`, `throw`, `exit`, …; at file level the top-level statement list is
  checked too). Bodies of nested closures/functions/classes are separate
  scopes; a closure's `use ($v)` is a read in the outer scope. Classify
  each as read or write. Stop (no report) as soon as more than one write or
  more than one read has been seen. Also stop (no report) if a write occurs
  as the left-hand side of a plain `=` assignment whose statement is
  immediately preceded (ignoring whitespace) by a doc-style comment carrying
  exactly one `@var` tag whose variable is `$V` — accept both `/** @var T $v */`
  and the inline form `/* @var T $v */`. Writes include plain/compound
  assignment targets, destructuring targets, `foreach` value/key targets,
  `global`/`static` declarations, `catch` variables and `unset()`; everything
  else (including use as an object in `$v->p = …` or `$v->m()`, `isset`,
  string interpolation) is a read. The consuming occurrence itself counts as a
  read and the assignment itself counts as a write, so a match requires the
  assignment and the consumer to be the only reachable accesses.

- **D9 (version)** — if `VAL` is a `new` expression, report only when the
  configured PHP level is at least 5.4 (the fix may produce
  `(new X())->member`). Other values have no version requirement.

When D1/D2/D3 + D4–D9 all hold, report.

## Exceptions (no report)
- **E1** The function/method returns by reference (return consumers only).
- **E2** `V` is a by-reference parameter or a by-reference closure import of
  the consumer's scope.
- **E3** Within a function-like scope or at file level, `V` is read or
  written more than once in reachable code (e.g. the value is read again later, the variable is
  reassigned, the variable was a parameter that is then overwritten and read
  in the value of the assignment, the variable is destructured and then read
  again; at file level `$o = clone $o; return $o->x;` or a name reused by a
  later reachable statement).
- **E4** A write to `V` carries an inline `@var` type annotation (see D8).
- **E5** Compound assignment (`$v += 1; return $v;`).
- **E6** The consumer argument is not a plain variable or a single-level
  `->` access on one (static access, array access, interpolation, expression).
- **E7** The previous statement is not a plain assignment to the same
  variable, or is a doc comment, or there is no previous statement.
- **E8** Assignment text longer than 80 characters while
  `ALLOW_LONG_STATEMENTS` is off.
- **E9** `new` value under PHP < 5.4.

## Report
- Range: the left-hand-side variable of the assignment (`$name`, including the
  `$`) — not the consumer.
- Severity: warning.
- Message: `Variable ${name} is used only once; inline its value.`

## Fix
- **F1** If the assignment statement is immediately preceded (as its previous
  non-whitespace sibling) by a doc comment `/** … */`, delete that comment.
  (Whitespace before the comment is left as is.)
- **F2** Delete the whitespace run immediately following the assignment
  statement (so the consumer moves into the assignment's position).
- **F3** Replace the subject variable occurrence in the consumer with the
  source text of `VAL` (parentheses around the original value already
  stripped). When `VAL` is a `new` expression, a ternary (full or short
  `?:`), a `clone` expression or a `??` expression **and** the subject is the
  object of a `->` access (D4 second bullet), insert `(` + VAL + `)` instead.
  Otherwise insert `VAL` unchanged — no parentheses are added in any other
  case.
- **F4** Delete the assignment statement (including its `;`).

Resulting shapes:
- `$m = 7;⏎return $m;` → `return 7;`
- `$e = new Oops('x');⏎throw $e;` → `throw new Oops('x');`
- `$o = new Box();⏎return $o->size;` → `return (new Box())->size;`
- `$o = clone $proto;⏎return $o->reset();` → `return (clone $proto)->reset();`
- `$o = $a ?? $b;⏎return $o->id;` → `return ($a ?? $b)->id;`
- `$o = $a ?: $b;⏎return $o->id;` → `return ($a ?: $b)->id;`
- `$pair = [3, 4];⏎[$p, $q] = $pair;` → `[$p, $q] = [3, 4];`
- `$m = (1 + 2);⏎return $m;` → `return 1 + 2;`

## Options
| Option | Type | Default | Effect |
|---|---|---|---|
| `ALLOW_LONG_STATEMENTS` | bool | false | When false, assignments whose text exceeds 80 characters are left alone (a long expression may deserve a name). When true, length is ignored. |
| `ANALYZE_RETURN_STATEMENTS` | bool | true | Inspect `return` consumers (D1). |
| `ANALYZE_THROW_STATEMENTS` | bool | true | Inspect `throw` consumers (D2). |
| `ANALYZE_ARRAY_DESTRUCTURING` | bool | true | Inspect `list()`/`[]` destructuring consumers (D3). |

## PHP versions
- `new` values are reported only for PHP ≥ 5.4 (D9).
- Upstream positive fixture runs at PHP 7.0 with all four options true
  (including `ALLOW_LONG_STATEMENTS = true`); the false-positive fixture runs
  with default options at the IDE test default level (below 7.1). Neither
  depends on 7.1+ features.
- Short `[...] =` destructuring only parses from 7.1; `list()` everywhere.
  `?->` only exists from 8.0. `throw` as an expression (8.0) is only
  inspected when it is the whole statement.

## Examples

```php
<?php
<warning descr="Variable $total is used only once; inline its value.">$total</warning> = 40 + 2;
return $total;

function makeUser() {
    <warning descr="Variable $user is used only once; inline its value.">$user</warning> = new User();
    return $user->name;
}

function failWith($why) {
    <warning descr="Variable $err is used only once; inline its value.">$err</warning> = new DomainError($why);
    throw $err;
}

function split2() {
    <warning descr="Variable $parts is used only once; inline its value.">$parts</warning> = explode(':', 'a:b');
    [$head, $tail] = $parts;
    return $head . $tail;
}

function pick($a, $b) {
    <warning descr="Variable $chosen is used only once; inline its value.">$chosen</warning> = $a ?: $b;
    return $chosen->label;
}

function copyOf($proto) {
    <warning descr="Variable $dup is used only once; inline its value.">$dup</warning> = clone $proto;
    return $dup->fresh();
}

function &byRef() {
    $slot = null;
    return $slot;
}

function twice() {
    $n = compute();
    $n = $n * 2;
    return $n;
}

function readAgain() {
    $cfg = load();
    [$k] = $cfg;
    return $cfg ? $k : null;
}

function annotated() {
    /** @var Gadget $g */
    $g = factory();
    return $g;
}

function staticAccess() {
    $cls = whichClass();
    return $cls::build();
}

function compound($v) {
    $v .= '!';
    return $v;
}

$fn = function () use (&$acc) {
    $acc = 5;
    return $acc;
};

function refParam(&$out) {
    $out = 1;
    return $out;
}
```

```php
<?php
return 40 + 2;

function makeUser() {
    return (new User())->name;
}

function failWith($why) {
    throw new DomainError($why);
}

function split2() {
    [$head, $tail] = explode(':', 'a:b');
    return $head . $tail;
}

function pick($a, $b) {
    return ($a ?: $b)->label;
}

function copyOf($proto) {
    return (clone $proto)->fresh();
}

function &byRef() {
    $slot = null;
    return $slot;
}

function twice() {
    $n = compute();
    $n = $n * 2;
    return $n;
}

function readAgain() {
    $cfg = load();
    [$k] = $cfg;
    return $cfg ? $k : null;
}

function annotated() {
    /** @var Gadget $g */
    $g = factory();
    return $g;
}

function staticAccess() {
    $cls = whichClass();
    return $cls::build();
}

function compound($v) {
    $v .= '!';
    return $v;
}

$fn = function () use (&$acc) {
    $acc = 5;
    return $acc;
};

function refParam(&$out) {
    $out = 1;
    return $out;
}
```

## Divergences
- **File-level counting (custos diverges from upstream).** Upstream skips
  D8 entirely at file level, so it inlines `$r` even when the variable is
  read again later in the file, or when the value reads the variable itself
  (`$o = clone $o; return $o->x;` would become `return (clone $o)->x;`,
  reading an undefined `$o`). custos counts file-level accesses like a
  function body. Accesses after an unconditional top-level `return`/`throw`
  are unreachable and not counted, so consecutive dead-code pairs such as
  `$x = 1; return $x; $x = 2; return $x;` are still each reported.
- Whether parameters count as an initial write in D8 is not observable from
  upstream fixtures. Recommendation: do not count the parameter declaration
  itself; reads of the parameter inside the assigned value count as reads.
- Upstream decides reachability from its IDE control-flow graph. custos uses
  the statement-list approximation of D8 (an access is unreachable when a
  terminating statement precedes it in an enclosing list); it does not
  model constant conditions or loops.
- Wrapping in F3 is limited to `new`/ternary/`clone`/`??` values. Other
  low-precedence values used as the object of `->` (e.g. `$o = $a . $b;
  return $o->x;`, assignments, `instanceof`, arithmetic) are inlined without
  parentheses and change meaning. Recommendation: also wrap any value whose
  precedence is lower than member access (binary/unary/assignment/`yield`/
  `print`/`include`/closure); no upstream fixture covers this.
- A consumer such as `$v->m($v)` reads `V` twice, so D8 suppresses it at
  file level as well as in functions.
- **Nameless `@var` (custos diverges).** `/** @var \Illuminate\Auth\RequestGuard */
  $guard = $this->auth->guard('sanctum'); return $guard;` is the usual
  PHPStan form of an inline type annotation and counts like `@var T $guard`
  for E4/D8 (Monica middleware): inlining would drop the narrowed type.
  F1 also removes the whitespace between a deleted doc comment and the
  statement, so no indentation-only line is left.
