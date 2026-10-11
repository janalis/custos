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
- **D4b** When the closure is created where `$this` exists, the search
  region must not contain `include`, `include_once`, `require`,
  `require_once` or `eval`: the loaded code runs in the closure's scope and
  may use the `$this` the closure was given (a view renderer including a
  template). `$this` exists in an instance method or a property hook, also
  through enclosing non-static closures and arrow functions. In a function,
  a static method, a static closure or file-level code the closure holds no
  object of its own, so loaded code does not block the report (an autoloader
  `spl_autoload_register(function ($class) { require …; })`).
- **D5** The search region must not contain a method call whose class part is
  a name (`parent` in any letter case, or a class name such as `Runner::m()`
  or `Base::m()`) that resolves to a **non-static** method: inside an instance
  context PHP runs such a call on the current object. A call that resolves to
  a static method, or that cannot be resolved, does not block.
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
- **D6** Usage collection. The closure's value is forwarded unchanged by
  parentheses, by either operand of `??`, by the branches of a ternary (for
  the short form `?:`, the left operand too) and by the result of a `match`
  arm; let *V* be the outermost such expression (the closure itself when
  there is none) and *C* the direct syntactic parent of *V*. Collect a list
  of *usage sites*:
  - **D6a** *C* is a call's argument list (the closure is passed directly as an
    argument): the only usage site is that argument list.
  - **D6b** *C* is a plain assignment (`=`, not compound) whose left side is a
    simple variable `$v`, written as a statement of its own. When the value
    of the assignment is used in turn (`return $v = function () {…};`,
    `$this->hook = $v = function () {…};`) the closure escapes: no report. If the closure has an enclosing function-like scope
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
    At file level (no enclosing function-like scope with a block body) the
    variable is a global. The whole file is scanned for the name with the
    same classification, and the closure counts as escaping when the file
    contains anything that can reach a global out of sight: `$GLOBALS`, a
    variable-variable, `include`/`require`, `eval`, or a call to `compact`,
    `extract` or `get_defined_vars`.
  - **D6c** *C* is an array element value: if the element is keyed
    (`key => <closure>`), the usage site is that key/value pair; if unkeyed,
    the usage site is the array literal itself.
  - **D6d** *C* is a `return` statement: no report (the caller may bind it).
  - **D6e** *C* is a method call with *V* as receiver
    (`(function () {…})->call($obj)`, `(…)->bindTo(…)`, `(…)::m()`): the
    usage site is that call.
  - **D6f** *C* is an expression statement (the value is discarded), a call
    invoking *V* (`(function () {…})()`) or an array element using *V* as its
    key: no usage sites.
  - Any other parent (the expression of an arrow function, `yield`, an
    operand of another operator, …): the closure escapes, no report.
- **D7** Every usage site must be *safe* (an empty list is trivially safe):
  - **S1** argument list of a method call (`->` or `::`):
    - method name `bind` (any case): safe only if the call has at least two
      arguments, the second is the constant `null` (any case), and the call
      resolves to the built-in `\Closure::bind`; otherwise unsafe;
    - any other name, `->` call: unsafe;
    - any other name, `::` call (including `self::`, `static::`, `parent::`,
      `Foo::`): safe when it resolves to a static method declared by PHP
      itself; *unverified* when it resolves to a user-defined static method,
      or when the class, the method or a dynamic method name is unknown;
      unsafe when it resolves to an instance method or to a magic method
      (`@method`, or `__callStatic()` without such a method).
  - **S2** argument list of a plain function call: safe when the callee
    resolves to a built-in function; *unverified* when it resolves to a
    user-defined function of the global namespace. Unresolved callees,
    dynamic callees and namespaced functions are unsafe.
  - **S3** argument list of anything else (`new Foo(...)`, …): unsafe.
  - **S4** method call with the variable (or the closure itself, D6e) as
    receiver: safe iff the method name
    is `bindTo` (any case), there is at least one argument, the first argument is
    `null`, and the call resolves to `\Closure::bindTo` (the variable is known to
    hold a closure, so resolution succeeds whenever the class `Closure` exists).
  - **S5** keyed array pair: unsafe, at file level too (see Divergences).
  - **S6** unkeyed array literal: unsafe.
- Report when D1–D5b hold, the closure does not escape (D6) and no usage
  site is unsafe. An *unverified* site keeps the report but withholds the
  fix (F2).

## Exceptions (no report)

- **E1** Already `static` closures / arrow functions.
- **E2** Closures with an empty body or a body of comments only.
- **E3** Arrow functions when `SUGGEST_FOR_SHORT_FUNCTIONS` is `false`.
- **E4** `$this` used anywhere in the search region (even in a nested function).
- **E5** `parent::method()` or `ClassName::method()` resolving to an
  instance method.
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
- **E11** `include`/`require`/`eval` in the search region of a closure
  created in an instance method or a property hook.
- **E12** Closures forwarded through `??`, `?:` or `match` into a `return`
  or a property; closures yielded, produced by an arrow function, or used as
  the receiver of a call other than `bindTo(null, …)`.
- **E13** A file-level variable holding the closure when the file binds it,
  passes it to an instance method, or reaches its globals indirectly
  (`$GLOBALS`, `$$name`, `include`, `eval`, `compact`, …).

## Report

- Range: the `function` keyword of a classic closure, or the `fn` keyword of an
  arrow function (also when attributes precede it; see Divergences).
- Severity: info (`weak_warning` in fixtures).
- Message: `Closure does not use $this; declare it static.`

## Fix

- **F1** Insert `static ` immediately before the `function` / `fn` keyword.
  Nothing else changes (`function () {...}` → `static function () {...}`,
  `fn() => 1` → `static fn() => 1`).
- **F2** No fix is offered when a usage site is *unverified* (S1/S2): the
  closure is handed to a user-defined function or static method, or to an
  unknown static target, whose body may bind it to an object
  (`test('adds', function () {…})` in a test framework,
  `Registry::macro('upper', function () {…})`). The report stays as a hint.

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
    'home' => function () { return 'index'; },
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
- **Class-name instance calls; loaded code — custos diverges from
  upstream** (D5, D4b). Upstream only looks at `parent::`. A closure calling
  `Runner::helper()` or `Base::work()` (instance methods of the enclosing
  class or an ancestor), or including a template that uses `$this`, is
  reported upstream and fails once static. custos stays silent; loaded code
  only blocks where the closure captures an object (D4b), and a file-level
  closure is assumed not to, although a template included from a method
  would give it one.
- **Forwarded values and receivers — custos diverges from upstream** (D6,
  D6e). Upstream looks at the direct parent only, so
  `$this->hook = $hook ?? function () {…};`, `return $c ?: function () {…};`
  and `(function () {…})->call($obj)` are reported although the closure is
  stored, returned or bound to an object. custos follows the value through
  `??`, `?:`, `match` and parentheses, applies S4 to a closure used as a
  receiver, and treats any context it does not recognise as an escape. The
  same holds for an assignment whose value is used (`return $v = …`).
- **File-level variables — custos diverges from upstream** (D6b). Upstream
  collects nothing for `$v = function () {…};` outside a function and
  reports it even when the file goes on to `Closure::bind($v, $obj)`. custos
  classifies every use of the name in the file and gives up when the file
  can reach its globals indirectly. A use from another file that includes
  this one stays out of sight.
- **User-defined consumers — custos diverges from upstream** (S1, S2, F2).
  Upstream treats every global function and every `::` call as a consumer
  that cannot bind the closure and offers the fix. Only PHP's own functions
  and static methods are known not to; for a user-defined or unknown
  consumer custos keeps the report and withholds the fix.
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
  Unknown classes stay reported, without a fix (F2).
