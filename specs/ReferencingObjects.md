---
id: ReferencingObjects
group: Confusing constructs
kind: syntax
needs: []
php: { min: "", max: "" }
---

# ReferencingObjects

## Summary

Since PHP 5, object handles are passed and assigned by value of the handle:
the callee already works on the same object. Taking objects by reference
(`Foo &$x` parameters, `$x = &new Foo()`) is a PHP 4 leftover that only
confuses readers — unless the reference is really used to replace the
caller's variable.

## Detection

### By-reference object parameters

- **D1** A parameter of a named function or of a method (including abstract
  and interface methods, which have no body). Parameters of closures and
  arrow functions are not inspected.
- **D2** The parameter is declared by reference (`&` before the name) and
  has **no** default value.
- **D3** The parameter has a declared type (in the signature; docblocks do
  not count) and that type is **not** made up solely of these names:
  `string`, `int`, `float`, `bool`, `array`, `mixed`, `iterable`, `null`,
  `callable` (case-insensitive; a leading `\` is ignored, so `\string`
  counts as `string`; `?T` means `T|null`). Examples that pass D3: `Foo`,
  `?Foo`, `Foo|string`, `object`, `self`, `\DateTimeInterface`. Examples
  that fail D3: `string`, `?\string`, `int|null`, `mixed`, `iterable`,
  `callable`, `?callable`. A
  parameter without type fails D3.
- **D4** Usage check in the function body (if there is a body): for every
  variable node in the body, at any depth (nested closures and their `use`
  lists included), whose name equals the parameter name:
  - if its direct parent is an assignment (`=`, `=&`, or a compound
    assignment such as `.=`/`??=`) and the variable is **not** the assigned
    value (i.e. it is the assignment target) → the parameter is used as a
    real reference → no report;
  - if it is used as a logical operand → no report. "Logical operand" means,
    after climbing out of any parentheses: it is the whole condition of an
    `if`, `elseif`, `while` or `do … while`; or the operand of `!`; or an
    operand of `&&`, `||`, `and`, `or`; or the condition of a full ternary
    `$p ? … : …` (not of a short ternary `?:`).
  Otherwise (no such use, or no body at all) → report.
- **D3a** (custos, see Divergences) The type has no member other than
  `null` from the D3 list (`array|\ArrayAccess`, `Foo|string`,
  `Foo|int` fail): for such values the reference matters.
- **D2a** (custos, see Divergences) The parameter is not a promoted
  constructor property (no visibility or `readonly` modifier): a promoted
  by-reference parameter binds the property to the caller's variable.
- **D4b** (custos, see Divergences) Further uses that make the reference
  matter → no report. The parameter variable is:
  - a target inside a destructuring pattern at any nesting depth
    (`list($p, $q) = …`, `[[$p]] = …`, `foreach ($rows as [$p, $x])`);
    a variable that is only a pattern *key* is a read;
  - the key or value variable of a `foreach`;
  - the variable of a `catch` clause;
  - the operand of `++`/`--`;
  - captured by reference: the value of a reference assignment
    (`$alias = &$p`), a by-reference array item (`[&$p]`, `array(&$p)`),
    or a by-reference closure import (`use (&$p)`);
  - returned (`return $p;`) from a function or method that itself returns
    by reference (`function &pick(Foo &$p)`);
  - the base of an element access (`$p[] = 1`, `$p['k']['j'] = 1`,
    `$p['k']`) while its type admits `null` (`?Foo`, `Foo|null`): writing
    an element of a null turns the caller's variable into an array.

  For a by-reference **variadic** parameter (`Foo &...$rest`) any mention
  of the variable in the body is enough: its elements are references that
  survive array copies, so a write through any copy reaches the caller.
  Only a variadic that the body never mentions is reported.

  `unset($p)`, `global $p` and `static $p` are *not* such uses: they only
  re-bind the local name and never touch the caller's variable.
- **D4c** (custos, see Divergences) The body may write local variables
  without naming them → no report: it contains a variable variable
  (`$$name`), a call to the global `extract()`, an `include`/`require` or
  an `eval`.
- **D4a** (custos, see Divergences) The parameter is passed directly as an
  argument whose matching parameter is by reference (positional, variadic
  tail or named), or to a function, method or constructor that cannot be
  resolved → no report (the callee may re-assign it).
- **D6** (custos, see Divergences) No report for methods without a body
  (abstract and interface methods), trait methods, methods of anonymous
  classes, methods that override a method of an ancestor class, interface
  or trait (or whose class has an ancestor that cannot be resolved), and
  methods overridden by a known descendant: removing the `&` from one
  signature makes it incompatible with the others (fatal error).

### By-reference instantiation

- **D5** A `new` expression whose direct parent is an assignment by
  reference: `$x = &new Foo()`, `$x = & new Foo`, `$x =& new Foo(...)`
  (any whitespace between `=` and `&`). A parenthesised `new`
  (`$x = &(new Foo)`) is not matched.

## Exceptions (no report)

- **E1** Parameters with a default value (`Foo &$x = null`).
- **E2** Untyped or scalar/array/mixed/iterable/callable-typed reference
  parameters.
- **E3** Reference parameters that are assigned to (`$x = new Foo;`,
  destructuring, `foreach`/`catch` variable, `++`), captured by reference,
  or tested in a boolean context in the body; by-reference variadics the
  body mentions at all; nullable parameters used as the base of an element
  access; bodies that write variables by name (D4c).
- **E6** Promoted constructor parameters (`public Foo &$x`).
- **E4** Closures / arrow functions.
- **E5** Plain assignments `$x = new Foo()`.
- **E7** (custos, see Divergences) Code in a test context: the file name
  ends with `Test.php`, `Spec.php` or `.phpt` or the path has a
  `/Fixtures/` directory, or the enclosing class name ends with `Test` or
  its namespace has a `Tests` or `Test` segment. Fixtures there declare
  by-reference objects on purpose, for tests that compare the signature.

## Report

- D1–D4: range = the whole parameter, from the first character of its type
  (or of its attributes/modifiers, if any) to the end of the variable name,
  e.g. `Foo & $item`, `Foo &$item`, `Foo& $item`. Severity: warning.
  Message: `Objects are handed over by handle already; drop the '&' before
  '${name}'.`
- D5: range = the `new` expression (from `new` to the end of its argument
  list or class name). Severity: warning. Message: `Objects are handed over
  by handle already; assign the new instance without '&'.`

## Fix

- **F1** (parameter) Remove the `&` and make exactly one space separate the
  type from the variable name: replace the text between the end of the type
  and the start of `$name` with a single space.
  `Foo & $item` → `Foo $item`; `Foo &$item` → `Foo $item`;
  `Foo& $item` → `Foo $item`. Nothing else in the signature changes.
- **F2** (instantiation) Replace the text from the `=` up to the start of the
  `new` keyword with `= ` (an equals sign followed by one space):
  `$x = & new Foo()` → `$x = new Foo()`; `$x = &new Foo()` → `$x = new Foo()`;
  `$x =& new Foo` → `$x = new Foo`.

## Options

None.

## PHP versions

No gating. `= &new` is a parse error since PHP 7.0, but it is parsed and
reported at every level (upstream fixtures run at the test default, between
5.6 and 7.0).

## Examples

```php
<?php
$cache = & <warning descr="Objects are handed over by handle already; assign the new instance without '&'.">new ArrayObject([])</warning>;
$queue =&<warning descr="Objects are handed over by handle already; assign the new instance without '&'.">new SplQueue</warning>;
$plain = new SplStack();

function attach(<warning descr="Objects are handed over by handle already; drop the '&' before '$node'.">DOMNode &$node</warning>, array &$list, &$any, int &$n) {}
function detach(<warning descr="Objects are handed over by handle already; drop the '&' before '$node'.">?DOMNode& $node</warning>, ?string &$label, iterable &$it) {}

interface Visitor {
    public function visit(
        <warning descr="Objects are handed over by handle already; drop the '&' before '$tree'.">Tree & $tree</warning>,
        mixed &$extra
    );
}

function replace(DOMNode &$node) {
    $node = new DOMText('x');
}
function check(DOMNode &$a, DOMNode &$b, DOMNode &$c, DOMNode &$d = null) {
    if (($a)) {}
    return !$b || ($c ? 1 : 0);
}
$fn = function (DOMNode &$node) {};
```

```php
<?php
$cache = new ArrayObject([]);
$queue = new SplQueue;
$plain = new SplStack();

function attach(DOMNode $node, array &$list, &$any, int &$n) {}
function detach(?DOMNode $node, ?string &$label, iterable &$it) {}

interface Visitor {
    public function visit(
        Tree $tree,
        mixed &$extra
    );
}

function replace(DOMNode &$node) {
    $node = new DOMText('x');
}
function check(DOMNode &$a, DOMNode &$b, DOMNode &$c, DOMNode &$d = null) {
    if (($a)) {}
    return !$b || ($c ? 1 : 0);
}
$fn = function (DOMNode &$node) {};
```

## Divergences

- **E7 — custos diverges from upstream.** Upstream also reports test
  fixtures, and its fix rewrites signatures that their tests assert.
- Upstream deletes "the token before the name (after skipping one whitespace
  run)"; for a by-reference variadic `Foo &...$rest` that token is `...`,
  so the fix would drop the variadic marker instead of `&`. custos always
  removes the `&` itself (F1 applied to the `&` token, keeping `...`).
- **`callable` is not an object type — custos diverges from upstream** (D3).
  Upstream reports `callable &$cb` with its "objects are handed over by
  handle" message, but a callable may be a string or an array, for which the
  reference is not redundant in the way the message claims. custos treats
  `callable` like the other non-object pseudo-types. `object &$o` is still
  reported unless reassigned or tested in the body (D4 already covers the
  reassigned case, so no extra rule is needed for it).
- Compound assignments count as "written to" (D4) because upstream treats
  every assignment kind alike; custos does the same.
- **custos diverges — references that matter (D3a, D4a, D6).** Found on
  Drupal, Laravel and Nextcloud: upstream reports `array|\ArrayAccess
  &$context` with `$context['k'] = 1` (the reference matters for arrays),
  a by-reference parameter forwarded to another by-reference parameter
  (`$step->process($node)` whose implementation re-assigns `$node`), and
  interface/overriding methods whose fix produces "Declaration of
  Replace::process(Node &$node) must be compatible with
  Step::process(Node $node)" — a fatal error `php -l` does not catch.
  custos skips types with a scalar/array member, parameters forwarded by
  reference (or to unresolved callees), and methods that take part in an
  override relation, have no body, or live in traits or anonymous classes.
- **custos diverges — indirect writes and promoted properties (D2a, D4b).**
  Upstream only recognizes a plain assignment whose target is the parameter
  itself. A swap helper (`list($a, $b) = array($b, $a)`), a `foreach` or
  `catch` variable, an alias (`$r = &$a; $r = new Node`), an element write
  on a by-reference variadic, and a promoted `public Node &$node` all rely
  on the reference; removing the `&` there silently changes what the caller
  observes. custos does not report them. The same goes for an element
  write on a nullable parameter (`?Node &$a` with `$a[] = 1`) and for
  bodies that write variables by name (D4c: `extract()`, `$$name`,
  `include`, `eval`).

Recorded in `internal/testing/conformance/testdata/ea-divergences.json`.
