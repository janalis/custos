---
id: DeprecatedConstructorStyle
group: Language level migration
kind: syntax
needs: []
php: { min: "", max: "7.4" }
---

# DeprecatedConstructorStyle

## Summary

A method named after its class used to act as the constructor (PHP 4 style).
This form is deprecated since PHP 7.0 and no longer a constructor in PHP 8.0;
the method should be called `__construct`.

## Detection

The rule only runs when the target PHP version is below 8.0 (see E7).

Visit every method declaration `M` (class member `function name(...)`, with or
without body — abstract methods count).

- **D1** `M` is not `static`.
- **D2** `M` is declared directly inside a named class-like declaration `C`
  that is a `class` (abstract, final or plain). Traits and interfaces are
  excluded (see E2/E3).
- **D3** The name of `M` equals the name of `C`, compared
  **case-insensitively** as PHP compares class and method names (class
  `Ledger` with method `Ledger` or `ledger` matches).
- **D4** `C` does not itself declare (own members only, inherited methods do
  not count) a method named `__construct`.

All four hold → report.

The namespace of the file is **not** considered: a matching method in a class
declared inside a `namespace` is reported too (see Divergences).

## Exceptions (no report)

- **E1** The method is `static` (`static public function Ledger()`).
- **E2** The container is a trait.
- **E3** The container is an interface.
- **E4** The class also declares its own `__construct` (the name-matching
  method is then a legacy alias, kept for backward compatibility).
- **E5** (removed: names are compared case-insensitively, see D3.)
- **E6** Anonymous classes (they have no name to match).
- **E7** The target PHP version is 8.0 or later. There a method named after
  its class is an ordinary method that callers invoke by name
  (`class Query { function query() {} }`), not a constructor.

## Report

- Range: the method's name identifier token only (e.g. `Ledger` in
  `public function Ledger()`), not the modifiers, `function` keyword or
  parameter list.
- Severity: error (rendered as deprecated/struck-through in IDEs).
- Message: `Class '{class}' uses an old-style constructor; rename it to __construct.`

## Fix

- **F1** Replace the method's name identifier token with `__construct`.
  Nothing else changes: modifiers, parameters, body, doc comments and
  whitespace stay as they are. Call sites of the old name (e.g.
  `parent::Ledger()` in subclasses or `$this->Ledger()`) are not touched.
- **F2** No fix is offered when the method declares a return type: a
  constructor cannot have one, so the rename alone would not give valid PHP.
  The report is kept.

## Options

None.

## PHP versions

Reported only for targets below PHP 8.0 (E7). Upstream does not gate the
rule (see Divergences).

## Examples

```php
<?php
class Ledger
{
    private $rows;

    public function <error descr="Class 'Ledger' uses an old-style constructor; rename it to __construct.">Ledger</error>(array $rows)
    {
        $this->rows = $rows;
    }
}

abstract class Shape
{
    abstract protected function <error descr="Class 'Shape' uses an old-style constructor; rename it to __construct.">Shape</error>($sides);
}

class Counter
{
    public static function Counter() { return new self(); }
}

class Wallet
{
    public function __construct($amount) {}
    public function Wallet($amount) { $this->__construct($amount); }
}

class Gauge
{
    public function <error descr="Class 'Gauge' uses an old-style constructor; rename it to __construct.">gauge</error>() {}
}

trait Logger
{
    public function Logger() {}
}

interface Printer
{
    public function Printer();
}
```

```php
<?php
class Ledger
{
    private $rows;

    public function __construct(array $rows)
    {
        $this->rows = $rows;
    }
}

abstract class Shape
{
    abstract protected function __construct($sides);
}

class Counter
{
    public static function Counter() { return new self(); }
}

class Wallet
{
    public function __construct($amount) {}
    public function Wallet($amount) { $this->__construct($amount); }
}

class Gauge
{
    public function __construct() {}
}

trait Logger
{
    public function Logger() {}
}

interface Printer
{
    public function Printer();
}
```

## Divergences

- **PHP 8.0 and later (custos diverges)**: upstream reports at every language
  level. From PHP 8.0 a same-named method is no longer a constructor, so the
  report would be wrong and the rename would take a regular method away from
  its callers (and fail to compile when it declares a return type). custos
  stays silent there (E7); analyse with a target below 8.0 to find legacy
  constructors before migrating.
- **Return type (custos diverges)**: the rename is withheld for a method that
  declares a return type (F2).
- **Namespaced classes**: since PHP 5.3.3 a method named after its class is
  *not* a constructor when the class lives in a namespace; upstream still
  reports it and the fix would turn a plain method into a constructor
  (behavior change). No upstream fixture covers this. custos skips classes
  declared inside a non-global namespace.
- **Enums**: upstream only excludes traits and interfaces, so an `enum Foo`
  with a method `Foo()` would be reported, although enums cannot have
  constructors. custos skips enums.
- **`__construct` lookup case**: upstream's own-method lookup for
  `__construct` — whether `__CONSTRUCT` counts — is not covered by fixtures.
  custos matches it case-insensitively, as PHP does for method names.
- **Name match case (custos diverges)**: upstream compares class and method
  names case-sensitively, so `class Ledger { function ledger() {} }` is
  missed, although PHP treated that method as the old-style constructor.
  custos compares case-insensitively (D3).
