---
id: AccessModifierPresented
group: Code style
kind: syntax
needs: []
php: { min: "", max: "" }
---

# AccessModifierPresented

## Summary

Class members that rely on PHP's implicit public visibility (methods declared
with no visibility keyword, properties declared with `var`/`static`/`readonly`
only, constants declared with a bare `const`) should spell out their visibility
explicitly so intent is clear to readers.

## Detection

The rule visits every class-like declaration: classes (named and anonymous),
abstract classes, interfaces, traits and enums. Only members declared directly
in that declaration are inspected (no inherited/used members).

D1. **Methods.** For each method that has a name token and whose effective
    visibility is public (i.e. no `private`/`protected` keyword): report when
    the method's modifier keyword list (possibly empty) does not contain the
    text `public` (case-insensitive). This covers:
    - no modifiers at all: `function run() {}`;
    - only `static`, `final`, `abstract` (any combination, any order);
    - methods inside interfaces, traits and enums with no visibility.

D2. **Properties** (non-constant fields, i.e. declared in a property
    declaration statement): for each declared property variable whose
    effective visibility is public, take the modifier keyword list of the
    enclosing property declaration statement; report when that keyword text
    (lowercased) contains none of `public`, `protected`, `private`. Typical
    triggers: `var $x;`, `static $x;`, `readonly int $x;`,
    `static readonly`… A statement declaring several properties
    (`var $a, $b;`) reports **each** property name.

D3. **Constants** (only when option `ANALYZE_CONSTANTS` is on **and** the
    configured PHP language level is ≥ 7.1): report a class constant when it is
    the **first** constant of its `const` statement **and** the statement
    carries no visibility keyword (`public`/`protected`/`private`). Other
    modifiers do not count as visibility: `final const X = 1;` (PHP 8.1) is
    reported. Only the first name of a multi-constant statement is reported
    (`const A = 1, B = 2;` reports `A` only).

D4. When option `ANALYZE_INTERFACES` is off, interfaces are skipped entirely
    (methods and constants). With it on (default) interfaces are handled like
    any other class-like.

## Exceptions (no report)

E1. A method/property whose modifier text contains `public` (D1) or any of
    `public`/`protected`/`private` (D2). Note D2 uses substring matching, so
    asymmetric-visibility property modifiers such as `public(set)`,
    `protected(set)`, `private(set)` (PHP 8.4) count as explicit visibility
    and are never reported, with or without hooks.
E2. Methods/properties that are `private` or `protected`.
E3. Constants that already have a visibility keyword (`final public const X`).
E4. The second and later constants in a `const` statement.
E5. Constants when PHP level < 7.1 (constant visibility did not exist), or
    when `ANALYZE_CONSTANTS` is off.
E6. Members without a name token (parse-error recovery).
E7. Members of interfaces when `ANALYZE_INTERFACES` is off.

## Report

- Range: the member's name token only:
  - method: the identifier after `function` (e.g. `run`);
  - property: the variable token including `$` (e.g. `$count`);
  - constant: the constant identifier (e.g. `LIMIT`).
- Severity: info (fixtures tag it `weak_warning`).
- Message: `Declare the visibility of '{name}' explicitly.` where `{name}` is
  the member name without `$`.

## Fix

F1. **Methods and properties** (D1, D2): the declaration's modifier keyword
    list is rewritten as a canonical list built from the original keywords, in
    this exact order, separated by single spaces:
    `final` (if present) · `abstract` (if present) · `public` ·
    `static` (if present) · `readonly` (if present).
    The keyword `var` is dropped (it is replaced by `public`). Any other
    original keyword is not carried over. Only the span of the modifier
    keywords is replaced; the type, the `function` keyword, the name and the
    whitespace that followed the original modifier list are kept untouched.
    - empty modifier list (method with no keywords): insert `public ` right
      before the `function` keyword (after attributes / doc comment);
    - `static function f()` → `public static function f()`;
    - `final function f()` → `final public function f()`;
    - `abstract function f();` → `abstract public function f();`;
    - `var $x;` → `public $x;`; `static $x;` → `public static $x;`;
    - `readonly   int $x;` → `public readonly   int $x;` (original spacing
      after the list is preserved).
    For a statement declaring several properties, every reported property
    yields the same edit; apply it once.
F2. **Constants** (D3): insert `public ` immediately before the `const`
    keyword of the statement: `const A = 1, B = 2;` → `public const A = 1, B = 2;`;
    `final const X = 1;` → `final public const X = 1;`.

## Options

| Option | Type | Default | Effect |
|---|---|---|---|
| ANALYZE_INTERFACES | bool | true | When false, interface declarations are not inspected at all. |
| ANALYZE_CONSTANTS | bool | true | When false, class constants are never reported (D3 disabled). |

## PHP versions

- D3/F2 only when the project PHP level is ≥ 7.1.
- Methods/properties: no gating.
- Conformance note: the upstream fixture run without an explicit PHP level
  expects a bare `const` **not** to be reported (upstream's test default level
  is below 7.1), while the run at 7.1 expects it reported. The harness's
  "default" PHP level for EA cases without a level must therefore resolve to
  < 7.1 for this rule to pass (or the case must be treated as such).

## Examples

```php
<?php

class Invoice
{
    var <weak_warning descr="Declare the visibility of 'total' explicitly.">$total</weak_warning>;
    static <weak_warning descr="Declare the visibility of 'registry' explicitly.">$registry</weak_warning> = [];
    readonly int <weak_warning descr="Declare the visibility of 'number' explicitly.">$number</weak_warning>;
    protected $lines;
    private(set) string $currency;

    function <weak_warning descr="Declare the visibility of 'render' explicitly.">render</weak_warning>() {}
    final static function <weak_warning descr="Declare the visibility of 'make' explicitly.">make</weak_warning>() {}
    public function save() {}
}

abstract class Shape
{
    abstract function <weak_warning descr="Declare the visibility of 'area' explicitly.">area</weak_warning>();
}

interface Printable
{
    function <weak_warning descr="Declare the visibility of 'printOut' explicitly.">printOut</weak_warning>();
}
```

```php
<?php

class Invoice
{
    public $total;
    public static $registry = [];
    public readonly int $number;
    protected $lines;
    private(set) string $currency;

    public function render() {}
    final public static function make() {}
    public function save() {}
}

abstract class Shape
{
    abstract public function area();
}

interface Printable
{
    public function printOut();
}
```

PHP ≥ 7.1:

```php
<?php

class Limits
{
    const <weak_warning descr="Declare the visibility of 'MAX_ROWS' explicitly.">MAX_ROWS</weak_warning> = 500,
          MIN_ROWS = 1;
    protected const PAGE = 20;
}
```

```php
<?php

class Limits
{
    public const MAX_ROWS = 500,
          MIN_ROWS = 1;
    protected const PAGE = 20;
}
```

## Divergences

- Promoted constructor parameters are not covered upstream in a reliable way
  (they are not part of a property statement). Recommendation: do not report
  them.
- Typed class constants (PHP 8.3, `const int X = 1;`): upstream behaviour is
  uncertain (the type may count as "something before the name"). Recommendation:
  treat the type as not being a modifier and report like an untyped constant;
  no EA fixture covers it.
- custos diverges: `final const X` (PHP 8.1) is reported (D3). Upstream treats
  any modifier before `const` as explicit visibility, but `final` says nothing
  about visibility, which stays implicitly public. The fix inserts `public`
  before `const`, giving `final public const X`.
