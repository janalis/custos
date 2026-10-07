---
id: ReturnTypeCanBeDeclared
group: Language level migration
kind: semantic
needs: [names, types, hierarchy, index]
php: { min: "7.0", max: "" }
---

# ReturnTypeCanBeDeclared

## Summary
A method without a declared return type whose returned values (and `@return`
documentation) all agree on one type — optionally nullable, or nothing at all
(`void`) — can carry a native return type declaration. Declaring it lets the
engine enforce the contract and documents it in the signature.

## Detection

### Candidates
- **D1** Methods only (class, abstract class, interface, trait and enum
  methods). Plain functions and closures are never reported. Property hooks
  (`get`/`set` hooks of PHP 8.4 properties) are never reported.
- **D2** The method has no declared return type.
- **D3** The method name is not one of (case-insensitive, as PHP compares
  method names: `__TOSTRING` is excluded too):
  `__construct`, `__destruct`, `__call`, `__callStatic`, `__get`, `__set`,
  `__isset`, `__unset`, `__sleep`, `__wakeup`, `__toString`, `__invoke`,
  `__set_state`, `__clone`, `__debugInfo`.
- **D4** Abstract methods (including all interface methods) are examined only
  when they have a doc comment containing a `@return` tag; otherwise skipped.

### Inferred type set
- **D5** Infer the method's return type set `R0` as the union of:
  - the types listed in its `@return` doc tag, if any;
  - for each `return expr;` belonging to the method itself (not to nested
    closures/functions/classes), the inferred type of `expr`;
  - `void` for each bare `return;`.
  Expression types come from normal inference: literals (`''` → string, …),
  `new X` → `\X` (resolved FQN), parameters → their declared type or
  `@param` doc type, built-in function return types (stubs),
  `@var`-annotated variables, etc. A type that cannot be determined is
  *unknown* (e.g. `mixed`-returning calls are `mixed`, which is known but not
  declarable; an unresolvable call, or an untyped parameter, is unknown).
- **D5b** (custos) For a `return $this->p;` whose property `p` has no
  native type, is not static/promoted/magic, has no default or a `null`
  default, and is not assigned (non-null) by a top-level statement of this
  class's constructor, also add `null`: such a property holds null until
  written, whatever its `@var` says (`@var int` on a nullable ORM column).
  When the property's own type is unknown (no `@var`), that return
  contributes `null` instead of an unknown member (the `@return` tag
  describes the written values).
  A doc type that is not a valid type or class name (`@return {array}`,
  `@return self::KIND_*`) contributes nothing (unknown).
- **D6** If `R0` contains any unknown member, stop — unless `R0` has exactly
  two members, exactly one of which is known (a contribution from an
  unresolvable parent/interface declaration must not block the report; in our
  model such contributions are simply not added).
- **D7** Normalise each known type (case-insensitive for the keys below):
  any type containing `[]` → `array`; `boolean`, `true`, `false` → `bool`;
  `integer` → `int`; `\Closure` → `callable`; `$this` → `static`; the
  built-in names `array iterable string bool int float number null void mixed
  callable resource static self object never parent`, written with or without a leading
  backslash, map to themselves without the backslash. Anything else (class
  names) stays as its fully-qualified name with a leading `\`. Let `R` be the
  resulting **set** (duplicates collapse).
- **D8 Generators**: if `R` lacks `\Generator` and the method's own body
  contains a `yield` (not one inside a nested closure), add `\Generator`;
  additionally, if the method contains no `return` statement at all (searching
  the whole body, nested closures included), remove `null`.
- **D9 Implicit null** (only when `R` is non-empty and the method is not
  abstract):
  - if `R` contains neither `null` nor `void`, and the last statement of the
    method body is neither a `return` nor a `throw`, add `null`;
  - then, if `R` is exactly `{null}`: take the first `return` statement in the
    body in document order (nested closures included); if it has a value that
    is not the null literal, remove `null` (making `R` empty).

### Decision (L = configured language level)
- **D10 Empty set** (`|R| = 0`), only when L ≥ 7.1: let `r` be the first
  `return` statement in the method in document order (any depth). Suggest
  `void` when `r` does not exist or belongs to a nested closure/function.
- **D11 Single type** (`|R| = 1`, type `t`):
  - suggested `s` = `void` if L ≥ 7.1 and `t` ∈ {`null`, `void`}; otherwise
    `s` = compact(`t`) (see *compact(t)* below).
  - acceptable when `t` starts with `\`, or `t` ∈ {`self`, `array`,
    `callable`, `bool`, `float`, `int`, `string`}, or `s` ∈ {`self`,
    `static`}; or, failing that, when L ≥ 7.1 and `s` = `void`.
    (`iterable`, `object`, `mixed`, `resource`, `number`, and `null` below
    7.1 are never suggested.)
  - **static guard**: let `onlyStatic` = the method's `@return` tag exists and
    consists solely of the type `static` (no other type, no description). If
    `onlyStatic` and L < 8.0 → no report. If `onlyStatic` and L ≥ 8.0 → the
    suggestion becomes `static`.
- **D12 Two types** (`|R| = 2`), only when L ≥ 7.1: if `R` contains `void`,
  remove it; otherwise if it contains `null`, remove it. If exactly one type
  `t` remains:
  - `s` = `void` if `t` ∈ {`null`, `void`}, else compact(`t`).
  - acceptable when `t` starts with `\`, or `t` is in the D11 list, or `s` =
    `self` → suggestion `?s`; otherwise when `s` = `void` → suggestion `void`.
    (`static` is not accepted here.)
- Three or more types: no report.
- **D14 Return statements must stay valid** (custos diverges): unless the
  method's own body contains a `yield`, no report when the suggestion is
  `void` and the method's own body (nested closures/functions excluded)
  has a `return expr;` (any expression, `return null;` included — a compile
  error in a void function), or when the suggestion is anything else and the
  own body has a bare `return;` (a compile error under a non-void type).

### compact(t)
Only applies when `t` starts with `\` or equals `static`; otherwise returns
`t` unchanged.
1. If option LOOKUP_PHPDOC_RETURN_DECLARATIONS is on and the method's
   `@return` tag lists a type written exactly `self` or `$this` → `self`.
2. Else if `t` is `static` → `static`.
3. Else look for an import: walk outwards from the containing class through
   every enclosing statement list (namespace body, file top level); in each,
   consider the `use` import declarations directly in that list; the first
   class import whose fully-qualified target equals `t` (case-sensitive) gives
   its alias if it has one, else its last name segment.
4. Else, if the class is in a non-global namespace `\N\` and `t` starts with
   `\N\`, remove that prefix (`\N\Sub\Iface` in namespace `N` → `Sub\Iface`,
   same-namespace class → its short name).
5. Else `t` unchanged (e.g. `\stdClass` from a namespaced class stays
   `\stdClass`).

### Overridden methods
- **D13** The method counts as *overridden* when its class is not `final`, the
  method is not `final` and not `private`, and some other type declares its own
  method with the same name, among: all ancestors (parent classes
  transitively, implemented interfaces transitively, used traits
  transitively), or any known subclass/implementor anywhere in the project.
  Overridden methods are still reported, but without a fix and with a message
  that recommends changing the whole hierarchy.

## Exceptions (no report)
- **E1** Language level below 7.0.
- **E2** Functions, closures, property hooks, magic methods from D3, methods
  already declaring a return type.
- **E3** Abstract/interface methods without `@return`.
- **E4** Partially unknown inferred type (D6), `mixed`/`object`/`iterable`
  results, more than two types, or two non-null types (`int|string`).
- **E5** `@return static` only, below PHP 8.0.
- **E6** A method returning a value and also falling off the end with three
  resulting types (e.g. `string`, `\Generator`, `null`).

## Report
- Range: the method's name identifier.
- Severity: info (weak warning).
- Message: `Declare ': {type}' as the return type.`; for overridden methods:
  `Declare ': {type}' as the return type (update the whole hierarchy with a
  signature refactoring).` where `{type}` is the suggestion (`void`, `int`,
  `?\Foo`, `self`, `static`, `Sub\Iface`, …).

## Fix
- **F1** (not offered for overridden methods) Insert `: {type}` right after
  the closing `)` of the parameter list, i.e. `)` + `: ` + type.
  - Non-abstract: `function f($x) { … }` → `function f($x): T { … }`. If there
    was no whitespace between `)` and `{` (`f($x){`), a separating whitespace
    must end up between the type and `{` (upstream reformats so that `{`
    starts a new line; any whitespace compares equal).
  - Abstract/interface: `function g();` → `function g(): T;`.
  - Everything else (doc comments, body) is unchanged.

## Options
| Option | Type | Default | Effect |
|---|---|---|---|
| LOOKUP_PHPDOC_RETURN_DECLARATIONS | bool | true | Lets `@return self` / `@return $this` turn a class-typed suggestion into `self` (compact step 1). |

## PHP versions
- Nothing below 7.0.
- `void` and nullable `?T` suggestions (D10, D11-void, D12) need ≥ 7.1.
- `static` from an `@return static`-only tag needs ≥ 8.0.
- Upstream fixtures run at 7.1, one at 8.0 (static) and one at 8.4 (property
  hooks).

## Examples

All examples assume L = 7.1 (or later, below 8.0) with default options.

```php
<?php
namespace Shop {
    use Shop\Billing\Invoice as Bill;

    interface Pricing {
        /** @return int|null */
        function <weak_warning descr="Declare ': ?int' as the return type.">discount</weak_warning>();
        function undocumented();
    }

    final class Cart {
        public function __toString()
        {
            return 'cart';
        }

        /** @param float $p */
        public function <weak_warning descr="Declare ': float' as the return type.">price</weak_warning>($p) { return $p; }

        public function <weak_warning descr="Declare ': void' as the return type.">clear</weak_warning>() { $this->n = 0; }

        public function <weak_warning descr="Declare ': ?\ArrayObject' as the return type.">maybe</weak_warning>($f) {
            if ($f) {
                return new \ArrayObject();
            }
        }

        public function <weak_warning descr="Declare ': Bill' as the return type.">bill</weak_warning>() { return new \Shop\Billing\Invoice(); }

        public function <weak_warning descr="Declare ': Billing\Receipt' as the return type.">receipt</weak_warning>() { return new \Shop\Billing\Receipt(); }

        /** @return $this */
        public function <weak_warning descr="Declare ': self' as the return type.">touch</weak_warning>() { return $this; }

        /** @param mixed $v */
        public function raw($v) { return $v; }

        public function both($f) { return $f ? 1 : 'one'; }

        public function chars($s) { yield $s; return 1; }
    }
}

namespace Shop\Billing {
    class Invoice {}
    class Receipt {}
}
```

```php
<?php
namespace Shop {
    use Shop\Billing\Invoice as Bill;

    interface Pricing {
        /** @return int|null */
        function discount(): ?int;
        function undocumented();
    }

    final class Cart {
        public function __toString()
        {
            return 'cart';
        }

        /** @param float $p */
        public function price($p): float { return $p; }

        public function clear(): void { $this->n = 0; }

        public function maybe($f): ?\ArrayObject {
            if ($f) {
                return new \ArrayObject();
            }
        }

        public function bill(): Bill { return new \Shop\Billing\Invoice(); }

        public function receipt(): Billing\Receipt { return new \Shop\Billing\Receipt(); }

        /** @return $this */
        public function touch(): self { return $this; }

        /** @param mixed $v */
        public function raw($v) { return $v; }

        public function both($f) { return $f ? 1 : 'one'; }

        public function chars($s) { yield $s; return 1; }
    }
}

namespace Shop\Billing {
    class Invoice {}
    class Receipt {}
}
```

Overridden case (reported, no fix):

```php
<?php
class Shape {
    /** @param int $n */
    public function <weak_warning descr="Declare ': int' as the return type (update the whole hierarchy with a signature refactoring).">sides</weak_warning>($n) { return $n; }
}
class Square extends Shape {
    public function <weak_warning descr="Declare ': int' as the return type (update the whole hierarchy with a signature refactoring).">sides</weak_warning>($n) { return 4; }
}
```

Static guard (L = 7.4 → nothing; L = 8.0 → `: static`):

```php
<?php
abstract class Node {
    /** @return static */
    abstract public function copy();
}
```

## Divergences
- **Void with valued returns, types with bare returns (custos diverges):**
  upstream suggests `: void` for a method whose only result is null —
  including `return null;` or returning a `null`/`void`-documented value —
  and `?T` for a method mixing `return;` with `return new T;`. Applying
  either declaration is a compile error ("A void function must not return a
  value" / "A function with return type must return a value"). custos skips
  such methods (D14). Generators are unaffected (a bare `return;` is legal
  there). EA case affected: `return-type-hints.php` (`returnVoid` with
  `@param $x void`, and the `@param $x null` variant, are no longer
  reported).
- A method with no `@return` that returns `$this`/`new static` infers
  `static`; upstream then suggests `: static` regardless of language level (the
  8.0 guard only looks at the doc tag), producing invalid code below 8.0.
  Recommendation: below 8.0 never suggest `static` (suggest nothing). No
  upstream fixture covers it.
- D10 looks only at the *first* `return` in document order: if it lives in a
  nested closure while the method itself also has a value-less `return;`
  later, `void` is still suggested (harmless) — but with a value-returning
  later `return` this cannot happen because `R` would not be empty. Keep.
- Compact step 4 removes every occurrence of the namespace prefix, not only the
  leading one (`\A\B\A\C` in namespace `A` would become `BC`). Recommendation:
  strip only the leading prefix.
- The inferred-type model is PhpStorm's; exotic inference (generics, closures'
  `use` variables, conditional `@return`) may differ. Only the patterns in
  D5–D9 must match upstream fixtures: `@param`-typed parameters returned as is,
  `new` expressions, string literals, bare `return;`, missing final return,
  generators, and `mixed` from untyped built-ins such as `json_decode()`
  results indexed by `[]`.
- **Magic-method case (custos diverges):** upstream excludes the magic
  methods by exact name, so `__TOSTRING()` or `__DebugInfo()` get a return
  type suggestion (and a fix that may conflict with the magic signature).
  custos compares the names case-insensitively (D3).
- **Implicitly null properties (custos diverges).** Upstream trusts the
  property's `@var` for `return $this->p;`, so an entity getter
  `/** @return int */ function getId() { return $this->id; }` over
  `/** @var int */ private $id;` gets `: int` — a TypeError whenever the
  property is still null (new object, nullable column). custos adds `null`
  for untyped properties without a non-null default that the constructor
  does not assign (D5b) and suggests `?int`.
- **`never` and malformed doc types (custos diverges).** `return exit();`
  infers `never`, which upstream treats as a class name (`: \never`, invalid
  before 8.1 and a compile error in a returning method after); custos treats
  `never` (and `parent`) as non-suggestible built-ins. Doc types that are not
  type or class names (`{array}`) are ignored instead of becoming `\{array}`.
