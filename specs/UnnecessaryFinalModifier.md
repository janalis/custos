---
id: UnnecessaryFinalModifier
group: Code style
kind: syntax
needs: []
php: { min: "", max: "" }
---

# UnnecessaryFinalModifier

## Summary

`final` on a method is redundant when the class itself is `final` (nothing can
extend it) or when the method is `private` (it cannot be overridden anyway,
constructors and other magic methods aside).

## Detection

Visit every method declaration in a class-like body.

- **D1** The method carries the `final` modifier.
- **D2** The method is a regular method, not a PHP 8.4 property hook
  (`final get => ...` / `final set { ... }` inside a property declaration are
  ignored).
- **D3** Either:
  - **D3a** the directly containing class is declared with the `final`
    keyword (any method, any visibility, magic methods included); or
  - **D3b** the method is `private` **and** its name does not start with `__`
    (two underscores, case-sensitive check on the literal prefix).

## Exceptions (no report)

- **E1** `final public|protected` methods in a non-final class.
- **E2** `final private` magic-named methods (`__construct`, `__clone`,
  `__destruct`, any `__name`) in a non-final class.
- **E3** Property hooks.
- **E4** Methods without `final`.

## Report

- Range: the method name identifier.
- Severity: info.
- Message: `Redundant final: the method cannot be overridden anyway.`

## Fix

- **F1** Remove the `final` keyword together with the whitespace that follows
  it; keep all other modifiers in their original order:
  `final public function a()` → `public function a()`,
  `private final static function b()` → `private static function b()`.
  Attributes and doc comments before the method are untouched.

## Options

None.

## PHP versions

None (property hooks exist from PHP 8.4; the EA fixture for them is run at 8.4
and expects no report).

## Examples

```php
<?php
class Engine
{
    final protected function start() {}
    final private function <weak_warning descr="Redundant final: the method cannot be overridden anyway.">ignite</weak_warning>() {}
    private static final function <weak_warning descr="Redundant final: the method cannot be overridden anyway.">spark</weak_warning>() {}
    final private function __clone() {}
}

final class Turbine
{
    final public function <weak_warning descr="Redundant final: the method cannot be overridden anyway.">spin</weak_warning>() {}
    final protected function <weak_warning descr="Redundant final: the method cannot be overridden anyway.">__wakeup</weak_warning>() {}
    public function idle() {}
}

class Gauge
{
    public int $level {
        final get => 1;
    }
}
```

```php
<?php
class Engine
{
    final protected function start() {}
    private function ignite() {}
    private static function spark() {}
    final private function __clone() {}
}

final class Turbine
{
    public function spin() {}
    protected function __wakeup() {}
    public function idle() {}
}
```

## Divergences

- Upstream rebuilds the whole modifier list from its internal representation
  (canonical order, and likely adds an explicit `public` when none was
  written), and replaces the method's first child, which drops attributes on
  attributed methods. custos removes only the `final` token (F1). Results are
  identical on EA fixtures (comparison is whitespace-normalised); the
  canonical reordering of e.g. `static final public` is not reproduced —
  acceptable since no fixture covers it.
- Enums (implicitly final) are not treated as final classes; only the explicit
  `final` keyword counts. Same for anonymous classes.
