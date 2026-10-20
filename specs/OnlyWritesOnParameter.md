---
id: OnlyWritesOnParameter
group: Unused
kind: semantic
needs: [flow, types]
php: { min: "", max: "" }
---

# OnlyWritesOnParameter

## Summary

A parameter, closure import or local variable that is only ever written to
(elements appended, incremented, concatenated onto…) but never read is dead
work: the changes are lost when the function returns (by-value semantics).
Likewise a closure `use` import that the closure body never touches, or an
inline assignment whose target is never read afterwards, is unused.

## Detection

The rule has three entry points that all feed one shared *access analysis*
(defined after them). "Scope" means a function declaration, a non-abstract
method, or a closure (arrow functions behave like closures but cannot hold
the relevant statements in practice).

### Entry 1 — parameters

- **D1** For each scope, each parameter that
  - has a non-empty name,
  - is **not** by-reference (`&$p`),
  - and whose *native* declared type (the type hint; doc-block types are
    ignored) does **not** contain an object type. A type hint contains an
    object type when any of its members is `object` or a class/interface
    name (including `self`/`static`-like class references and nullable or
    union members such as `?Foo`, `int|Foo`). Members that are built-in
    scalars/pseudo types — `array`, `iterable`, `string`, `int`, `float`,
    `bool`, `false`, `true`, `null`, `void`, `mixed`, `callable`,
    `resource` — and the class name `Closure` (case-insensitive, treated like
    `callable`) are not object types. A variadic parameter is always treated
    as `array` (so `Foo ...$xs` is analyzed). Untyped parameters are analyzed.
- **D2** Run the access analysis for that name in the scope.

### Entry 2 — closure imports

For each closure with a `use (...)` list, for each imported variable with a
non-empty name:

- **D3** By-reference import (`use (&$v)`): report **U** (unused) on the
  imported variable when the closure body contains **no** reachable access
  to `$v` at all (the `use` list itself does not count as an access inside
  the closure).
- **D4** By-value import: run the access analysis for `$v` in the closure;
  if it found zero accesses, report **U** on the imported variable. (The
  analysis may itself report **W** findings, see below.)
- **D4a** (custos refinement, see Divergences) Object imports: before
  reporting **W** findings for a by-value import, take the inferred type of
  `$v` in the enclosing scope at the closure's position (type inference:
  declared and doc types of a parameter, assignments such as `new Foo`,
  call return types, narrowing). If that type contains an object type by the
  D1 criterion (`object` or any class/interface other than `Closure`; e.g.
  `ArrayObject`, `?Collection`, `array|Bag`), drop all **W** findings of that
  analysis: writes through an object handle (`$bag['k'] = 1` on an
  `ArrayAccess`, `$bag[] = 1`) are visible outside the closure. Unknown or
  `mixed` types are analyzed as before. **U** is unaffected.
- **D4b** (custos refinement, see Divergences) Includes: no **U** finding is
  reported for any import of a closure (D3 or D4) whose body contains an
  `include` / `include_once` / `require` / `require_once` expression
  (anywhere in the body, including arrow-function bodies, but not inside
  nested closures, named functions or classes, which have their own scope).
  This holds whatever `IGNORE_INCLUDES` says: an included file may read any
  variable of the closure's scope.

### Entry 3 — local assignments

For each plain assignment `$v = …` or by-reference assignment `$v = &…`
(compound assignments `+=`, `.=`, `??=`… are not entry points):

- **D4c** (custos refinement, see Divergences) Object locals: for an Entry 3
  local, drop all **W** findings when some plain assignment `$v = expr` of
  the scope gives `expr` an inferred type with an object member (D1
  criterion): `$t = new ArrayObject(); $t[] = 1;` writes into the object.
- **D4d** (custos refinement, see Divergences) A `global $v` declaration
  counts as a read: the variable is bound to the global one, so writes to it
  are never lost.
- **D5** The target is a simple variable with a non-empty literal name that
  is not one of `_GET`, `_POST`, `_SESSION`, `_REQUEST`, `_FILES`, `_COOKIE`,
  `_ENV`, `_SERVER`, `GLOBALS`, `HTTP_RAW_POST_DATA`.
- **D6** The assignment's direct parent is one of:
  - parentheses `( $v = … )`,
  - an array index `$a[$v = …]`,
  - an equality/identity comparison `==`, `!=`, `===`, `!==` (as either
    operand, e.g. `false === $v = f()`),
  - an expression statement (`$v = …;`),
  - another plain/by-ref assignment (the inner part of `$a = $v = …`).
  Any other parent (function-call argument, `if`/`while` condition, `return`,
  array literal element, logical operators, …) → not an entry point.
- **D7** The nearest enclosing scope exists (top-level code is skipped),
  `$v` is not one of that scope's parameters and not one of its `use`
  imports.
- **D8** Run the access analysis for `$v` in that scope. (A variable assigned
  N times is analyzed N times; findings must be de-duplicated by range.)

### Access analysis for name `v` in scope `S`

Collect all accesses to `$v` inside `S`'s body that are reachable by control
flow from `S`'s entry, in control-flow order (nested closures/functions are
separate scopes and not included, but a closure's `use ($v)` list is an
access in the enclosing scope). Parameter declarations are not accesses. An
assignment that is **not** directly an expression statement (its value is
consumed: in parentheses, an index, a comparison, a chained assignment…)
yields **two** accesses on the same target node: a write followed by a read.

If there are no accesses, the result is "zero accesses" (no finding here).
Otherwise keep counters `reads`, `writes`, a flag `ref` (initially false) and
a list of *targets*. For each access, with `X` the variable node and `P` its
parent, apply the first matching rule:

- **A1 container of an array access** (`$v[...]`, `$v[...][...]`): let `T`
  be the outermost array access in the chain and `Q` its parent.
  - `Q` is an assignment (plain, by-ref or compound) whose target is `T`:
    `writes++`; if `ref` is set or `Q` is a compound assignment (`.=`, `+=`,
    `??=`, …) → `reads++`; otherwise add `T` to targets.
  - `Q` is a `++`/`--` (prefix or postfix): add `T` to targets, `writes++`.
  - otherwise `reads++`.
- **A2 operand of a unary operator** (`P` is unary): if the operator is `++`
  or `--`: `writes++`, and if `ref` → `reads++` else add `P` (the whole
  `++$v` / `$v--` expression) to targets. Then, for any unary operator
  (`++`, `--`, `!`, `-`, casts, `@`, …), if `P` is **not** directly an
  expression statement → `reads++`.
- **A3 operand of a binary operator** (`P` is any binary expression, incl.
  `.`, `??`, comparisons, logical ops): `reads++`, then continue with A6/A7
  (which will count another read; harmless).
- **A4 compound assignment** (`P` is `$x op= …` and its target is a simple
  variable named `v`): `writes++`; if `ref` → `reads++` else add `X` to
  targets; if `P` is not directly an expression statement → `reads++`.
- **A5 plain/by-ref assignment** (`P` is `… = …`):
  - its target is a simple variable named `v`: `writes++`; if `ref` →
    `reads++`; if the assignment is by reference (`= &`) set `ref`. Then, if
    the whole analysis has **exactly two** accesses and both are on this
    same node (the write+read pair of an inline assignment whose result is
    consumed, and nothing else ever touches `$v`), report **U** on the
    assignment target variable and stop the analysis (it returns a non-zero
    access count; no W findings). Nothing is added to targets.
  - else its value is directly a simple variable named `v` → `reads++`.
  - else continue with A6/A7.
- **A6** `P` is a call argument list, a closure `use` list, `unset(...)`,
  `empty(...)`, `isset(...)`, or a `foreach` header (subject, key or value
  variable) → `reads++`.
- **A7** otherwise use the access's own nature: a write (e.g. destructuring
  target, `catch` variable, `static` declaration (`global`: D4d), `foreach` by
  reference handled above) → add `X` to targets and `writes++`; a read
  (return, interpolation, call on `$v->m()`, property fetch, …) → `reads++`.

After all accesses: when `reads == 0` and `writes > 0`, and not suppressed
(E4), and (`IGNORE_INCLUDES` is on **or** `S`'s body contains no
`include`/`include_once`/`require`/`require_once` anywhere, nested closures
included) → report **W** on every distinct target.

Notable consequences:

- A plain `$v = value;` statement counts as a write but is never itself a W
  target; so a local only assigned and never read produces no finding
  (`$tmp = 0;` alone is silent).
- `$p[] = x`, `$p['k'] = x`, `++$p`, `$p--`, `$p .= 'x'`, `$p += 1` are W
  targets when `$p` is never read.
- `$p['k'] .= 'x'` and `$p['k'] ??= 0` count as reads (no finding).
- After `$v = &$other`, later writes also count as reads (no finding).

## Exceptions (no report)

- **E1** By-reference parameters; parameters typed with a class/interface or
  `object` (but `Closure`-typed parameters are analyzed).
- **E2** Abstract methods and interface methods (no body).
- **E3** Any read of the variable anywhere reachable in the scope (including
  passing it to a call, `isset`, `empty`, a closure `use`, a comparison,
  concatenation, `??`).
- **E4** Rule-local suppression: if any W target's parent is an assignment
  that is directly an expression statement, and the statement's previous
  sibling (skipping whitespace and ordinary `//`/`/* */` comments) is a
  `/** … */` doc comment whose text contains both `@noinspection` and
  `OnlyWritesOnParameterInspection` (or custos' `@custos-ignore
  OnlyWritesOnParameter`), then **no** W finding of that analysis is
  reported at all (not only the annotated line).
- **E5** With `IGNORE_INCLUDES` off, scopes containing an include/require.
- **E6** Assignments in contexts not listed in D6 (e.g. `foo($v = 1)`,
  `if ($v = f())`, `while ($row = next($rs))`) are not entry points.
- **E5b** Write-only findings on a by-value closure import whose inferred
  type is an object (D4a).
- **E5c** Unused-import findings in a closure whose body includes/requires a
  file (D4b), for both values of `IGNORE_INCLUDES`.
- **E7** Unreachable code (after `return`, etc.) is not considered. Code is
  only unreachable when that is certain: a statement that ends abruptly but
  holds a `do … while` whose body has a `break` or `continue` does not cut
  off what follows (the loop condition is evaluated and the loop can exit:
  `do { if ($ok) { break; } return []; } while (false); return $v;`), and in
  a scope that declares a `goto` label every access counts as reachable.
  A `case` condition is evaluated before the statements of its case, so
  `case count($v): return 1;` reads `$v`.
- **E8** (custos refinement, see Divergences) Variables read by name: when
  the scope's body (nested functions, closures and classes excluded; arrow
  functions included) calls the global `compact()` with a string literal
  naming `$v` (directly or inside an array literal argument), or calls the
  global `get_defined_vars()`, the variable counts as read: no **W** and no
  **U** finding for it (parameters, imports — by value or by reference — and
  local assignments alike). A `compact()` argument that is not a string
  literal or an array of them (`compact($vars)`, `compact(...$names)`), and
  any variable variable (`$$name`, `${expr}`) in the body, count as reads of
  every variable. So does an `eval(...)` expression: the evaluated code runs
  in the scope and may read any of its variables.
- **E9** (custos refinement, see Divergences) Parameters read positionally:
  when the target PHP version is 7.0 or later and the scope's body (nested
  functions, closures, arrow functions and classes excluded) calls the
  global `func_get_args()` or `func_get_arg()`, every parameter of the scope
  counts as read. Before PHP 7.0 these functions return the values originally
  passed, so a write to the parameter is still lost and is reported.

## Report

- Ranges:
  - W, array element write: the outermost array access `T` as written, from
    the variable through the last `]` (e.g. `$items[]`, `$m[$k]`), not
    including ` = …`.
  - W, `++`/`--` on the variable: the whole unary expression (`++$n`, `$n--`).
  - W, compound assignment / other writes: the variable alone (`$s`).
  - U: the variable node only (`$v`), without a preceding `&` in a `use`
    list.
- Severity: info (rendered as an "unused" highlight; fixture markup
  `weak_warning`).
- Messages:
  - W: `Value is only written here and never read; the write is lost.`
  - U: `Variable is never used.`

## Fix

None.

## Options

| Option | Type | Default | Effect |
|---|---|---|---|
| `IGNORE_INCLUDES` | bool | `true` | When on, write-only findings are reported even if the scope contains `include`/`require` (which could read variables). When off, such scopes produce no W findings. U findings are unaffected. |

## PHP versions

- PHP 7.0+: `func_get_args()` / `func_get_arg()` return the current values
  of the parameters and therefore read them (E9).

## Examples

`IGNORE_INCLUDES = false`:

```php
<?php
class Report
{
    public function collect(array $rows, array &$sink, $tally, Closure $done, Iterator $it)
    {
        return function ($pending, array &$bucket) use (<weak_warning descr="Variable is never used.">$rows</weak_warning>, &<weak_warning descr="Variable is never used.">$sink</weak_warning>, $tally) {
            <weak_warning descr="Value is only written here and never read; the write is lost.">$pending['x']</weak_warning> = 1;
            $bucket[] = 2;
            <weak_warning descr="Value is only written here and never read; the write is lost.">$tally[]</weak_warning> = 3;
        };
    }

    public function counters($hits, $misses, $sum, &$total, Iterator $it)
    {
        <weak_warning descr="Value is only written here and never read; the write is lost.">$hits++</weak_warning>;
        <weak_warning descr="Value is only written here and never read; the write is lost.">--$misses</weak_warning>;
        <weak_warning descr="Value is only written here and never read; the write is lost.">$sum</weak_warning> *= 2;
        $total++;
        $it[] = 1;
    }

    public function inline_unused()
    {
        if (null !== (<weak_warning descr="Variable is never used.">$pos</weak_warning> = strpos('abc', 'b'))) {
            return true;
        }
        $map = [];
        $map[<weak_warning descr="Variable is never used.">$key</weak_warning> = 'k'] = 1;
        $map[$slot = 'z'] = 2;
        return $slot . count($map);
    }

    public function locals($flag)
    {
        $buffer = $other = [];
        if ($flag) {
            <weak_warning descr="Value is only written here and never read; the write is lost.">$buffer[]</weak_warning> = 1;
        }
        <weak_warning descr="Value is only written here and never read; the write is lost.">$other</weak_warning> .= 'x';

        $kept = '';
        $kept .= 'y';
        echo $kept ?? '';

        $alias = &$flag;
        $alias .= 'z';

        $unusedPlain = 5;
    }

    public function with_include($n, $path)
    {
        $n -= 1;
        require $path;
    }
}

function makeHandlers(ArrayObject $registry, array $plain, string $template)
{
    $store = new SplObjectStorage();
    $register = function ($name) use ($registry, $plain, $store) {
        $registry[$name] = true;          // object import, the write is visible
        $store[] = $name;                 // inferred SplObjectStorage
        <weak_warning descr="Value is only written here and never read; the write is lost.">$plain[$name]</weak_warning> = true;
    };
    $render = function () use ($template, $registry) {
        require __DIR__ . '/view.php';    // the view may read $template and $registry
    };
    $noop = function () use (<weak_warning descr="Variable is never used.">$template</weak_warning>) {
        $inner = function () { include 'x.php'; };   // nested closure: its include does not count
    };
}

function silenced(array $out)
{
    /** Buffers. @noinspection OnlyWritesOnParameterInspection */
    $out['a'][] = 1;
    $out['b'][] = 2;
}
```

## Divergences

- **compact() / get_defined_vars() (E8) — custos refinement, not upstream.**
  Upstream's flow analysis does not see names passed to `compact()`, so
  `if (!($page = $repo->find()))) { throw …; } return compact('page');`
  reports `$page` as never used, and a parameter written then exported via
  `compact()` as a lost write (found on real code). custos treats those
  names (and every variable, after `get_defined_vars()`, `compact()` of a
  computed list — phpBB's `extract(trigger_event(…, compact($vars)))` idiom —
  or a variable variable) as read.
- **eval(), func_get_args() (E8, E9) — custos refinement, not upstream.**
  `function () use ($id) { return eval('return $id + 1;'); }` uses its
  import, and `$label .= '!'; return func_get_args();` returns the modified
  parameter on PHP 7.0+; neither is reported.
- **Reachability after do-while and goto (E7) — custos refinement.** A read
  placed after `do { if ($ok) { break; } return []; } while (false);`, or
  after a label reached by `goto`, is executed, so the earlier writes are
  not lost. custos is deliberately cautious here: any `break`/`continue`
  inside such a do-while, and any label in the scope, keeps the following
  code reachable.
- **Object imports (D4a/E5b) — custos refinement, not upstream.** Upstream
  checks the native type only for parameters; a closure import has no type
  hint, so `use ($obj)` followed by `$obj['k'] = …` on an `ArrayObject` /
  `ArrayAccess` is reported as a lost write although the write goes through
  the shared object (found on real code). custos skips W findings when the
  import's inferred type is an object. In the upstream fixture the reported
  imports are untyped (`$log`) or array-typed, so conformance is unaffected.
- **Includes in closures (D4b/E5c) — custos refinement, not upstream.**
  Upstream reports `use ($config)` as unused even when the closure body
  does `include $path`, where the included file reads `$config` (found on
  real code). custos suppresses unused-import findings in such closures. No
  upstream fixture has an include inside a closure, so conformance is
  unaffected.
- A variable assigned in several statements triggers the analysis once per
  assignment upstream. custos emits each reported range once.
- The exact set of contexts in which the flow graph records the extra read
  of an inline assignment target (D-analysis "two accesses") is an upstream
  engine detail. custos records it for any assignment that is not directly
  an expression statement.
- A2 counts nothing for a non-`++/--` unary used as a bare statement
  (`!$v;`). Harmless; keep.
- **custos diverges — globals, by-reference aliases, object locals (D4c,
  D4d).** Found on WordPress and Nextcloud: upstream reports writes to a
  variable declared `global` (`global $mode; $mode = 'x';`, 95 findings),
  `$c['n']++` after `$c = &$info['count']` (the A1 `++/--` branch ignored the
  reference, unlike the assignment branch), and `$t[] = 1` on a local
  holding an `ArrayObject`. custos counts `global` declarations as reads,
  treats `++/--` on an element through a reference like an assignment, and
  drops W findings for locals assigned an object-typed value. No upstream
  fixture covers these shapes.
- **Anonymous class arguments (custos diverges).** The constructor
  arguments of `new class ($a, $b) { … }` are evaluated in the enclosing
  scope: a closure import passed only
  there (`use ($organisation)`, Matomo tests) is a use, not "never used".
- **Element writes on values of unknown type (custos diverges, extends
  D4c).** For a local variable, when every lost write is an element write
  (`$v['k'] = …`) and the variable's inferred type there is unknown or has
  an object member, nothing is reported: the value
  (`$spec = $e->getParam('inputSpec')`) may be an ArrayAccess object such
  as ArrayObject, whose writes are not lost (laminas-form listeners).
  Parameters keep the upstream behavior.
- **`mixed` values (custos diverges).** A local holding a value of type
  `mixed` (`$list = $prop->getValue($owner); $list[] = $item;`, the
  Doctrine `UnitOfWork` patch in Mautic) may hold an `ArrayAccess` object
  just like an unknown value, so `mixed` counts as an object member in D4c
  and in the element-write exception above.
