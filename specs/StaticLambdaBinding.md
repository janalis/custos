---
id: StaticLambdaBinding
group: Probable bugs
kind: semantic
needs: [names, hierarchy, index]
php: { min: "5.4", max: "" }
---

# StaticLambdaBinding

## Summary

A `static` closure or arrow function has no bound object: using `$this` in it,
or calling an instance method of the parent class through `parent::`, fails at
runtime. Either drop `static` or stop relying on the object.

## Detection

Visit every anonymous function: classic closures and arrow functions.

- **D1** The configured PHP level is 5.4 or newer.
- **D2** The anonymous function's first token is the `static` keyword
  (`static function (...) {...}`, `static fn (...) => ...`).
- **D3** Collect candidates, in source order (pre-order, depth-first):
  - classic closure: every variable node and every method/static-method call
    node anywhere inside its body block (deep: nested closures, arrow
    functions, anonymous classes, string interpolation `"{$this->x}"` are all
    entered). Parameters and the `use (...)` list are not searched;
  - arrow function: the body expression itself (the expression after `=>`),
    followed by every variable and method-call node anywhere inside it.
- **D4** Report **every** candidate that matches, in that order (each
  report carries the same fix):
  - **D4a** a variable named exactly `this` whose *binding owner* is the
    static function being inspected. The binding owner is the nearest
    enclosing function (closure, arrow function, function or method),
    skipping non-static arrow functions: an arrow function inherits the
    object binding of the scope it is defined in, so `$this` in
    `static function () { return fn () => $this; }` belongs to the static
    closure and fails at runtime. A `$this` inside a nested classic closure
    or a nested `static fn` belongs to that nested function and does not
    match.
  - **D4b** a static-style call whose class part is the `parent`
    keyword, compared case-insensitively as PHP does (`Parent::`, `PARENT::`
    qualify; custos diverges) — `parent::name(...)` — that resolves to
    a method declared **non-static** and whose binding owner (as in D4a) is
    the inspected function. Unresolvable calls and calls resolving to static
    methods do not match.

## Exceptions (no report)

- **E1** Non-static closures / arrow functions.
- **E2** `$this` / `parent::` that belongs to a nested classic closure (or a
  nested `static fn`, reported for itself) inside the static one.
- **E3** `parent::m()` where `m` is static or cannot be resolved; `self::`,
  `static::`, `ClassName::` calls; `parent::CONST`, `parent::$prop`.
- **E4** PHP level below 5.4.

## Report

- Range:
  - D4a: the `$this` variable token only (5 characters, `$this`);
  - D4b: the whole call, from `parent` to the closing `)` of its arguments.
- Severity: error.
- Messages:
  - D4a: `Static closures have no $this; remove 'static' or avoid $this.`
  - D4b: `Calling an instance method of the parent class requires an object; this closure is static.`

## Fix

- **F1** (both cases) Delete the `static` keyword of the inspected anonymous
  function. Removing the single whitespace run that followed it is
  recommended (`static function` → `function`, `static fn` → `fn`); the
  comparison is whitespace-collapsed so either form matches.

## Options

None.

## PHP versions

Requires PHP ≥ 5.4 (static closures). Arrow functions naturally require 7.4
syntax but are inspected whenever they parse; upstream fixtures run at the
test-default level (below 7.1) and still expect arrow functions to be reported,
so do not gate arrow functions on 7.4.

## Examples

```php
<?php
class Base {
    public function render() {}
    public static function make() {}
}

class Widget extends Base {
    private $label = 'w';

    public function handlers() {
        return [
            function () { return $this->label; },
            static function () { return strtoupper(<error descr="Static closures have no $this; remove 'static' or avoid $this.">$this</error>->label); },
            static fn ($s) => $s . <error descr="Static closures have no $this; remove 'static' or avoid $this.">$this</error>->label,
            static function () { return fn () => 1; },
            static function () { return function () { return $this->label; }; },
            static function () { return fn () => <error descr="Static closures have no $this; remove 'static' or avoid $this.">$this</error>->label; },
            static function ($x) { <error descr="Calling an instance method of the parent class requires an object; this closure is static.">parent::render($x)</error>; },
            static fn () => <error descr="Calling an instance method of the parent class requires an object; this closure is static.">parent::render()</error>,
            static function () { return parent::make(); },
        ];
    }
}
```

```php
<?php
class Base {
    public function render() {}
    public static function make() {}
}

class Widget extends Base {
    private $label = 'w';

    public function handlers() {
        return [
            function () { return $this->label; },
            function () { return strtoupper($this->label); },
            fn ($s) => $s . $this->label,
            static function () { return fn () => 1; },
            static function () { return function () { return $this->label; }; },
            function () { return fn () => $this->label; },
            function ($x) { parent::render($x); },
            fn () => parent::render(),
            static function () { return parent::make(); },
        ];
    }
}
```

## Divergences

- **Keyword case (custos diverges).** Upstream recognises only the
  lower-case text `parent`, so `PARENT::render()` in a static closure is
  missed although PHP treats it as the same keyword and the call fails the
  same way. custos compares it case-insensitively (D4b).
- `parent::m()` inside a nested *non-static* closure within a static closure
  is still reported by upstream (no scope check for D4b), and the fix then
  strips `static` from the outer closure. custos applies the binding-owner
  check to D4b as well.
- **Arrow functions inherit the binding (custos diverges from upstream).**
  Upstream attributes `$this` to the nearest arrow function, so
  `static function () { return fn () => $this->x; }` is not reported
  although the arrow function has no object either and the call fails at
  runtime. custos skips non-static arrow functions when looking for the
  owner (D4a/D4b).
- **Every occurrence reported (custos diverges from upstream).** Upstream
  stops at the first offending `$this`/`parent::` of a function, leaving
  the others unmarked. custos reports each one; they share the single fix
  (removing `static`).
- **Attributed closures (custos fix).** D2 looks for the `static` keyword
  after the function's attributes: `#[Pure] static fn () => $this->x` is
  reported like the unattributed form (the attribute groups are not part of
  the check; the fix removes only `static`). An earlier custos version
  compared the function's first token, which is `#[` there, and missed it.
