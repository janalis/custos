---
id: SenselessProxyMethod
group: Unused
kind: semantic
needs: [names, hierarchy, index]
php: { min: "", max: "" }
---

# SenselessProxyMethod

## Summary

An override whose only statement forwards its parameters unchanged to the
same-named parent method — with the same signature, modifiers and
documentation tags — adds nothing; inheritance already provides that
behaviour. Remove it.

## Detection

Visit every class (interfaces and traits are skipped; enums are not
relevant) and each method `M` declared directly in it.

### Shape of `M`

- **D1** `M` is not abstract, not private, and carries **no attributes**
  (`#[...]` on the method itself).
- **D2** `M`'s body contains exactly one statement (ordinary comments and
  `/** */` doc comments are not statements).
- **D3** That statement is either `return E;` or an expression statement
  `E;`, and `E` is **directly** (no parentheses, no cast, no `@`) a static
  method call whose class part is the keyword `parent` (any letter case) and
  whose method name is `M`'s name (compared case-insensitively, as PHP
  compares method names).
- **D4** The call passes exactly as many arguments as `M` declares
  parameters, and the i-th argument is a plain variable whose name is the
  i-th parameter's name (`parent::save($a, $b)` for `save($a, $b)`).
  Unpacking (`...$xs`), named arguments, casts, reordering, literals → no
  report.

### Signature unchanged

- **D5** The call resolves to a parent method `PM` (unresolvable → no
  report).
- **D6** `PM` has the same number of parameters as `M`, and both have
  identical: abstract-ness (so an abstract `PM` → no report), static-ness,
  final-ness, and visibility (absent modifier = `public`). Neither returns
  by reference (`function &m()`): a forwarded call result is not a
  reference to the parent's variable.
- **D7** For each parameter position i:
  - default values: both absent, or both present and *equivalent* (same
    structure ignoring whitespace/comments, or identical text);
  - `M`'s default must also *mean* the same in both classes, so it contains,
    at any depth:
    - no magic constant (`__CLASS__`, `__LINE__`, …);
    - no `self`, `static` or `parent` (`self::LIMIT` reads the child's
      constant in `M` and the parent's in `PM`);
    - no other class or constant name, except when both declarations sit
      in the same namespace block of the same file (same namespace, same
      imports) or the name is fully qualified (`\Lib\Table::WIDTH`).
      `null`, `true` and `false` are always fine;
  - declared (native) parameter types are equal after name resolution;
    untyped vs typed differs;
  - neither parameter is variadic: `M` receives its variadic arguments
    packed in one array, so passing that variable on (D4) hands the parent
    a single array instead of the original arguments;
  - both are by-reference or both are by-value;
  - `M`'s parameter carries no attributes;
  - for `__construct`: if either side's parameter is constructor-promoted,
    both must be promoted with the same visibility, that visibility is not
    `private` (a private promoted property is the child's own and would
    vanish with the constructor), and neither property has an asymmetric
    write visibility or hooks.
  Parameter names are not compared.
- **D8** Return types: both absent, or both present and equivalent
  (structure or identical text; `\Exception` vs `\RuntimeException`,
  `string` vs none → differ).
- **D9** Doc-block tags: only if `M` has a doc comment — take the set of its
  tags (`@something …` elements; free description text is ignored). `PM`'s
  tags likewise (empty if `PM` has no doc comment). Require the same number
  of tags, and every tag of `M` to have an equivalent tag in `PM`. When `M`
  has no doc comment, nothing is compared.

### Return value preserved

- **D10** When the statement is `E;` (no `return`), the parent method must be
  known to produce no value, otherwise no report (removing the child would
  start returning the parent's value). `PM` produces no value when:
  - its native return type is `void` or `never`; or
  - it has no native return type and its declaration is available (same
    file): its body contains no `return` with an expression and no
    `yield`/`yield from` (nested closures, arrow functions, functions and
    classes are not searched); or
  - it has no native return type, its declaration is not available, and its
    doc `@return` type is `void`.
  Any other native return type, or an unavailable declaration without a
  `void` doc return, → no report.
  `return parent::m($a);` qualifies regardless of what the parent returns.

When D1–D10 all hold and E7–E9 do not apply, report.

- **Name case.** Wherever this rule compares two expressions for
  equivalence, the names PHP resolves case-insensitively — function and
  method names, class names in calls, `new`, `instanceof` and `::`
  accesses, and keywords — compare case-insensitively (`Cache::$map['k']`
  matches `cache::$map['k']`, `$o->Name()` matches `$o->name()`); variable,
  property and constant names stay case-sensitive.

## Exceptions (no report)

- **E1** Abstract/private methods, methods with attributes, interfaces and
  traits.
- **E2** Bodies with more than one statement or zero statements.
- **E3** Arguments modified, reordered, added or dropped.
- **E4** Calling a different parent method (`parent::other($a)`), or
  `self::`/`static::`/`$this->` calls.
- **E5** Visibility/static/final changes; by-reference returns; different
  defaults (incl. added or removed defaults); defaults using magic
  constants, `self`/`static`/`parent` or names that may resolve differently
  in the parent's file; type changes; variadic parameters; by-reference
  marker changes; parameter attributes; promotion changes, private
  promotion, asymmetric or hooked promoted properties; return type changes;
  extra or different doc tags on the child.
- **E6** The child calls the parent without `return` while the parent
  returns a value, is a generator, declares a non-void return type, or
  cannot be shown to return nothing (D10).
- **E7** (custos, see Divergences) Below PHP 8.0, a `__construct` proxy in a
  class outside any namespace that also declares a method named like the
  class (case-insensitively).

- **E8** (custos, see Divergences) The class gets a method of the same
  name from a trait: a trait it uses, or one used by such a trait,
  declares it, or a `use … { x as name; }` adaptation of the class or of
  one of those traits introduces it. A used trait that cannot be resolved
  counts as declaring it. A trait method takes precedence over an inherited
  one, so the override is what keeps the parent's implementation in
  effect.

- **E9** (custos, see Divergences) A parameter has a default value and the
  parent method's body, when it is declared in the analysed file, calls
  `func_num_args()`, `func_get_args()` or `func_get_arg()`. The proxy
  passes every declared parameter, so the parent sees an argument the
  caller left out; without the proxy it would not.

## Report

- Range: the method's name identifier.
- Severity: info (weak warning).
- Message: `Method '{name}' only forwards to its parent; remove it.`

## Fix

- **F1** Remove the method:
  - if the method's previous sibling (skipping whitespace and ordinary
    comments) is a `/** … */` doc comment, delete it (whitespace before it
    stays);
  - if whitespace immediately follows the method's closing `}`, delete that
    whitespace run;
  - delete the method itself (from its first modifier/`function` keyword
    through the closing `}`).

## Options

None.

## PHP versions

None (attributes, promoted parameters and return types are simply parsed at
any level).

## Examples

```php
<?php
class Store
{
    protected function __construct($cfg = []) {}

    public function put($key, $value = 0) { return true; }

    public function get($key) {}

    public function drop($key) {}

    public function tag($key) {}

    /**
     * @api
     */
    public function flush($all) {}

    public function label($text = __CLASS__) {}

    public function size(): int { return 0; }

    public function keys(array $filter) { return []; }

    public static function make() {}
}

class DiskStore extends Store
{
    public function __construct($cfg = [])
    {
        parent::__construct($cfg);
    }

    /** Disk variant. */
    public function <weak_warning descr="Method 'put' only forwards to its parent; remove it.">put</weak_warning>($k, $v = 0)
    {
        return parent::put($k, $v);
    }

    public function <weak_warning descr="Method 'get' only forwards to its parent; remove it.">get</weak_warning>($key)
    {
        // just delegate
        parent::get($key);
    }

    public function drop($key)
    {
        parent::drop(trim($key));
    }

    /**
     * @internal
     */
    public function tag($key)
    {
        parent::tag($key);
    }

    /**
     * @api
     */
    public function <weak_warning descr="Method 'flush' only forwards to its parent; remove it.">flush</weak_warning>($all)
    {
        parent::flush($all);
    }

    public function label($text = __CLASS__)
    {
        parent::label($text);
    }

    public function size(): ?int
    {
        return parent::size();
    }

    public function keys($filter)
    {
        return parent::keys($filter);
    }

    #[Cached]
    public static function make()
    {
        parent::make();
    }
}
```

```php
<?php
class Store
{
    protected function __construct($cfg = []) {}

    public function put($key, $value = 0) { return true; }

    public function get($key) {}

    public function drop($key) {}

    public function tag($key) {}

    /**
     * @api
     */
    public function flush($all) {}

    public function label($text = __CLASS__) {}

    public function size(): int { return 0; }

    public function keys(array $filter) { return []; }

    public static function make() {}
}

class DiskStore extends Store
{
    public function __construct($cfg = [])
    {
        parent::__construct($cfg);
    }

    public function drop($key)
    {
        parent::drop(trim($key));
    }

    /**
     * @internal
     */
    public function tag($key)
    {
        parent::tag($key);
    }

    public function label($text = __CLASS__)
    {
        parent::label($text);
    }

    public function size(): ?int
    {
        return parent::size();
    }

    public function keys($filter)
    {
        return parent::keys($filter);
    }

    #[Cached]
    public static function make()
    {
        parent::make();
    }
}
```

## Divergences

- **custos diverges from upstream** on dropped returns (D10). Upstream also
  reports a child that calls `parent::m($a);` without `return` when the
  parent returns a value; deleting such a child makes calls start returning
  the parent's value instead of `null`, so the suggested removal changes
  behaviour. custos only reports that shape when the parent provably
  returns nothing. Upstream's fixtures expect the report, so the affected
  EA case is listed in `testdata/ea-divergences.json`.
- Equivalence of parameter types is by resolved type set; a doc-only type
  change is ignored.
- **Letter case — custos diverges from upstream.** Upstream requires the
  class part to be exactly `parent` and the method name to match `M`'s name
  case-sensitively, so `PARENT::send()` or `parent::Send()` inside `send()`
  is missed although PHP calls the same parent method. custos compares both
  case-insensitively.
- **Name case (custos diverges).** Upstream compares the expressions
  textually, so operands that differ only in the case of a function, method
  or class name (`Stats::$n` vs `stats::$n`), which PHP treats as the same,
  are not recognised as equivalent. custos folds the case of those names
  (Detection, "Name case").
- **custos diverges — PHP 4 constructors (E7).** Below PHP 8.0, removing a
  `__construct` proxy from a non-namespaced class that also has a method
  named like the class turns that method into the constructor (found on
  WordPress's `pomo` streams, where the legacy method calls
  `self::__construct()`: infinite recursion after the fix). custos does not
  report such proxies.
- **custos diverges — signature details (D6, D7).** Upstream does not
  compare by-reference and variadic markers and treats a variadic
  parameter as an `array` one, and accepts any pair of textually equal
  defaults and any pair of equally visible promoted parameters. Removing
  such an override changes what the parent receives (one packed array
  instead of the arguments, a copy instead of a reference), which constant
  a default reads (`self::X`, or a name imported differently in the
  parent's file) or drops the child's private promoted property. custos
  does not report these.
- **custos diverges — argument counting (E9).** `passed($a, $b = null)`
  forwarding to a parent that returns `func_num_args()` answers 2 for
  `passed(1)`, and 1 once the override is removed. custos does not report
  a proxy with optional parameters over such a parent. A parent declared
  in another file is not inspected.
- **custos diverges — trait methods (E8).** With `use Loud;` in the class,
  deleting `greet()` makes `Loud::greet()` the effective method instead of
  the parent's. custos does not report an override shadowing a trait
  method.
