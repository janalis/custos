---
id: MissUsingParentKeyword
group: Code style
kind: semantic
needs: [names, index, hierarchy]
php: { min: "", max: "" }
---

# MissUsingParentKeyword

## Summary

`parent::m()` is meant for calling the overridden implementation of the
current method. When it calls some *other* inherited method that nobody in the
hierarchy overrides, it is equivalent to `$this->m()` / `self::m()`, and using
`parent::` there only misleads readers (and silently bypasses future
overrides).

## Detection

D1. Node: a method call `parent::name(...)` where the class part is written
    `parent` in any letter case (`Parent`, `PARENT`: PHP keywords are
    case-insensitive) and the method name is a plain identifier.
D2. The nearest enclosing function-like scope is a **method** (not a closure,
    arrow function, plain function or top-level code), that method is
    **not static**, and it belongs to a class that is **not a trait** (classes,
    abstract classes, anonymous classes qualify).
D3. The called name differs from the enclosing method's name
    (case-insensitive, as PHP compares method names).
D4. The enclosing class does **not** declare a method with the called name
    itself (own declarations only; PHP's case-insensitive method name
    matching).
D5. The method is not overridden below: unless the enclosing class is `final`,
    no class anywhere in the project that (directly or transitively) extends
    the enclosing class declares its own method with that name.
D6. The call resolves (through the parent class chain) to a method
    declaration. Unresolvable calls are not reported.
When D1–D6 hold, report.

## Exceptions (no report)

E1. `parent::sameName()` from inside `sameName()` — the legitimate use.
E2. The current class declares the method itself.
E3. Some descendant class declares the method (the `parent::` call may be a
    deliberate bypass of those overrides). Skipped when the current class is
    `final` (no descendants possible).
E4. Calls inside static methods, closures/arrow functions, functions, or
    trait methods.
E5. Unresolvable target method.
E6. Class part other than the `parent` keyword (an explicit
    class name).

## Report

- Range: the whole call expression, from `parent` through the closing `)`.
- Severity: info (fixtures tag it `weak_warning`).
- Message: `Call it as '{replacement}' instead of through 'parent::'.`

## Fix

F1. Replace the whole call with:
    - `self::{name}({args})` when the resolved method is declared `static`;
    - `$this->{name}({args})` otherwise;
    where `{name}` is the called name as written and `{args}` is the verbatim
    source text between the call's parentheses (empty when there are no
    arguments).

## Options

| Option | Type | Default | Effect |
|---|---|---|---|

## PHP versions

None.

## Examples

```php
<?php

class Widget
{
    public function paint($color) {}
    public static function registry() {}
    public function resize() {}
    public function layout() {}
}

class Button extends Widget
{
    public function layout()
    {
        parent::layout();
        <weak_warning descr="Call it as '$this->paint(&quot;red&quot;, 2)' instead of through 'parent::'.">parent::paint("red", 2)</weak_warning>;
        <weak_warning descr="Call it as 'self::registry()' instead of through 'parent::'.">parent::registry()</weak_warning>;
        parent::resize();
        parent::missing();
        $fn = function () { parent::paint('x'); };
    }

    public static function build()
    {
        parent::paint('blue');
    }
}

class IconButton extends Button
{
    public function resize() {}
}
```

```php
<?php

class Widget
{
    public function paint($color) {}
    public static function registry() {}
    public function resize() {}
    public function layout() {}
}

class Button extends Widget
{
    public function layout()
    {
        parent::layout();
        $this->paint("red", 2);
        self::registry();
        parent::resize();
        parent::missing();
        $fn = function () { parent::paint('x'); };
    }

    public static function build()
    {
        parent::paint('blue');
    }
}

class IconButton extends Button
{
    public function resize() {}
}
```

## Divergences

- D3 compares names case-sensitively upstream, so `parent::Layout()` inside
  `layout()` passes D3; it is always stopped by D4 instead (the enclosing
  method is an own declaration with that name). custos compares
  case-insensitively, which changes nothing observable.
- **Keyword case (custos diverges).** Upstream requires the class part to be
  written exactly `parent`, so `Parent::paint()` / `PARENT::paint()` are
  missed although PHP treats them identically. custos accepts any letter
  case (D1).
- Methods imported from traits used by the current class are not treated as
  own declarations in D4 upstream. Recommendation: treat trait-imported
  methods as own (no report), since `$this->m()` would then call the trait
  method rather than the parent one. Not covered by fixtures.
