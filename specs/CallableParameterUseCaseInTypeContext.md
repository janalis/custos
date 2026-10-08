---
id: CallableParameterUseCaseInTypeContext
group: Architecture
kind: semantic
needs: [names, types, hierarchy, index, flow, stubs]
php: { min: "", max: "" }
---

# CallableParameterUseCaseInTypeContext

## Summary
A parameter's declared/documented type is a contract. Testing it with an
`is_*()` function that can never succeed (or never fail) for that type is dead
logic, and re-assigning the parameter a value of an unrelated type breaks the
contract for every later reader of the variable.

## Detection
Applies to every named function and method (abstract/interface methods have
no body and thus no uses). Closures and arrow functions are not inspected as
owners of parameters.

- **D0** Skip the function/method entirely in a test context: file path ends
  with `Test.php`, `Spec.php` or `.phpt`, or contains `/Fixtures/`; or the
  enclosing class FQN ends with `Test`, or contains `\Tests\` or `\Test\`.

### Parameter type set `P`
For each parameter:
- **D1** Raw types = union of: the declared type (signature), the `@param`
  docblock types for that parameter, and the type of the default value (if
  any; `= null` → null, `= []` → array, `= ''` → string, …). For a variadic
  parameter (`...$xs`, typed or not) the raw type set is just `{array}`.
  Unknown/unresolvable entries are dropped.
- **D2** Normalise each raw type with N (below). Build `P`:
  - `mixed` or `object` present → skip this parameter entirely;
  - `callable` → add `callable`, `array`, `string` and the class name
    `\Closure` (literal, not re-normalised);
  - `iterable` → add `iterable`, `array` and the class name `\Traversable`;
  - anything else → add the normalised name.
- **D3** `P` empty → skip the parameter.
- **D4** `P` is exactly `{null}` and the default value is the `null`
  constant → skip the parameter.

**N — normalisation** of one type name: if it contains `[]` → `array`.
Otherwise look up its lowercase form, with or without a leading `\`:
`array`; `iterable`; `string`; `bool`/`boolean`/`true`/`false` → `bool`;
`int`/`integer` → `int`; `float`; `number`; `null`; `void`; `mixed`;
`callable`; `closure` (i.e. `\Closure` in any case) → `callable`; `resource`;
`static`/`$this` → `static`; `self`; `object`. Any other name is a class name
and is kept as its FQN with leading `\` and original case (e.g.
`\App\User`).

### Uses of the parameter
- **D5** Consider every access (read or write) to the parameter's variable
  that is reachable in the function body's control flow from the entry,
  in source order — **all** of them, not only the first one per path.
  Accesses inside nested closures/arrow functions/classes are not part of
  this function's flow and are ignored. Each access is analysed with D6/D7.
  A "class name" in `P` below means an entry starting with `\` other than
  exactly `\Closure`, or the entries `self`/`static`.

### D6 — `is_*` checks
- **D6a** The access is directly an argument (any position, not wrapped in
  parentheses or other expressions) of a plain function call (not a method
  or static call) that resolves to one of the global functions in the table
  below. Names compare case-insensitively (`IS_INT`, `\Is_Int` count); a call
  resolving to a same-named namespaced function (declared in the current
  namespace or imported with `use function`) does not count.
- **D6b** The check is *plausible* when:
  | function | plausible if `P` contains |
  |---|---|
  | `is_array` | `array` or `iterable` |
  | `is_string` | `string` |
  | `is_bool` | `bool` |
  | `is_int` | `int` or `number` |
  | `is_float` | `float` or `number` |
  | `is_resource` | `resource` |
  | `is_numeric` | — if `P` contains `string` the access is skipped; otherwise plausible if `number`, `float` or `int` |
  | `is_callable` | `callable`, `array`, `string` or `\Closure` |
  | `is_object` | `object`, `callable`, or any class name |
  | `is_a` | `object`, `string`, or any class name |
  Any other function (including aliases `is_integer`, `is_long`,
  `is_double`) → the access is ignored.
- **D6c** Not plausible → report the call. The message depends on whether
  the call's direct parent is a `!` negation (no parentheses in between):
  negated → "always true" message, otherwise "always false" message.

### D7 — re-assignments
- **D7a** The access is the target of a plain assignment `$p = value` (or
  `$p = &value`); compound assignments (`.=`, `+=`, `??=` …) and
  destructuring are not considered. The assignment's target must be the
  parameter variable itself.
- **D7b** `R` = normalised (N) types of `value`, unknowns dropped, computed
  with the expression type rules **T** of `specs/UnnecessaryCasting.md`
  (section "Expression type rules (T-rules)", including the R-function
  overrides for `str_replace`, `parse_url`, `abs`, `microtime`, …), plus:
  `$this` has the enclosing class's FQN as type; a method call's type is its
  callee's declared return type ∪ `@return` types (`self`/`static`
  normalised to `self`/`static`); an unresolvable callee yields no type.
- **D7c** If `R` has 2 or more entries:
  - if `R` contains `string` or `array`:
    - if it contains `bool` and `value` is a plain function call → remove
      `bool`;
    - else if it contains `null` and `value` is a plain function call →
      remove `null`;
  - else if `R` contains `null` and `P` contains a class name → remove
    `null`.
- **D7d** Remove `mixed` from `R`.
- **D7e** For each `t` in `R` (iteration order unspecified; stop at the first
  violation, so at most one report per assignment):
  1. If `t` is `self` or `static`, try to translate it to a class:
     - let `X = value`; if `value` is `A ?? B` and `A` is equivalent to the
       assigned variable, let `X = B`;
     - if `X` is a method call (`->`, `?->` or `::`): if its receiver is a
       class name (`Foo::make()`), `t` = that class's FQN (if it resolves); else
       take the receiver expression's type, keep only class names (entries
       starting with `\`), and if there are 1 or 2 of them, `t` = one of them;
     - if `t` is still `self`/`static` → skip this `t` (no report for it).
  2. `t` is compatible when `t ∈ P`; or when `t` is a class name that
     resolves to at least one known class/interface/trait and the
     *closure* of `t` intersects the union of the closures of the class
     names in `P`. Closure of a class-like = itself, all ancestor classes,
     all interfaces implemented/extended transitively, all traits used
     transitively by any of them. (So a parent class assigned to a child-typed
     parameter is compatible, since both closures contain the parent.) When
     `t` does not resolve to a known class-like, or any class-like named in
     its closure (parent, interface, trait, at any depth) does not resolve,
     compatibility is unknown: treat `t` as compatible (no report). A
     non-class `t` not in `P` is incompatible.
  3. Incompatible → report `value` with the type `t` in the message; stop.

## Exceptions (no report)
- **E1** Parameters typed or documented as `mixed` or `object`, untyped
  parameters without default value (empty `P`), and `= null` defaults with no
  other type.
- **E2** `is_*` calls that agree with `P` (e.g. `is_array` on `iterable`,
  `is_object`/`is_string`/`is_array` on `callable`, `is_int` on `number`),
  `is_numeric` on a string-capable parameter, unrelated functions.
- **E3** Assignments whose value type is unknown, or `mixed` only.
- **E4** Core functions returning `string|false`, `string|null`, `array|false`
  … (D7c) assigned to a string/array parameter; nullable object values
  assigned to object-typed parameters.
- **E5** Values whose class shares an ancestor or interface with the
  parameter's class (D7e.2); values whose class, or part of whose
  hierarchy, is unknown to the index (`new Unknown()`, a class extending an
  unknown base).
- **E6** Uses in unreachable code or in nested closures.
- **E7** Test contexts (D0).

## Report
- D6: range = the whole function call (`is_int($p)`), without a preceding
  `!`. Severity: warning. Messages:
  - not negated: `This check is always false for the declared parameter
    type; is the parameter being reused?`
  - negated: `This check is always true for the declared parameter type; is
    the parameter being reused?`
- D7: range = the assigned value expression (right-hand side only, without
  the `=`). Severity: warning. Message: `Assigning a value of type {t} does
  not match the parameter's declared type.`

## Fix
None.

## Options
None.

## PHP versions
No gating. All syntax (nullable/union types, `??`, variadics) is parsed at
every level; upstream fixtures run at the test default (between 5.6 and 7.0).

## Examples

```php
<?php
/** @param int[] $ids */
function pick(array $rows, iterable $feed, $ids, callable $cb, float ...$nums) {
    $ok = is_array($rows) && is_array($feed) && is_callable($cb) && is_object($cb);
    $bad = <warning descr="This check is always false for the declared parameter type; is the parameter being reused?">is_string($rows)</warning>
        || <warning descr="This check is always false for the declared parameter type; is the parameter being reused?">is_bool($feed)</warning>
        || !<warning descr="This check is always true for the declared parameter type; is the parameter being reused?">is_float($ids)</warning>
        || <warning descr="This check is always false for the declared parameter type; is the parameter being reused?">is_float($nums)</warning>;
    return [$ok, $bad];
}

/**
 * @param mixed  $any
 * @param number $qty
 */
function loose($any, $qty, string $code) {
    return [is_int($any), is_float($qty), is_numeric($code), is_long($code)];
}

interface Shape {}
class Circle implements Shape {}
class Ring extends Circle {}
class Label {}

function redraw(Shape $s, Ring $r = null, int $n = 0, string $txt = '', array $list = []) {
    $s = $s ?? new Circle();
    $r = $r ?? new Circle();
    $r = <warning descr="Assigning a value of type \Label does not match the parameter's declared type.">new Label()</warning>;
    $n = 2 * $n + 1;
    $n = <warning descr="Assigning a value of type float does not match the parameter's declared type.">$n / 2.5</warning>;
    $txt = str_replace('a', 'b', $txt);
    $txt = substr($txt, 1);
    $list = explode(',', $txt);
    $list = <warning descr="Assigning a value of type string does not match the parameter's declared type.">'none'</warning>;
    $s = <warning descr="Assigning a value of type null does not match the parameter's declared type.">null</warning>;
    $txt .= 5;
}

function untyped($flag = null) {
    $flag = 'on';
}
```

## Divergences
- **Flow-aware variables and `?:` (custos diverges):** inherited from the
  shared T-rules (see UnnecessaryCasting): variable types are narrowed by
  the guards on the path (`if (null !== $next) { $p = $next; }` assigns a
  non-null value), and the left part of `?:` never contributes `false`
  (`$p = preg_split(…) ?: [];` assigns an array).
- **`preg_*` replacement results may be null (custos diverges):** inherited
  from the shared T-rules (see UnnecessaryCasting): `preg_replace`,
  `preg_replace_callback`, `preg_filter` and `preg_replace_callback_array`
  include `null` in their result type.
- **Function-name matching (custos diverges from upstream).** Upstream
  matches the written last name segment case-sensitively, so `IS_STRING($p)`
  is skipped while a namespaced user `is_string()` is checked. custos
  resolves the call: any case of the global function counts, a same-named
  namespaced function does not.
- The reported type in the D7 message depends on set iteration order when
  several types are incompatible (e.g. `abs($str)` → `int|float` against a
  `string` parameter). Only the range and severity are compared, so any
  incompatible member may be named.
- Upstream depends on the IDE's full type engine; where the T-rules do not
  determine a type, results may differ. Recommendation: when in doubt, prefer
  "unknown" (no report).
- **Unresolvable classes (custos diverges):** upstream treats a value class
  it cannot resolve as incompatible, so `$shape = new UnknownClass();` on a
  `Shape` parameter is reported although the unknown class may well
  implement `Shape` (a class from a missing dependency, a generated class).
  custos treats an unresolvable class, or a class whose hierarchy is only
  partly known, as unknown and stays silent (D7e.2).
- **Sound arithmetic (custos diverges):** the arithmetic T-rule's
  heuristic types `int + <unknown>` (and string operands) as float, so
  `$ttl = time() + $lifeTime;` with an untyped `$lifeTime` was reported
  against an `int` parameter. custos uses the sound arithmetic typing of
  UnnecessaryCasting here too: an operand of unknown type gives an unknown
  result (no report); numeric strings give `int|float`.
- **`mb_convert_encoding()` (custos diverges):** inherited from the shared
  T-rules (see UnnecessaryCasting): a string input yields `string|false`,
  so `$text = mb_convert_encoding($text, 'UTF-8');` on a `string`
  parameter is not reported as assigning an array.
- **`mixed` in a value type (custos diverges).** D7d removes `mixed` from
  `R` and checks what is left, so `$p = func_get_arg(2)` (stub type
  `mixed|false`) is reported as assigning a `bool` to a `string|null`
  parameter, although `mixed` already admits any type. custos treats a
  value type containing `mixed` as unknown (E3): no report.
- **Untyped parameters (custos diverges).** A parameter without a declared
  type and without a `@param` tag accepts any value; its default value is
  only one of them (`$extraWhere = false` later assigned a string, common in
  Matomo and PrestaShop). custos skips such parameters instead of taking
  the default's type as P (D1). EA case affected: `parameter-types-checks.php`
  (one fewer report, listed divergence).
