---
id: StaticClosureCanBeUsed
group: Code style
kind: semantic
needs: [names, hierarchy, index, stubs]
php: { min: "5.4", max: "" }
---

# StaticClosureCanBeUsed

## Summary
A closure (or arrow function) that never touches the object it was created in
can be declared `static`. That prevents an implicit `$this` binding, makes the
scope explicit and avoids keeping the surrounding object alive.

## Detection
Visit every anonymous function: classic closures (`function (...) use (...) { ... }`)
and arrow functions (`fn (...) => expr`).

- **D1** The configured PHP level is 5.4 or newer.
- **D2** The function is not already static: no `static` keyword appears among
  the tokens that precede its parameter list (attributes such as `#[Foo]` may
  come before `static`; they do not matter).
- **D3** Candidate shape:
  - classic closure: its body block contains at least one statement. An empty
    statement (a lone `;`) counts as a statement; comments alone do not; `{}` does
    not qualify;
  - arrow function: qualifies only when option `SUGGEST_FOR_SHORT_FUNCTIONS` is
    `true` (default).
- **D4** Search region: for a classic closure, its body block only (parameters
  and `use` list are not searched); for an arrow function, the entire arrow
  function node. The search is deep: it enters nested closures, arrow
  functions, anonymous classes, string interpolations, etc. The region must not
  contain any variable named exactly `this`.
- **D5** The search region must not contain a method call whose class part is
  the name `parent` in any letter case (`parent::name(...)`, `PARENT::...`) that resolves (through the
  enclosing class's parent chain) to a **non-static** method. A `parent::` call
  that resolves to a static method, or that cannot be resolved, does not block.
  `parent::$prop` and `parent::CONST` are not considered.
- **D5b** The search region must not contain a method call whose class part is
  the name `self` or `static` (any case) and whose nearest enclosing class-like
  is the closure's own enclosing class-like (calls inside an anonymous class
  nested in the region are ignored), unless the call resolves (through the
  enclosing class and its ancestors) to a **static** method. A call that
  resolves to an instance method, does not resolve, or has a dynamic method
  name blocks the report: once the closure is static such a call has no
  object to run on. `self::$prop`, `self::CONST`, `new static` are not
  considered.
- **D6** Usage collection. Let *C* be the direct syntactic parent of the
  closure expression (parentheses are **not** looked through). Collect a list of
  *usage sites*:
  - **D6a** *C* is a call's argument list (the closure is passed directly as an
    argument): the only usage site is that argument list.
  - **D6b** *C* is a plain assignment (`=`, not compound) whose left side is a
    simple variable `$v`. If the closure has an enclosing function-like scope
    with a block body (function, method or closure), scan that scope's body
    deeply (including inside the closure itself and inside nested functions) for
    every variable named `v`, except the assignment's own target. For each such
    occurrence:
    - its direct parent is an argument list → add that argument list;
    - its direct parent is a method-call expression (normally `$v->m(...)` or
      `$v::m(...)`) → add that method call;
    - its direct parent is a function call whose callee is the variable
      itself (`$v(...)`) → ignored (invoking it cannot rebind it);
    - anything else (returned, re-assigned, stored, captured in a `use` list,
      passed through a property, …) → the closure escapes: no report.
    At top level (no enclosing function-like scope) nothing is collected.
  - **D6c** *C* is an array element value: if the element is keyed
    (`key => <closure>`), the usage site is that key/value pair; if unkeyed,
    the usage site is the array literal itself.
  - **D6d** The closure (after skipping any parentheses around it) is the
    expression of a `return` statement: no report (the caller may bind it).
  - Any other parent (parenthesized expression not returned, property
    assignment, ternary, …): no usage sites.
- **D7** Every usage site must be *safe* (an empty list is trivially safe):
  - **S1** argument list of a method call (`->` or `::`):
    - method name `bind` (any case): safe only if the call has at least two
      arguments, the second is the constant `null` (any case), and the call
      resolves to the built-in `\Closure::bind`; otherwise unsafe;
    - any other name: safe iff the call uses `::` (static call, including
      `self::`, `static::`, `parent::`, `Foo::`).
  - **S2** argument list of a plain function call: safe iff the callee
    resolves to a function declared in the global namespace (built-in or a
    user-defined global function). Unresolved callees, dynamic callees and
    namespaced functions are unsafe.
  - **S3** argument list of anything else (`new Foo(...)`, …): unsafe.
  - **S4** method call with the variable as receiver: safe iff the method name
    is `bindTo` (any case), there is at least one argument, the first argument is
    `null`, and the call resolves to `\Closure::bindTo` (the variable is known to
    hold a closure, so resolution succeeds whenever the class `Closure` exists).
  - **S5** keyed array pair: safe iff it is not inside any function-like scope
    (i.e. top-level code).
  - **S6** unkeyed array literal: unsafe.
- Report when D1–D5b hold, the closure does not escape (D6b, D6d) and D7
  holds for all usage sites.

## Exceptions (no report)
- **E1** Already `static` closures / arrow functions.
- **E2** Closures with an empty body or a body of comments only.
- **E3** Arrow functions when `SUGGEST_FOR_SHORT_FUNCTIONS` is `false`.
- **E4** `$this` used anywhere in the search region (even in a nested function).
- **E5** `parent::method()` resolving to an instance method.
- **E6** Passed to a non-static method call (`$obj->run($closure)`), to
  `bind`/`bindTo` with a non-null scope/object, to a namespaced or unknown
  function, to a constructor.
- **E7** Used as a keyed array value inside a function/method/closure, or as an
  unkeyed array value anywhere.
- **E8** PHP level below 5.4.
- **E9** Returned closures (`return function () {...};`,
  `return (fn () => 1);`); closures held in a variable that is returned,
  stored, re-assigned, captured or otherwise used outside the recognised call
  shapes.
- **E10** `self::m()` / `static::m()` that is not known to call a static
  method.

## Report
- Range: the `function` keyword of a classic closure, or the `fn` keyword of an
  arrow function (also when attributes precede it; see Divergences).
- Severity: info (`weak_warning` in fixtures).
- Message: `Closure does not use $this; declare it static.`

## Fix
- **F1** Insert `static ` immediately before the `function` / `fn` keyword.
  Nothing else changes (`function () {...}` → `static function () {...}`,
  `fn() => 1` → `static fn() => 1`).

## Options
| Option | Type | Default | Effect |
|---|---|---|---|
| SUGGEST_FOR_SHORT_FUNCTIONS | bool | true | When false, arrow functions are never reported. |

## PHP versions
- Nothing reported below PHP 5.4.
- Arrow functions exist from PHP 7.4 (parser concern only).

## Examples

```php
<?php
abstract class Base {
    public function work() {}
    public static function tool() {}
}

final class Runner extends Base {
    private $limit = 3;

    public function cases() {
        $squares = array_map(<weak_warning descr="Closure does not use $this; declare it static.">function</weak_warning> ($n) { return $n * $n; }, [1, 2]);
        $capped  = array_map(function ($n) { return min($n, $this->limit); }, [4]);
        $helper  = array_map(<weak_warning descr="Closure does not use $this; declare it static.">function</weak_warning> ($n) { return parent::tool(); }, [5]);
        $inst    = array_map(function ($n) { return parent::work(); }, [6]);
        $noop    = array_map(<weak_warning descr="Closure does not use $this; declare it static.">function</weak_warning> () { ; }, []);
        $empty   = array_map(function () {}, []);
        $twice   = array_map(<weak_warning descr="Closure does not use $this; declare it static.">fn</weak_warning> ($n) => $n * 2, [7]);
        $self    = array_map(fn ($n) => $n + $this->limit, [8]);
    }

    public function rebinding() {
        $detached = <weak_warning descr="Closure does not use $this; declare it static.">function</weak_warning> () { return 42; };
        $copy = Closure::bind($detached, null, self::class);
        $other = $detached->bindTo(null);

        $attached = function () { return 7; };
        $attached->bindTo($this);
    }

    public function dispatching($queue) {
        $job = <weak_warning descr="Closure does not use $this; declare it static.">function</weak_warning> () { return 'ok'; };
        Registry::push($job);

        $task = function () { return 'later'; };
        $queue->push($task);

        new Worker(function () { return 1; });
    }

    public function scopedMap() {
        return ['a' => function () { return 1; }];
    }
}

$routes = [
    'home' => <weak_warning descr="Closure does not use $this; declare it static.">function</weak_warning> () { return 'index'; },
    'about' => static function () { return 'about'; },
];
$plain = [function () { return 0; }];
```

After fix (only changed lines shown):

```php
<?php
        $squares = array_map(static function ($n) { return $n * $n; }, [1, 2]);
        $helper  = array_map(static function ($n) { return parent::tool(); }, [5]);
        $noop    = array_map(static function () { ; }, []);
        $twice   = array_map(static fn ($n) => $n * 2, [7]);
        $detached = static function () { return 42; };
        $job = static function () { return 'ok'; };
    'home' => static function () { return 'index'; },
```

## Divergences
- **Escaping variables — custos diverges from upstream** (D6b). When the
  variable holding the closure is used in an unrecognised way (returned,
  stored, re-assigned, captured), upstream throws away every collected usage
  site and ends up reporting the closure; adding `static` can then break code
  that binds the closure elsewhere. custos treats such a use as unsafe and
  stays silent. Direct invocation `$v(...)` is the one exception, since
  calling a closure never rebinds it.
- **Closures stored outside a plain variable — custos diverges from
  upstream** (D6b). Upstream collects no usage site for
  `$this->hook = function () {…};` (or an array element, a static property,
  a compound `??=` assignment) and reports the closure, although whoever
  reads that storage later may bind it, exactly like a variable that is
  stored elsewhere. custos treats such an assignment as an escape and stays
  silent.
- **Returned closures — custos diverges from upstream** (D6d). Upstream
  reports `return function () {...};` because it finds no usage site, but
  the caller receives the closure and may `bindTo()` an object; a static
  closure makes that fail. custos does not report returned closures.
- **Attributed closures.** Upstream highlights the first child of the function,
  which is the attribute group for `#[A] function () {}`, and inserts `static`
  before it, producing invalid code (`static #[A] function`). Recommendation:
  always highlight the `function`/`fn` keyword and insert `static ` right before
  it (`#[A] static function`). No fixture covers this.
- **`self::`/`static::` instance calls — custos diverges from upstream**
  (D5b). Upstream only inspects `parent::` calls, so a closure calling
  `self::instanceMethod()` is reported, and once it is static the call fails
  for lack of an object. custos blocks the report unless the call is known
  to target a static method.
- Named arguments (`foo(callback: function () {})`): unclear whether the
  argument-list parent check matches upstream; treat a named argument as a
  normal argument of that list.
- **Letter case (custos diverges).** Upstream matches `parent` (D5) and the
  method names `bind`/`bindTo` (S1/S4) case-sensitively. A closure calling
  `PARENT::instanceMethod()` is then reported, and the static closure fails
  at runtime; `Closure::BIND($c, $obj)` is treated as an ordinary static call
  (safe) and the closure reported although it gets bound to an object.
  PHP keywords and method names are case-insensitive; custos compares them
  case-insensitively.
- **Keyed array values at file level; facades (custos diverges).** A
  closure stored as a keyed array value is not reported at file level
  either: such arrays are often returned to the including code, which may
  bind the closure (Grav's `updates/*.php` return `['postflight' =>
  function () {…}]`, run with `$closure->call($this)`; a static closure
  would only warn and never run). A static call reaches the closure's
  consumer as a static method only when the method is declared static: a
  magic `@method static` or a class with `__callStatic()` and no such
  method (Laravel/October facades: `Cache::extend('x', fn …)` binds the
  closure to the cache manager) counts as an unknown target (E6).
  Unknown classes keep upstream's behaviour.
