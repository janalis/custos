---
id: CallableMethodValidity
group: Probable bugs
kind: semantic
needs: [names, index, hierarchy, types]
php: { min: "", max: "" }
---

# CallableMethodValidity

## Summary

`is_callable()` on a method callback that names a non-public method, or a
non-static method without an object, describes a callback that cannot really
be invoked from the outside (or will break when invoked statically). Flag the
argument so the method's visibility/static-ness gets fixed.

## Detection

Visit every function call.

- **D1** The call resolves to the global function `is_callable`: the name is
  compared case-insensitively, as PHP does (`Is_Callable(...)` qualifies);
  `\is_callable(...)` and an unqualified call in the global namespace match;
  an unqualified call in a namespace matches only when no `is_callable`
  function is declared in that namespace or imported with `use function`;
  `Foo\is_callable(...)` does not match (custos diverges).
- **D2** Exactly one argument is passed. Call it `A` (as written, including
  any parentheses around it).
- **D3** Run *value discovery* (below) on `A`. Continue only when it yields
  exactly **one** value `V`.
- **D4** `V` is either
  - a string literal (single- or double-quoted; see Divergences for
    interpolated strings), or
  - an array literal (`[...]` or `array(...)`) with exactly **two** elements.
- **D5** `V` describes a *member* callback:
  - string: `"Class::method"` (class part may carry a leading `\`; it is a
    fully-qualified name, not resolved against the current namespace/imports);
    a plain function name such as `'trim'` is not a member callback → stop;
  - two-element array: element 0 is the class/object part, element 1 is a
    string literal holding the method name.
- **D6** The callback resolves to a method declaration `M`: the class part is
  resolved to class(es) — a string literal names a class by FQN, `X::class`
  names class `X` (resolved with the file's namespace/imports), any other
  expression contributes its inferred object type(s) — and the method name is
  looked up (case-insensitively, including inherited methods) in those
  classes. Unresolvable class or method → stop.
- **D7 (visibility)** If `M` is not `public` (private or protected) and is
  not accessible from the call's class scope → report kind P. The class
  scope is the nearest enclosing class-like declaration (closures and arrow
  functions keep it; a named function declared inside a method has none;
  top-level code has none). `M` is accessible when:
  - `M` is private and the scope is `M`'s declaring class, or `M` is
    declared in a trait that the scope class uses (directly or through its
    ancestors);
  - `M` is protected and the scope class is the declaring class, a subclass
    of it, or one of its ancestors; for an anonymous class, its `extends`
    class must be a subclass of the declaring class.
  Class names compare case-insensitively. An inaccessible method is still
  not reported when a magic handler takes the call (E7).
- **D8 (static-ness)** If `M` is not `static`, report kind S when:
  - `V` is a string literal (`'Class::method'`), or
  - `V` is an array and element 0's inferred type contains `string` or
    `callable` (e.g. a string literal class name `'Shop'`, a variable holding
    a string), or element 0 is a `X::class` expression.
  Element 0 being an object (`new Shop()`, `$this`, a typed object variable)
  does not trigger S.

  S is not reported when PHP binds the current `$this` to the callback
  (custos diverges): the call has a `$this` and its class scope (as in D7)
  is the class the callback names or one of its subclasses. A `$this`
  exists inside a non-static method, including its closures and arrow
  functions unless one of them (at any nesting level) is declared `static`.
  The named class is the one written in the callback (`parent::class`,
  `self::class`, `Base::class`, `'Base'`, `'Base::boot'`), not the class
  declaring `M`. Inside a trait the final class is unknown, so a non-static
  trait method is treated as compatible with any named class.

Both P and S can be reported for the same argument (two findings on the same
range), P first.

### Value discovery

Shared procedure, applied to an expression `e` (with a visited-set guarding
against cycles; an already-visited node yields nothing):

1. Strip any parentheses around `e`.
2. Ternary `c ? a : b` → union of discovery on `a` and on `b`.
3. `a ?? b` → union of discovery on `a` and `b`.
4. Variable `$v` → find the nearest enclosing function, method or closure. If
   there is none (top-level code), the result is **empty**. Otherwise: if a
   parameter of that scope is named `v` and has a default value, add
   discovery of the default. Then for every plain assignment `$v = …`
   anywhere inside that scope's body (regardless of position, including
   inside nested closures), add discovery of the assigned value; for chained
   assignments `$v = $w = X` use the innermost value `X`. Destructuring is
   ignored. **Unstable variable** (custos refinement, see Divergences): if,
   anywhere in that same scope body (same search area as the assignments,
   nested closures included), `$v` is the operand of `++` / `--` (prefix or
   postfix) or the target of a compound assignment (`+=`, `-=`, `*=`, `/=`,
   `%=`, `**=`, `.=`, `&=`, `|=`, `^=`, `<<=`, `>>=`, `??=`), the variable's
   value set is **unknown**.
5. Property fetch (`$o->p`, `X::$p`) → resolve to the property declaration;
   if it has a default value whose source text does not end with the property
   name, add discovery of it. Also add discovery of values from plain
   assignments to an equivalent property fetch inside the current scope and
   inside the declaring class's constructor.
6. Class constant `X::C` → resolve; add discovery of its value expression.
7. Global constant (other than `true`/`false`/`null`) → resolve to its
   `define('C', value)` declaration (including stub declarations such as
   `PHP_INT_MAX`) and add the value expression itself (not further
   discovered). Unresolved → empty. (Whether top-level `const C = value;`
   declarations resolve here upstream is unverified; custos treats them like
   `define()`.)
8. Anything else (including `true`/`false`/`null`) → the expression itself.

**Unknown result.** If any variable expanded during a discovery (at any
depth: through a ternary branch, a `??` operand, an assignment chain, a
parameter default) has an unknown value set, the **whole** discovery result
is *unknown*, not merely missing that variable's values. Every consumer of
value discovery (this rule and every spec that refers to this procedure)
treats an unknown result as "stop, no report" for the check that relies on
it — even where an empty or a multi-value result would lead to a report, and
even where the consumer would otherwise filter the candidates (e.g. "keep
only string literals"). Example: `$n = 0; foreach ($xs as $x) { ++$n; }
mt_rand(1, $n);` — `$n` is unknown, so nothing is reported.

## Exceptions (no report)

- **E1** `is_callable()` with zero or 2+ arguments.
- **E2** Value discovery yields zero or several candidates (e.g. a top-level
  variable, a ternary with two different literals).
- **E2b** Value discovery result is unknown (a variable on the path is
  incremented/decremented or compound-assigned in its scope).
- **E3** Arrays with one element (`['Shop::open']`) or 3+ elements; plain
  function-name strings.
- **E4** Method not found (unknown class; magic `__call`/`__callStatic`
  handlers are not considered for a method that does not exist).
- **E5** Public method + object receiver; public static method in any form.
- **E6** Non-public method checked from a scope that may call it (D7), e.g.
  `is_callable([$this, 'helper'])` inside the declaring class or, for a
  protected method, inside a subclass.
- **E7** The method exists but is not accessible from the call's scope (P
  would be reported), and the named class declares or inherits the magic
  handler PHP routes the call to: `__call` when an object is involved (an
  object receiver, or a callback without an object to which `$this` is
  bound as in D8), `__callStatic` for a callback without an object and
  without a bound `$this`. Neither P nor S is reported. The handlers are
  not interchangeable (`__callStatic` does not rescue `[$obj, 'hidden']`,
  `__call` alone does not rescue `'Cls::hidden'` outside the class), and
  an accessible non-static method referenced without an object is still
  reported S even when `__callStatic` exists: PHP does not route it there.
- **E8** Non-static method named through its class from a non-static
  method of that class or of a subclass (D8), e.g.
  `is_callable([parent::class, 'boot'])` before `parent::boot()`.

## Report

- Range: the argument `A` exactly as written in the call (when `A` is a
  variable whose value was discovered elsewhere, the variable is
  highlighted, not the literal).
- Severity: warning.
- Messages (`{m}` = the resolved method's declared name):
  - P: `Method '{m}' is not public, so the callback cannot be invoked from outside.`
  - S: `Method '{m}' is not static but is referenced without an object.`

## Fix

None.

## Options

None.

## PHP versions

None.

## Examples

```php
<?php
class Shop
{
    public function open() {}
    protected function audit() {}
    public static function hours() {}
    private static function vault() {}
}

function probe(Shop $shop)
{
    $handler = ['Shop', 'open'];
    return [
        is_callable('ucfirst'),
        is_callable(<warning descr="Method 'open' is not static but is referenced without an object.">[Shop::class, 'open']</warning>),
        is_callable(<warning descr="Method 'open' is not static but is referenced without an object.">'Shop::open'</warning>),
        is_callable(<warning descr="Method 'open' is not static but is referenced without an object.">$handler</warning>),
        is_callable(['Shop', 'hours']),
        is_callable('Shop::hours'),
        is_callable([$shop, 'open']),
        is_callable([$shop, 'hours']),
        is_callable(<warning descr="Method 'audit' is not public, so the callback cannot be invoked from outside.">[$shop, 'audit']</warning>),
        is_callable(<warning descr="Method 'vault' is not public, so the callback cannot be invoked from outside.">[Shop::class, 'vault']</warning>),
        is_callable(['Shop::open']),
        is_callable([$shop, 'missing']),
    ];
}
```

Unknown discovery result (no report):

```php
<?php
class Gate { protected function pass() {} }
function pick(bool $alt) {
    $cb = 'Gate::pass';
    if ($alt) { $cb .= 'Twice'; }         // compound write: $cb is unknown
    return [is_callable($cb), is_callable($alt ? $cb : 'trim')];
}
```

Both findings on one argument (private and non-static, string class name):

```php
<?php
class Locker { private function seal() {} }
function check() {
    return is_callable(<warning descr="Method 'seal' is not public, so the callback cannot be invoked from outside."><warning descr="Method 'seal' is not static but is referenced without an object.">['Locker', 'seal']</warning></warning>);
}
```

Bound `$this` and magic handlers (no report unless marked):

```php
<?php
class Engine
{
    public function start() {}
    private function prime() {}
}

class Turbo extends Engine
{
    public function start()
    {
        $late = static fn() => is_callable(<warning descr="Method 'start' is not static but is referenced without an object.">[Engine::class, 'start']</warning>);
        return [is_callable([parent::class, 'start']), is_callable('Engine::start'), $late];
    }

    public static function build()
    {
        return is_callable(<warning descr="Method 'start' is not static but is referenced without an object.">[parent::class, 'start']</warning>);
    }
}

class Relay
{
    private function fetch() {}
    public function __call($name, $args) {}
}

function relay(Relay $relay)
{
    return [
        is_callable([$relay, 'fetch']),
        is_callable(<warning descr="Method 'fetch' is not public, so the callback cannot be invoked from outside."><warning descr="Method 'fetch' is not static but is referenced without an object.">'Relay::fetch'</warning></warning>),
    ];
}
```

## Divergences

- **Function name (custos diverges).** Upstream matches the written name
  case-sensitively and without resolving it, so `IS_CALLABLE([$this, 'helper'])` is missed while
  a namespaced user function of the same name (declared in the namespace,
  imported with `use function`, or called qualified) is checked as if it
  were the builtin. custos compares the name case-insensitively and
  requires the call to reach the global function (D1).
- **Unstable variables in value discovery — custos refinement, not
  upstream.** Upstream ignores `++`/`--` and compound assignments when
  collecting a variable's values, so `$n = 0; … ++$n; mt_rand(1, $n)` is
  analysed as if `$n` were always `0` (IncorrectRandomRange then reports
  min > max), and `$spec = 'P1D'; $spec .= 'T2H';` is checked as `'P1D'`
  (found on real code). custos makes such a variable's value set, and the
  whole discovery result, unknown; consumers stay silent. This affects every
  rule using value discovery (the specs referring to this procedure and
  those restating it: ForeachInvariants, DynamicCallsToScopeIntrospection,
  GetTypeMissUse, MockingMethodsCorrectness, OffsetOperations,
  StrTrUsageAsStrReplace, UnnecessaryCasting, UnnecessaryAssertion,
  NotOptimalRegularExpressions, PrintfScanfArguments). No upstream fixture of
  any of these rules relies on a variable that is also incremented or
  compound-assigned (PrintfScanfArguments' only such fixture expects no
  report), so conformance is unaffected.
- Interpolated double-quoted strings (`"$cls::open"`) are string literals but
  cannot describe a fixed callback; upstream behaviour unverified. custos
  skips strings containing interpolation.
- Keyed two-element arrays (`['c' => Shop::class, 'm' => 'open']`) are not
  valid callables; custos skips arrays with explicit keys.
- Short ternary `a ?: b` in value discovery: custos takes the union of
  discovery on `a` and `b`. No fixture covers it.
- **Calling scope (custos diverges):** upstream reports every non-public
  method, wherever `is_callable()` is called. PHP evaluates visibility from
  the calling scope, so inside the declaring class (or, for a protected
  method, inside the same hierarchy) `is_callable()` returns true and the
  callback is invokable there; reporting it is a false positive. custos
  checks the scope (D7, E6).
- **Bound `$this` (custos diverges):** upstream reports every non-static
  method referenced without an object. PHP attaches the current `$this` to
  such a callback when the calling class is the named class or a subclass,
  exactly as it does for `parent::method()`, so `is_callable()` is true and
  the callback runs with the right object. custos stays silent there (D8,
  E8) and keeps the report in static methods, static closures, unrelated
  classes and plain functions, where the callback really fails.
- **Magic handlers (custos diverges):** upstream reports a non-public
  method even when the class routes inaccessible calls to `__call` /
  `__callStatic`, which makes the callback invocable. custos stays silent
  when the handler matching the callback form exists (E7).
