---
id: MagicMethodsValidity
group: Probable bugs
kind: semantic
needs: [names, hierarchy, types]
php: { min: "", max: "" }
---

# MagicMethodsValidity

## Summary
PHP's magic methods (`__get`, `__toString`, `__clone`, …) have fixed contracts:
static-ness, visibility, number of parameters, return value. Violations are
fatal errors, silently ignored methods or subtle bugs. This rule validates
each magic method declared in a class-like, flags `__`-prefixed methods that
are not magic, and catches magic names typed with a single underscore.

## Detection
Visit every method declaration that:
- belongs to a class-like (class, trait, interface, enum, anonymous class);
- has a name identifier;
- has a name starting with `_`;
- is **not abstract** (abstract methods and all interface methods are
  skipped).

Then dispatch on the method name, compared **case-insensitively** (PHP
method names are case-insensitive: `__ToString` is `__toString`; see
Divergences). Messages show the name as written (`{m}`). Unless stated otherwise every check below reports on the method
**name identifier**. The checks are independent: one method may get several
reports (possibly on the same range).

Building-block checks:
- **C-static** "cannot be static": the method is declared `static`.
- **C-mstatic** "must be static": the method is not `static`.
- **C-public** "must be public": the method is `protected` or `private`
  (no modifier = public).
- **C-noargs** "takes no parameters": the method declares ≥ 1 parameter
  (optional or variadic included).
- **C-argc(N)** "exactly N parameters": parameter count ≠ N.
- **C-byref** "no by-reference parameters": at least one parameter is
  declared `&$p` (one report per method, not per parameter).
- **C-noreturn** "cannot return a value": every `return <expr>;` statement
  inside the method body, at any depth, whose nearest enclosing function-like
  is the method itself (returns inside closures/arrow functions/anonymous
  class methods are ignored). Bare `return;` is fine. Reports on the **whole
  return statement**, from `return` through the terminating `;`.
- **C-pair(M)** "needs companion M": the class-like has no method `M` — own,
  inherited from resolvable parents, or from used traits. An unresolvable
  parent contributes nothing (so the report still happens).
- **C-parent** "should call parent": all of:
  - the method has no `#[\Override]` attribute (written `#[Override]` in the
    global namespace, `#[\Override]`, or imported alias resolving to
    `\Override`);
  - the direct parent class resolves, and looking up the same method name on
    it (including what that parent inherits) finds a method that is neither
    abstract nor private;
  - nowhere inside the method (any depth, closures included) is there a
    method call or static method call (`->name(...)`, `::name(...)`,
    `parent::name(...)`) whose name equals the method's name
    (case-insensitive).
  The message names the class that *declares* the found parent method (its
  short name) and the method name.
- **C-returns(T)** "must return T" (T a set of allowed types):
  - If the method declares a return type: resolve it, drop unknown parts,
    normalise each part (`self` → the class FQN, scalar names lowercase,
    classes fully qualified). If the result is non-empty and contains a type
    not *accepted* by T → report on the name identifier. A type is accepted
    when it is in T, or when it is a class type that extends/implements a
    class type of T in the index (a subclass instance satisfies a class
    contract, e.g. `__set_state` returning a subclass).
  - Otherwise, if the method contains no `return` statement at all (any depth,
    closures included) → report on the name identifier (resolved = nothing),
    unless its body always ends in `throw`/`exit` (custos, see Divergences).
  - Otherwise, for each `return` statement at any depth:
    - with a value (parentheses stripped): infer its type; drop unknown parts.
      If the result is empty (inference failed, e.g. an untyped property) or
      every part is accepted by T → fine. Else, if the return belongs to a nested
      closure/arrow function/anonymous-class method → fine. Else report on
      the **whole return statement**.
    - without a value (`return;`) → report on the **whole return statement**
      (resolved = nothing); this happens even for a bare return in a nested
      closure.
- **C-version(V)**: the configured PHP level is below V → report "only
  available from V, unused before".

Dispatch:
- **D1** `__construct`: C-static, C-noreturn, and C-parent unless the file is
  a test context (D16).
- **D2** `__destruct`, `__clone`: C-static, C-noreturn, C-noargs, C-parent.
- **D3** `__get`, `__isset`, `__unset`: C-argc(1), C-static, C-public,
  C-byref, C-pair(`__set`).
- **D4** `__set`: C-argc(2), C-static, C-public, C-byref, C-pair(`__isset`),
  C-pair(`__get`).
- **D5** `__call`: C-argc(2), C-static, C-public, C-byref.
- **D6** `__callStatic`: C-argc(2), C-mstatic, C-public, C-byref.
- **D7** `__toString`: C-static, C-noargs, C-public, C-returns({`string`}).
- **D8** `__debugInfo`: C-static, C-noargs, C-public,
  C-returns({`array`, `null`}), C-version(5.6).
- **D9** `__set_state`: C-argc(1), C-mstatic, C-public,
  C-returns({the containing class FQN, `static`}); `static` is accepted but
  not shown in the message.
- **D10** `__invoke`: C-static, C-public.
- **D11** `__wakeup`: C-static, C-noargs, C-noreturn. (Visibility is **not**
  checked: a private `__wakeup` is fine.)
- **D12** `__unserialize`: C-static, C-public, C-argc(1), C-noreturn.
- **D13** `__sleep`, `__serialize`: C-static, C-public, C-noargs,
  C-returns({`array`}).
- **D14** `__autoload` (as a method): C-argc(1), C-noreturn, and always a
  report on the name identifier saying it is deprecated in favour of
  `spl_autoload_register()` (PHP 7.2). Reported at every PHP level.
- **D15** Any other name:
  - starting with `__` and not in the known non-magic list → report on the
    name identifier: only magic methods should use the `__` prefix. Known
    non-magic names (never reported): `__`, `__inject`, `__prepare`,
    `__toArray`, and SoapClient's `__doRequest`, `__getCookies`,
    `__getFunctions`, `__getLastRequest`, `__getLastRequestHeaders`,
    `__getLastResponse`, `__getLastResponseHeaders`, `__getTypes`,
    `__setCookie`, `__setLocation`, `__setSoapHeaders`, `__soapCall`
    (compared case-insensitively).
  - otherwise (single leading `_`, or a known non-magic name): if the name is
    exactly one of `_construct`, `_destruct`, `_call`, `_callStatic`, `_get`,
    `_set`, `_isset`, `_unset`, `_sleep`, `_wakeup`, `_toString`, `_invoke`,
    `_set_state`, `_clone`, `_debugInfo` (case-insensitive) → report on the
    name identifier
    (missing underscore), with fix F1. Other single-underscore names
    (`_serialize`, `_unserialize`, `_autoload`, `_foo`) are not reported.
- **D16** Test context (only affects D1): the file path ends with `Test.php`,
  `Spec.php` or `.phpt`, or contains `/Fixtures/`; or the containing class
  FQN ends with `Test` or contains `\Tests\` or `\Test\`.

## Exceptions (no report)
- **E1** Abstract methods and interface methods.
- **E2** Functions outside class-likes (a global `function __autoload()` is
  not visited).
- **E3** C-noreturn: returns of nested closures / arrow functions, and bare
  `return;`.
- **E4** C-returns: return values whose type cannot be inferred
  (`return $this->untyped;`), and incompatible returns located in nested
  closures.
- **E5** C-parent: `#[\Override]`, abstract or private parent method,
  unresolvable parent, any same-named method call in the body, and
  `__construct` in test context.
- **E6** Known non-magic `__` names (D15 list).

## Report
- Range: the method name identifier, except C-noreturn and the per-return
  branch of C-returns, which highlight the full `return …;` statement.
- Severity: error for every report (fixtures tag all of them `error`,
  including the deprecation report of D14 and, by the same rule, the version
  report of C-version).
- Messages (our wording; `{m}` = method name):
  - C-static: `{m} must not be static.`
  - C-mstatic: `{m} must be declared static.`
  - C-public: `{m} must be declared public.`
  - C-noargs: `{m} must not declare parameters.`
  - C-argc: `{m} must declare exactly {n} parameter(s).`
  - C-byref: `{m} must not take parameters by reference.`
  - C-noreturn: `{m} must not return a value.`
  - C-pair: `{m} needs a companion {companion} method.`
  - C-parent: `{m} does not call {ParentClass}::{m}().`
  - C-returns: `{m} must return {allowed}; got '{resolved}'.` where
    `{allowed}` joins T without `static` by `|` (e.g. `array|null`,
    `\Shop\Cart`) and `{resolved}` joins the offending normalised types
    (empty when nothing was returned).
  - C-version: `{m} only exists from PHP {v}; it is never called here.`
  - D14: `__autoload is deprecated since PHP 7.2; use spl_autoload_register().`
  - D15 prefix: `The '__' prefix is reserved for magic methods.`
  - D15 underscore: `'{m}' is not magic; did you mean '_{m}'?`

## Fix
- **F1** (D15 missing underscore only): prepend one `_` to the method name
  identifier (`_clone` → `__clone`). Nothing else changes.
- All other reports: no fix.

## Options
None.

## PHP versions
- C-version for `__debugInfo` fires only when the configured level is below
  5.6. EA cases without explicit level run at PhpStorm's test default, which
  is ≥ 5.6 and < 7.1: a valid `__debugInfo` is **not** reported there.
- D14 is reported regardless of level.

## Examples

```php
<?php
class Base
{
    public function __construct($id = 0) {}
    public function __clone() {}
}

class Point extends Base
{
    public function __construct($id = 0) { parent::__construct($id); }
    static public function <error descr="__clone must not be static.">__clone</error>() { parent::__clone(); }
    public function <error descr="__get must declare exactly 1 parameter(s)."><error descr="__get needs a companion __set method.">__get</error></error>() {}
    protected function <error descr="__call must be declared public.">__call</error>($verb, $args) {}
    public function <error descr="__callStatic must be declared static.">__callStatic</error>($verb, $args) {}
    public function <error descr="The '__' prefix is reserved for magic methods.">__fetchAll</error>() {}
    public function __soapCall() {}
}

class Line extends Base
{
    public function <error descr="__construct does not call Base::__construct().">__construct</error>($id = 0) {}
    #[\Override]
    public function __clone() {}
    public function __toString()
    {
        if (rand(0, 1)) {
            return $this->label;
        }
        <error descr="__toString must return string; got 'int'.">return 42;</error>
    }
    public function __wakeup()
    {
        <error descr="__wakeup must not return a value.">return true;</error>
    }
    public function <error descr="__sleep must return array; got ''.">__sleep</error>() {}
    public function <error descr="__debugInfo must return array|null; got 'string'.">__debugInfo</error>(): string { return ''; }
    public function <error descr="__isset must not take parameters by reference.">__isset</error>(&$key) {}
    public function <error descr="__set needs a companion __get method.">__set</error>($key, $value) {}
    public function __unset($key) {}
}

class Shape
{
    public static function __set_state($props) { return new static(); }
    public function <error descr="'_invoke' is not magic; did you mean '__invoke'?">_invoke</error>() {}
    public function <error descr="__autoload is deprecated since PHP 7.2; use spl_autoload_register().">__autoload</error>($cls) {}
}

interface Printable { public function __toString(); }
```

```php
<?php
class Base
{
    public function __construct($id = 0) {}
    public function __clone() {}
}

class Point extends Base
{
    public function __construct($id = 0) { parent::__construct($id); }
    static public function __clone() { parent::__clone(); }
    public function __get() {}
    protected function __call($verb, $args) {}
    public function __callStatic($verb, $args) {}
    public function __fetchAll() {}
    public function __soapCall() {}
}

class Line extends Base
{
    public function __construct($id = 0) {}
    #[\Override]
    public function __clone() {}
    public function __toString()
    {
        if (rand(0, 1)) {
            return $this->label;
        }
        return 42;
    }
    public function __wakeup()
    {
        return true;
    }
    public function __sleep() {}
    public function __debugInfo(): string { return ''; }
    public function __isset(&$key) {}
    public function __set($key, $value) {}
    public function __unset($key) {}
}

class Shape
{
    public static function __set_state($props) { return new static(); }
    public function __invoke() {}
    public function __autoload($cls) {}
}

interface Printable { public function __toString(); }
```

## Divergences
- Case-insensitive names (custos diverges from upstream). Upstream
  dispatches on the exact spelling, so a real magic method written
  `__ToString` or `__Get` is reported as a non-magic `__` method (and its
  contract is never checked), although PHP treats it as the magic method.
  custos matches magic names, the known non-magic list, the missing
  underscore list and the parent-call search case-insensitively.
- A bare `return;` inside a nested closure of `__toString`/`__sleep`/… is
  reported by C-returns upstream (inconsistent with E4). Recommendation:
  ignore returns that belong to nested function-likes entirely.
- Subclass returns (custos diverges from upstream). Upstream compares
  returned types by set membership, so `__set_state` returning an instance
  of a subclass of the containing class is reported, although a subclass
  instance satisfies the contract. custos accepts subtypes of the allowed
  class types (C-returns).
- **Always-throwing bodies (custos diverges).** Upstream reports a
  `__toString()` (or `__sleep()`, …) without any `return` even when every
  path throws — a deliberate "not supported" implementation, common in test
  doubles. Such a method never returns a wrong value; custos skips it when
  the body cannot complete normally.
- **`mixed` results (custos diverges).** A returned value typed `mixed`
  (`return call_user_func($this->fn);`) may well be of the required type;
  custos treats it like an unknown type (no report) instead of reporting
  `got 'mixed'`.
- **Missing-underscore fix dropped (custos diverges).** Upstream renames
  `_set` to `__set` (and the other D15 near-misses). The rename breaks every
  caller of the method (`$this->_set($n)`), and the existing signature may
  not satisfy the magic contract: a one-parameter `_set($number)` became a
  fatal "Method __set() must take exactly 2 arguments" (found on real
  code). custos keeps the report without a fix. EA case affected:
  `magic-methods-missing-underscore.php` (listed divergence).
- **`never` return type (custos diverges).** A declared `: never` on
  `__toString`, `__debugInfo`, `__serialize`, `__sleep` or `__set_state`
  is accepted by PHP (the method always throws); upstream reports it as
  "got 'never'". custos treats `never` as satisfying any return contract.
- **Inherited single-underscore names (custos diverges).** D15 does not
  report `_get`, `_set`… when a parent, interface or trait of the class
  declares the same method: the name is imposed by the hierarchy (an
  abstract cache class's `_get()` hook, 8 reports on PrestaShop), not a
  misspelt magic method. The declaring ancestor is reported as before
  unless it is abstract (E1).
- **Empty parent methods (custos diverges).** C-parent does not report a
  method whose parent version has an empty body and promotes no
  constructor parameter (`public function __construct() {}`): calling it
  would do nothing (Doctrine/PrestaShop base classes). Builtin parents are
  always checked (their stub bodies say nothing).
- **Single-underscore hooks (custos diverges, D15).** `_construct()`,
  `_get()`, `_call()`… are not reported when the class hierarchy also
  declares the real magic method (`__construct()` calling
  `$this->_construct()`, Magento models and blocks) or when the file calls
  the method by name (`$this->_get($k)`, `$this->_call($args)`): they are
  deliberate hooks, not misspelt magic methods.
