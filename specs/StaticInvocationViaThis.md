---
id: StaticInvocationViaThis
group: Code style
kind: semantic
needs: [names, types, hierarchy, index, stubs]
php: { min: "", max: "" }
---

# StaticInvocationViaThis

## Summary
Calling a static method through an object (`$this->make()` or `$obj->make()`)
hides the fact that no instance is involved. Use `self::`/`ClassName::` so the
call reads as what it is.

## Detection
Visit every method call expression.

- **D1** The call uses the object operator `->` (not `::`; the null-safe
  `?->` is not considered — see Divergences).
- **D2** The method name is a plain identifier (dynamic names such as
  `$obj->$name()` are skipped) and does **not** start with the lowercase
  prefix `static` (case-sensitive; e.g. `staticFactory()` is skipped).
- **D3** The receiver (left of `->`) is not itself a call: receivers that are
  function calls, method calls or static method calls (`foo()->m()`,
  `$a->b()->m()`, `A::b()->m()`) are skipped. Any other receiver (variable,
  property fetch, parenthesized `new`, array access, …) continues.
- **D4** The call resolves (via the receiver's inferred type and class
  hierarchy) to a method declared `static`.
- **D5** The resolved method is not excluded by the options (see E3/E4).
- Two variants:
  - **D6 (`$this` receiver)** The receiver is the variable `$this`, and the
    nearest enclosing function-like scope (function, method, closure, arrow
    function) is a **non-static method**. When `$this` appears inside a closure
    or arrow function, the nearest scope is that closure, so nothing is
    reported. Report variant A.
  - **D7 (other receiver)** The receiver is anything else, except a variable
    whose name matches a parameter of the nearest enclosing function-like
    scope, or a variable listed in that scope's `use (...)` clause. At top
    level (no scope) every variable receiver is eligible. Report variant B.

## Exceptions (no report)
- **E1** `::` calls (`self::m()`, `static::m()`, `Foo::m()`), calls on call
  results, dynamic method names, method names beginning with `static`.
- **E2** Non-static resolved method, or unresolved call.
- **E3** Option `EXCEPT_PHPUNIT_ASSERTIONS` (default true): the resolved
  method's declaring class FQN starts with `\PHPUnit` (case-sensitive) and,
  after replacing every `_` in the class FQN by `\`, starts with
  `\PHPUnit\Framework\`. Covers both `\PHPUnit\Framework\TestCase` and the
  legacy `\PHPUnit_Framework_TestCase` (plus anything else under that prefix).
  The same option also skips methods whose declaring class-like FQN starts
  with `\Symfony\Bundle\FrameworkBundle\Test\` (Symfony's PHPUnit
  assertion traits such as `WebTestAssertionsTrait`, `MailerAssertionsTrait`;
  custos diverges, see Divergences).
- **E4** Option `EXCEPT_ELOQUENT_MODELS` (default true): the resolved method is
  declared directly in class `\Illuminate\Database\Eloquent\Model` (calls on
  subclasses that resolve to the inherited method are skipped too; static
  methods declared in a subclass are not).
- **E5** `$this->m()` inside a static method, a plain function, a closure or
  arrow function.
- **E6** Receiver is a parameter or `use` variable of the nearest enclosing
  function-like scope (variant B only).

## Report
- Variant A (D6): range = the `$this` variable token only. Severity: warning.
  Message: `Static method {name}() called through $this; use self::{name}().`
  where `{name}` is the declared name of the resolved method.
- Variant B (D7): range = the whole method call expression, from the start of
  the receiver to the closing `)` of the argument list (no trailing `;`).
  Severity: warning. Message: `Static method {name}() called on an instance; call it with ::.`
  where `{name}` is the name as written at the call site.

## Fix
- **F1** Variant A only: replace `$this` with `self` and the `->` operator with
  `::`; everything else (method name, arguments, whitespace) is kept:
  `$this->build($x)` → `self::build($x)`.
- Variant B has no fix (the class name to use is not always obvious).

## Options
| Option | Type | Default | Effect |
|---|---|---|---|
| EXCEPT_PHPUNIT_ASSERTIONS | bool | true | Skip static methods declared in PHPUnit framework classes and Symfony's test assertion traits (E3). |
| EXCEPT_ELOQUENT_MODELS | bool | true | Skip static methods declared in Laravel's Eloquent base model (E4). |

## PHP versions
None.

## Examples

```php
<?php
class Palette {
    public static function shade($level) { return $level; }
    public static function staticTone() { return 0; }
    public function tint() { return 1; }

    public static function fromStatic() {
        return self::shade(1);
    }

    public function render() {
        $a = <warning descr="Static method shade() called through $this; use self::shade().">$this</warning>->shade(2);
        $b = $this->tint();
        $c = $this->staticTone();
        $d = self::shade(3);
        $e = function () { return $this->shade(4); };
    }
}

function paint(Palette $given) {
    $local = new Palette();
    $x = <warning descr="Static method shade() called on an instance; call it with ::.">$local->shade(5)</warning>;
    $y = <warning descr="Static method shade() called on an instance; call it with ::.">(new Palette())->shade(6)</warning>;
    $z = $given->shade(7);
    $w = makePalette()->shade(8);
    return function () use ($local) { return $local->shade(9); };
}
```

After fix:

```php
<?php
        $a = self::shade(2);
```
(variant B occurrences stay unchanged).

With `EXCEPT_PHPUNIT_ASSERTIONS = true`, both of these stay silent:

```php
<?php
namespace PHPUnit\Framework {
    class Assert { public static function assertTrue($v) {} }
    class CheckoutTest extends Assert {
        public function testTotal() { $this->assertTrue(true); }
    }
}
```

## Divergences
- **Symfony test assertions (custos diverges):** upstream exempts only
  `\PHPUnit\Framework\…` assertions, so Symfony's `$this->assertResponseIsSuccessful()`,
  `$this->assertSelectorTextContains()` or `$this->getMailerMessages()` —
  static methods of the FrameworkBundle `Test` assertion traits, documented
  and conventionally called through `$this` like PHPUnit's own — are
  reported by the hundred in functional test suites. custos extends E3 to
  `\Symfony\Bundle\FrameworkBundle\Test\`.
- Null-safe calls (`$obj?->make()`): upstream only checks the plain `->`
  token; treat `?->` as not matching (no report). No fixture covers it.
- E3 normalises underscores in the whole FQN string (class and method part),
  so a method name containing `_` is affected too; harmless because only the
  prefix is compared.
- **Late static binding in the fix (custos diverges).** Upstream rewrites
  `$this->m()` to `self::m()`. `$this->m()` resolves `m` on the runtime
  class, `self::m()` on the lexical class, so when a subclass overrides the
  static method the fix silently calls the parent's version. custos uses
  `static::` (and says so in the message), keeping `self::` only when `m`
  cannot be overridden: a private or final method, or a final class or an
  enum. EA case affected: `static-method-invocation-via-this.php` (listed
  divergence).
- **Abstract trait methods (custos diverges).** A trait may re-declare a
  method it needs as `abstract public static function assertTrue(...)`;
  `$this->assertTrue()` then resolves to that requirement instead of the
  real implementation (PHPUnit's, often not indexed), and the
  PHPUnit-assertion exception no longer applies. custos looks past such
  an abstract trait method for the implementation inherited from a parent
  (whose class then decides, e.g. the PHPUnit exception); when none is
  found and an ancestor does not resolve, the call is skipped. Inside the
  trait itself (no parent to consult) the abstract static requirement is
  used as before.
