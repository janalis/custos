---
id: DynamicInvocationViaScopeResolution
group: Code style
kind: semantic
needs: [names, index, hierarchy, types]
php: { min: "", max: "" }
---

# DynamicInvocationViaScopeResolution

## Summary
Calling an instance (non-static) method through `::` (`self::run()`,
`static::run()`, `$obj::run()`) hides that the call is really an instance call
and breaks in static contexts. Such calls should use `->`.

## Detection
D1. Node: a method call that uses the `::` operator, whose method name is a
    plain, non-empty identifier (no dynamic names like `X::$m()` or
    `X::{'m'}()`).

D2. The call must resolve to a method declaration that:
    - belongs to a class or trait (not an interface; abstract classes are
      fine);
    - is **not** declared `static`;
    - is **not** declared `abstract`.
    Unresolvable calls are never reported.
    Resolution: class names via imports/namespace; `self`/`static` to the
    enclosing class; expression bases via their inferred type; the method is
    looked up in the class and its ancestors (and used traits). The "resolved
    class" below is the class/trait that declares the found method.

D3. **Keyword/name form.** Let *L* be the exact source text left of `::`.
    When *L* is `static`, `self`, or the short name of the resolved class,
    compared case-insensitively (`SELF`, `foo` match: PHP keywords and class
    names are case-insensitive; a qualified spelling like `\App\Foo` or
    `App\Foo` does not match):
    - the nearest enclosing function-like scope must be a **method** (not a
      closure/arrow function/plain function/top level), and
    - that method's name must differ (case-insensitive) from the called method
      name, and
    - the class containing that method must have a method with the called name
      (own, inherited, or from traits); if the enclosing method has no
      containing class this check passes.
    Then:
    - D3a. enclosing method is **non-static** → report, with fix F1;
    - D3b. enclosing method is **static** → report, without fix.

D4. **Expression form.** Otherwise (L does not match D3): report, with fix F2,
    when the left side of `::` is an expression that is **not** a class-name
    reference (so not a name, not `parent`, not a non-matching class name)
    and **not** a function or method call result. Typical triggers:
    `$obj::run()`, `$this::run()`, `$this->dep::run()`, `$list[0]::run()`.
    This form does not depend on the enclosing scope.

## Exceptions (no report)
E1. The target method is static, abstract, or declared in an interface.
E2. `parent::run()` (and `parent::__construct()`) — never reported.
E3. `OtherClass::run()` where `OtherClass` (as written) is not the short name of
    the class declaring `run` — e.g. calling an ancestor's implementation by
    its class name when the method is declared in yet another class, or any
    fully-qualified spelling.
E4. D3 form inside a method of the same name (`Base::save()` called from
    `save()`, the classic "call the overridden implementation" pattern).
E5. D3 form when the enclosing class has no method of that name.
E6. D3 form outside a method (top level, function, closure, arrow function).
E7. D4 form whose base is a call result: `make()::run()`, `$f->get()::run()`.
E8. Unresolvable target (unknown class/type, undefined method).

## Report
- Range: the whole method call expression, from the start of the left side of
  `::` through the closing `)` of the argument list.
- Severity: warning (fixtures tag it `warning`).
- Message (D3a): `Call '{name}' with '$this->' instead of '::'.`
- Message (D3b, D4): `Call '{name}' on an instance with '->' instead of '::'.`
  (`{name}` is the method name.)

## Fix
F1. (D3a) Replace the left side of `::` (the `static`/`self`/class-name token)
    with `$this`, and the `::` token with `->`. Arguments and whitespace stay
    as they are: `self::render($a, $b)` → `$this->render($a, $b)`.
F2. (D4) Replace only the `::` token with `->`; the base expression is kept:
    `$job::run()` → `$job->run()`.
No fix for D3b.

## Options
| Option | Type | Default | Effect |
|---|---|---|---|

## PHP versions
None.

## Examples
```php
<?php

class Engine
{
    public function start() {}
}

class Car extends Engine
{
    public function start() {}
    public function honk() {}
    public static function build() {}

    public static function factory()
    {
        <warning descr="Call 'honk' on an instance with '->' instead of '::'.">self::honk()</warning>;
        self::build();
    }

    public function drive($speed)
    {
        <warning descr="Call 'honk' with '$this->' instead of '::'.">static::honk()</warning>;
        <warning descr="Call 'start' with '$this->' instead of '::'.">Car::start()</warning>;
        <warning descr="Call 'honk' with '$this->' instead of '::'.">self::honk()</warning>;
        parent::start();
        $cb = function () { self::honk(); };
    }
}

class SportsCar extends Car
{
    public function start()
    {
        Car::start();
        Engine::start();
    }
}

$car = new Car();
<warning descr="Call 'honk' on an instance with '->' instead of '::'.">$car::honk()</warning>;
$car::build();
```

```php
<?php

class Engine
{
    public function start() {}
}

class Car extends Engine
{
    public function start() {}
    public function honk() {}
    public static function build() {}

    public static function factory()
    {
        self::honk();
        self::build();
    }

    public function drive($speed)
    {
        $this->honk();
        $this->start();
        $this->honk();
        parent::start();
        $cb = function () { self::honk(); };
    }
}

class SportsCar extends Car
{
    public function start()
    {
        Car::start();
        Engine::start();
    }
}

$car = new Car();
$car->honk();
$car::build();
```

## Divergences
- Upstream, D3 accepts `Foo::m()` written with the declaring class's short name
  even when `Foo` is unrelated to the enclosing class, as long as the enclosing
  class happens to have a method `m` too; the fix then rewrites it to
  `$this->m()`, which changes semantics. Recommendation: additionally require
  the enclosing class to be the resolved class or one of its descendants
  (keep upstream behaviour only if conformance demands it — fixtures don't
  exercise the unrelated case).
- **Letter case (custos diverges).** Upstream matches `static`, `self` and
  the class short name case-sensitively in D3; `SELF::m()` or `car::m()`
  then fall to D4, which skips plain names, so they are never reported.
  custos compares case-insensitively (D3).
- **Bound calls keep their target (custos diverges).** `self::m()` and
  `Foo::m()` always run that class's `m()`; `$this->m()` runs a subclass
  override. D3a's fix is offered only when both run the same method: the
  name resolves on the enclosing class to the called method itself (not
  `Ancestor::m()` while the class overrides `m`, which works like
  `parent::m()`), and no subclass can override it — the method is private
  or final, the class is final or an enum, or no indexed descendant
  declares it. Traits never get the fix (`self` is the using class).
  `static::m()` keeps the fix. Otherwise the call is reported without a
  fix (Dolibarr's printipp: `BasicIPP`'s constructor calls
  `self::_initTags()`, which `CupsPrintIPP` overrides). Listed divergence
  `dynamic-method-invocation-via-scope-resolution.php`.
