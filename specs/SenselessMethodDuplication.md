---
id: SenselessMethodDuplication
group: Unused
kind: semantic
needs: [names, hierarchy, index, types]
php: { min: "", max: "" }
---

# SenselessMethodDuplication

## Summary

A child class method whose body is a copy of the inherited method's body is
dead weight: if the visibility is the same the override can simply be
deleted; if only the visibility differs, the override should delegate to
`parent::` instead of duplicating the code.

## Detection

Visit every method `M` (in a class `C`).

- **D1** `M` is not abstract, not private, and not deprecated (doc-block
  `@deprecated` tag, or a `#[\Deprecated]` attribute).
- **D2** Not a test context: the file path does not end with `Test.php`,
  `Spec.php` or `.phpt` and does not contain `/Fixtures/`; and the FQN of
  the containing class does not end with `Test` and does not contain
  `\Tests\` or `\Test\`.
- **D3** `C` is a class (not a trait or interface; enums are not relevant).
- **D4** `M` has a body with `n` statements, `1 ≤ n ≤ MAX_METHOD_SIZE`.
  Statements are the direct children of the body block; comments are not
  statements — neither `//`/`/* */` comments nor `/** */` doc comments are
  counted.
- **D5** Let `P` be the resolved **direct** parent class of `C`, and `PM` the
  method with the same name (case-insensitive lookup) found on `P` or
  anything `P` inherits (a grandparent's method counts when `P` does not
  override it). Stop if there is none, if `PM` is abstract, deprecated
  (as D1) or private, or if exactly one of `M` and `PM` is `static`.
- **D6** `PM` has a body whose statement count (same counting as D4) equals
  `n`.
- **D7** Compare the bodies statement by statement in order (comments and doc
  comments skipped): every pair must be *equivalent* — same node structure
  ignoring whitespace and comments anywhere inside, or identical source text.
  Modifiers (other than visibility, see D10) and attributes are **not**
  compared.
- **D7a** Signature check. `M` and `PM` must declare the same number of
  parameters and, position by position, the same parameter name, the same
  by-reference and variadic markers and the same declared type (written
  alike ignoring case and whitespace, with class names resolving to the same
  class). Neither side may use constructor promotion. For kind I (D10) the
  signatures must match in full: also equivalent default values (both absent,
  or equivalent expressions as in D7), the same return type and the same
  by-reference return marker. Kind X keeps its own declaration, so defaults,
  return type and by-reference return may differ there. Stop on any mismatch.
- **D8** Symbol check: collect, inside `M`'s body (and, for kind I, its
  parameter defaults), every class reference
  (e.g. `Foo` in `new Foo`, `Foo::X`, `Foo::bar()`, `instanceof Foo`),
  constant reference (e.g. `SOME_CONST`, also `true`/`false`/`null`) and
  plain function call name (method calls excluded). For each: if it
  resolves to a declaration, record its fully-qualified name as resolved by
  the file's namespace/imports (for unqualified functions/constants inside a
  namespace: the name resolution result, i.e. the global fallback when only
  the global one exists); otherwise record "unresolved". If the set is
  non-empty, require that it contains no "unresolved" entry and equals the
  same set computed for `PM`'s body. (So identical text referring to
  different classes/functions/constants in different namespaces, or to a
  missing class, is not reported.)
- **D9** Private-member guard: if `PM`'s body contains any member access
  (method call, property fetch, class constant or static member) whose
  object is `$this` or whose class part is literally `self`, and that
  access resolves to a **private** member → stop (the child cannot reach
  it). The same test is applied to `M`'s body seen from `C`: a `$this`
  access (also inside a nested closure) resolving to a private member of the
  child stops too, because the inherited method runs in the parent's scope
  and cannot reach it.
- **D11** Scope-bound code: stop when `M`'s body (or, for kind I, a
  parameter default) contains
  - `__CLASS__`, `__METHOD__`, `__NAMESPACE__` or `__LINE__` (`__FUNCTION__`,
    `__FILE__` and `__DIR__` are fine: the two declarations share the name
    and, being compared, the file);
  - a `static $var` declaration belonging to the method itself (not to a
    closure or function nested in it): an override owns its static
    variables, an inherited method shares the parent's;
  - a call to `get_class()` or `get_parent_class()` without arguments, or to
    `get_object_vars()`, `get_class_vars()` or `get_class_methods()`, whose
    result depends on the class the calling code is declared in.
- **D12** Trait guard: stop when `C` uses a trait that supplies a method of
  the same name (directly, through a nested trait, or through an `as` alias
  in `C`'s `use` block), or a trait that cannot be resolved. `C`'s own
  method is what keeps the trait's implementation from taking precedence
  over the inherited one.
- **D10** Report:
  - **I (identical)**: `M` and `PM` have the same visibility (absent
    modifier = `public`);
  - **X (proxy)**: visibilities differ (e.g. parent `protected`, child
    `public`).
- **Name case.** Wherever this rule compares two expressions for
  equivalence, the names PHP resolves case-insensitively — function and
  method names, class names in calls, `new`, `instanceof` and `::`
  accesses, and keywords — compare case-insensitively (`Cache::$map['k']`
  matches `cache::$map['k']`, `$o->Name()` matches `$o->name()`); variable,
  property and constant names stay case-sensitive.

## Exceptions (no report)

- **E1** Abstract, private or deprecated methods (child or parent); a
  static method paired with a non-static one.
- **E2** Test contexts (D2); traits and interfaces.
- **E3** Empty bodies, bodies longer than `MAX_METHOD_SIZE` statements, or
  differing statement counts.
- **E4** Any non-equivalent statement (`echo 1;` vs `echo 2;`).
- **E5** Same text resolving to different symbols, or unresolved symbols
  (D8); parent bodies touching private members via `$this`/`self`, child
  bodies touching the child's private members via `$this` (D9).
- **E7** Differing signatures (D7a): another parameter count, name, order,
  type, by-reference or variadic marker, promoted parameters; for kind I
  also another default (`limit($n = 50)` over `limit($n = 10)`), return type
  or by-reference return.
- **E8** Scope-bound bodies (D11): `return __CLASS__;`,
  `static $calls = 0;`, `get_object_vars($this)`.
- **E9** The class uses a trait providing the method, or an unknown trait
  (D12).
- **E6** No parent method (methods only declared in the child; interface
  methods do not count as parents).

## Report

- Range: the method's name identifier only.
- Severity: info (fixture markup `weak_warning`).
- Messages (`{name}` is the method name):
  - I: `Method '{name}' duplicates the inherited implementation; remove it.`
  - X: `Method '{name}' duplicates the inherited implementation; delegate to parent::{name}() instead.`

## Fix

- **F1 (I)** Remove the method:
  - if the method's previous sibling (skipping whitespace and ordinary
    comments) is a `/** … */` doc comment, delete that doc comment (the
    whitespace before it stays);
  - if whitespace immediately follows the method's closing `}`, delete that
    whitespace run;
  - delete the method itself (from its first modifier/attribute/`function`
    keyword through its closing `}`).
  The whitespace that preceded the method therefore now precedes the next
  member.
- **F2 (X)** Replace the method's body block (from `{` to `}` inclusive)
  with `{ RET parent::NAME(ARGS); }` where
  - `NAME` is the method name as declared;
  - `ARGS` is every parameter of `M` written as `$paramName`, in order,
    joined with `, ` (no `&`, no `...`, no defaults);
  - `RET` is `return ` (with trailing space) when `M` has a return type,
    otherwise empty. `M` has a return type when its declared return type is
    present and neither `void` nor `never`, or (no declaration) its `@return` doc type is not
    `void`, or its body contains a `return <expr>;` (outside nested closures
    /functions) or a `yield`.
  Output is compared whitespace-collapsed, so line layout of the new block
  is free; e.g. `{\n    return parent::save($row, $flag);\n}` is fine.
  No fix is offered (the report stays) when the delegating call would not
  be equivalent: `M` returns by reference (`function &m()`), or its body
  calls `func_num_args()`, `func_get_args()` or `func_get_arg()` — the
  call passes every declared parameter, including the ones the caller
  omitted.

## Options

| Option | Type | Default | Effect |
|---|---|---|---|
| `MAX_METHOD_SIZE` | int | `20` | Methods with more top-level statements than this are skipped. |

## PHP versions

None.

## Examples

```php
<?php
class Repo
{
    protected function load($id)
    {
        $row = ['id' => $id];
        return strtoupper(json_encode($row));
    }

    protected function save($row, $flag)
    {
        $this->log($row);
    }

    protected function count()
    {
        return 1;
    }

    protected function secret()
    {
        return $this->hidden();
    }

    private function hidden()
    {
        return 0;
    }

    public function log($x)
    {
        echo $x;
    }
}

class MiddleRepo extends Repo {}

class UserRepo extends MiddleRepo
{
    /**
     * Copied from Repo.
     */
    protected function <weak_warning descr="Method 'load' duplicates the inherited implementation; remove it.">load</weak_warning>($id)
    {
        // same code, different comments
        $row = ['id' => $id];

        /** stray doc block */
        return strtoupper(json_encode( $row ));
    }

    public function <weak_warning descr="Method 'save' duplicates the inherited implementation; delegate to parent::save() instead.">save</weak_warning>($row, $flag)
    {
        $this->log($row);
    }

    protected function count()
    {
        return 2;
    }

    protected function secret()
    {
        return $this->hidden();
    }
}
```

```php
<?php
class Repo
{
    protected function load($id)
    {
        $row = ['id' => $id];
        return strtoupper(json_encode($row));
    }

    protected function save($row, $flag)
    {
        $this->log($row);
    }

    protected function count()
    {
        return 1;
    }

    protected function secret()
    {
        return $this->hidden();
    }

    private function hidden()
    {
        return 0;
    }

    public function log($x)
    {
        echo $x;
    }
}

class MiddleRepo extends Repo {}

class UserRepo extends MiddleRepo
{
    public function save($row, $flag)
    {
        parent::save($row, $flag);
    }

    protected function count()
    {
        return 2;
    }

    protected function secret()
    {
        return $this->hidden();
    }
}
```

## Divergences

- Whether `true`/`false`/`null` and `self`/`static`/`parent` references
  resolve (and to which FQN) is an IDE detail upstream; an unresolved entry
  suppresses the report. custos treats `true`/`false`/`null` as resolved
  to themselves, and `self`/`static`/`parent` as the class they denote in
  each body (so a duplicated `new self()` resolves differently in
  child and parent and is not reported).
- F2 drops `...` for variadic parameters and `&` is irrelevant at call site;
  `parent::m($args)` for a variadic `...$args` changes behaviour.
  custos emits `...$args` for variadic parameters (no upstream fixture
  covers it).
- **Staticness compared — custos diverges from upstream** (D5). Upstream
  does not compare the `static` modifier, so a static child method whose
  body matches a non-static parent method (or the reverse) is reported as a
  duplicate and the proxy fix delegates across the static/instance
  boundary. The two are not interchangeable declarations (PHP rejects the
  override itself), so custos does not report them: the problem to fix
  there is the signature, not a duplicated body.
- **Name case (custos diverges).** Upstream compares the expressions
  textually, so operands that differ only in the case of a function, method
  or class name (`Stats::$n` vs `stats::$n`), which PHP treats as the same,
  are not recognised as equivalent. custos folds the case of those names
  (Detection, "Name case").
- **Signatures compared — custos diverges from upstream** (D7a). Upstream
  compares bodies only, so an override that exists to change a default
  value, a parameter name or order, a type or the return type is reported,
  and removing it changes results. custos requires matching signatures;
  the proxy kind, which keeps the declaration, tolerates other defaults and
  another return type.
- **Scope-bound bodies, traits and child-private members — custos
  diverges** (D9, D11, D12). Textually identical bodies are not duplicates
  when they read `__CLASS__`-like constants, own static variables, ask for
  the calling scope's class or members, read a private member of the child,
  or when a used trait would supply the method instead of the parent.
  Upstream reports these; custos stays silent because both fixes would
  change behaviour.
- **Proxy fix — custos diverges** (F2). `never` is treated like `void`
  (returning from a `never` method does not compile), and the fix is
  withheld for by-reference returns and bodies reading the actual argument
  list.
