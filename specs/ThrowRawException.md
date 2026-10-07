---
id: ThrowRawException
group: Control flow
kind: semantic
needs: [names, index, hierarchy, stubs]
php: { min: "", max: "" }
---

# ThrowRawException

## Summary
Throwing the base `\Exception` class gives callers nothing specific to catch;
a more specific (SPL) exception class communicates intent better. Separately,
throwing a standard-shaped exception without any constructor argument loses
the diagnostic message.

## Detection
Visit every `throw` (statement form and, on PHP 8+, expression form). Let `A`
be the thrown operand.

- **D1** `A` is a `new` expression whose class is given by a *name* (a class
  reference: `Foo`, `\Foo`, `Ns\Foo`, an imported alias). Dynamic class
  expressions (`new $cls`, `new (expr)`), `new static`/`new self`/`new parent`
  if they do not resolve to a fixed name, and anonymous classes (`new class
  {}`) are ignored. Parentheses around `A` are looked through
  (`throw (new \Exception('x'));` is handled like the unwrapped form; the
  report ranges and fix stay on the inner `new` expression).
- **D2 (raw exception)** The class reference resolves (via namespace/`use`
  rules) to the fully qualified name `\Exception`, compared
  case-insensitively (`\exception` counts).
  → report M1 on the class reference. Arguments are irrelevant.
  Examples: `throw new Exception('x')` in the global namespace,
  `throw new \Exception()` anywhere, `use Exception as Base; … throw new Base(…)`.
  An unqualified `Exception` inside a namespace without a matching `use`
  resolves to `\That\Namespace\Exception` and is **not** D2.
- **D3 (no arguments)** Only when D2 did not apply and option
  `REPORT_MISSING_ARGUMENTS` is on: the `new` expression has zero arguments
  (`new Foo()` or `new Foo`), and the resolved FQN maps to **exactly one**
  class declaration (project code or built-in stubs), and that class:
  - has a constructor — its own or the nearest inherited one — declaring
    **exactly 3 parameters** (as the built-in `Exception`/`Error` family does:
    message, code, previous), and
  - does not preset the message: neither the class itself, nor any
    **user-defined** class in its parent chain, nor a trait used by such a
    class (or by the class itself) declares a property named `message`.
    Built-in parents (`\Exception`, `\Error`, SPL classes) are not
    consulted: their `$message` is only the empty default.
  → report M2 on the whole `new` expression.

## Exceptions (no report)
- **E1** `\Exception` thrown with or without arguments is only ever reported
  via D2 (never additionally via D3).
- **E2** Zero-argument `new` of a class whose effective constructor has a
  parameter count other than 3 (e.g. a custom `__construct()` with none).
- **E3** Zero-argument `new` of a class that declares its own `$message`
  property (e.g. `protected $message = 'default text';`), or whose user
  parent class (at any level) or used trait declares one.
- **E4** Class name not resolvable to a single declaration (unknown class, or
  duplicate declarations of the same FQN) → no D3 report.
- **E5** Thrown values that are not `new` expressions (`throw $e`,
  `throw make()`).
- **E6** `REPORT_MISSING_ARGUMENTS` off disables D3 entirely (D2 unaffected).

## Report
- D2 range: the class reference token(s) after `new` only, e.g. `Exception`
  or `\Exception` (including the leading `\`), not `new`, not the arguments.
- D3 range: the whole `new` expression, from `new` through the closing `)`
  of the (empty) argument list, or through the class name when written
  without parentheses. `throw` and `;` are excluded.
- Severity: info (weak warning) for both.
- Messages:
  - M1: `Throw a more specific exception class than \Exception.`
  - M2: `Pass a message when throwing this exception.`

## Fix
- **F1** (D2 only) Replace the class reference text with
  `<qualifier>RuntimeException`, where `<qualifier>` is the namespace part
  written in front of the reference's last segment. When there is no
  qualifier, or the qualifier is just the leading `\`, the result is
  `\RuntimeException`. In practice every D2 case yields `\RuntimeException`:
  - `new Exception('a')` → `new \RuntimeException('a')`
  - `new \Exception('a')` → `new \RuntimeException('a')`
  - alias `new Base('a')` (alias of `\Exception`) → `new \RuntimeException('a')`
  Arguments, `use` statements and everything else are untouched.
- D3 has no fix.

## Options
| Option | Type | Default | Effect |
|---|---|---|---|
| REPORT_MISSING_ARGUMENTS | bool | true | Enables D3 (exceptions created without any argument). |

## PHP versions
No gating. Throw-as-expression (PHP 8: `$x ?? throw new …`, `fn() => throw …`)
is handled the same way as the statement form.

## Examples

```php
<?php
class QuietFailure extends \LogicException {
    function __construct()
    {
        parent::__construct("quiet", 7);
    }
}
class PresetFailure extends \DomainException {
    protected $message = 'preset text';
}
class PlainFailure extends \UnderflowException {}

function guard($n) {
    if ($n < 0) {
        throw new <weak_warning descr="Throw a more specific exception class than \Exception.">\Exception</weak_warning>("negative $n");
    }
    if ($n === 0) {
        throw new <weak_warning descr="Throw a more specific exception class than \Exception.">Exception</weak_warning>();
    }
    if ($n === 1) {
        throw <weak_warning descr="Pass a message when throwing this exception.">new OutOfRangeException</weak_warning>;
    }
    if ($n === 2) {
        throw <weak_warning descr="Pass a message when throwing this exception.">new PlainFailure()</weak_warning>;
    }
    if ($n === 3) {
        throw new QuietFailure();
    }
    if ($n === 4) {
        throw new PresetFailure();
    }
    throw new LengthException('too long');
}
```

```php
<?php
class QuietFailure extends \LogicException {
    function __construct()
    {
        parent::__construct("quiet", 7);
    }
}
class PresetFailure extends \DomainException {
    protected $message = 'preset text';
}
class PlainFailure extends \UnderflowException {}

function guard($n) {
    if ($n < 0) {
        throw new \RuntimeException("negative $n");
    }
    if ($n === 0) {
        throw new \RuntimeException();
    }
    if ($n === 1) {
        throw new OutOfRangeException;
    }
    if ($n === 2) {
        throw new PlainFailure();
    }
    if ($n === 3) {
        throw new QuietFailure();
    }
    if ($n === 4) {
        throw new PresetFailure();
    }
    throw new LengthException('too long');
}
```

Inside a namespace:

```php
<?php
namespace Billing;

use Exception;

throw new <weak_warning descr="Throw a more specific exception class than \Exception.">Exception</weak_warning>('declined');
throw new Exception\Declined('card'); // resolves to \Exception\Declined (via the import), not \Exception
```

## Divergences
- **D3 — custos diverges from upstream.** Upstream looks only at the thrown
  class's own property list, so `class NotFound extends BaseError {}` where
  `BaseError` sets `protected $message = 'not found';` is reported as
  "pass a message" although the inherited preset already supplies one.
  custos also walks user-defined parent classes and their traits; built-in
  parents are skipped because their `$message` is the empty default.
- **FQN case (custos diverges).** Upstream compares the resolved name with
  `\Exception` exact-case, so `throw new \exception()` is missed although
  PHP class names are case-insensitive. custos compares case-insensitively
  (D2).
- **Parenthesised operand (custos diverges).** Upstream only inspects a
  `throw` whose operand is directly a `new` expression, so
  `throw (new \Exception('x'));` escapes both checks. The parentheses change
  nothing at runtime, so custos looks through them (D1).
