---
id: InArrayMissUse
group: Performance
kind: syntax
needs: []
php: { min: "", max: "" }
---

# InArrayMissUse

## Summary
Two wasteful `in_array()` shapes:
- searching the keys list built by `array_keys($a)` — a direct key lookup
  (`array_key_exists`) avoids building the list;
- searching a one-element array literal — that is just a comparison with the
  single element.

## Detection
Common preconditions:
- **D1** A function call (not a method/static call) that resolves to the
  global function (names compared case-insensitively, as PHP does; `\` and a
  global `use function` import are fine, but a same-named function declared in
  the current namespace, one imported from another namespace, or a qualified
  non-global name such as `Ns\in_array` does not count): `in_array`.
- **D2** The call has **2 or 3** arguments. Let `N` = 1st argument (needle),
  `H` = 2nd argument (haystack), `S` = optional 3rd argument.

### Pattern K — keys lookup
- **D3** `H` is a function call (not a method call) that resolves to the
  global `array_keys` (case-insensitive, same resolution as D1) and which has **exactly 1**
  argument `A`. (`array_keys($a, $v)` with a search value does not qualify.)
- `S` is ignored entirely (strict or not, present or not).
- Report kind K on the `in_array` call.

### Pattern C — single-element haystack
Checked only when pattern K's first test fails (i.e. `H` is not a function
call named `array_keys`).
- **D4** `H` is an array literal (`[...]` or `array(...)`) with **exactly one**
  element. Let `V` be that element's value: for a `key => value` element the
  value part (the key is ignored), otherwise the element itself.
- **D5** Determine the *target* `T` and the polarity:
  - default: `T` = the call, polarity *exists*;
  - if the call's direct parent (no parentheses in between) is a logical not
    `!`: `T` = the `!` expression, polarity *missing*;
  - if the call's direct parent is a binary expression whose other operand is
    a boolean constant (`true`/`false`, any case, optionally `\`-qualified):
    - operator `==` or `===`: `T` = the binary expression, polarity *exists*
      when the constant is `true`, *missing* when it is `false`;
    - operator `!=`, `<>` or `!==`: `T` = the binary expression, polarity
      *missing* when the constant is `true`, *exists* when it is `false`;
    - any other operator (`&&`, `||`, `and`, …): `T` = the call, *exists*.
  - a binary parent whose other operand is not a boolean constant leaves the
    default (`T` = the call, *exists*).
- **D6** Strictness: strict when the option `FORCE_STRICT_COMPARISON` is on,
  or when `S` is present and is the constant `true` (any case). Otherwise
  loose (including `S` = `false`, `1`, a variable…).
- Report kind C on `T`.

## Exceptions (no report)
- **E1** Haystack array literal with zero elements or two or more.
- **E2** Haystack that is a variable, constant, method call, or a function
  call other than single-argument `array_keys`.
- **E3** 0, 1 or 4+ arguments.

## Report
- Range:
  - kind K: the `in_array(...)` call only (a surrounding `!` is outside the
    range and kept);
  - kind C: `T` exactly — the call, or the `!` expression (from `!` to the
    call's `)`), or the whole binary comparison (left operand start to right
    operand end).
- Severity: warning.
- Messages:
  - K: `Look the key up directly with '{replacement}'.`
  - C: `Compare directly: '{replacement}'.`

## Fix
- **F1 (kind K)** Replace the `in_array` call with
  `array_key_exists({N}, {A})` — verbatim texts of the needle and of the
  `array_keys` argument, separator `, `, bare function name (no qualifier) —
  except that it is written `\array_key_exists` when an unqualified call at
  that position would not reach the global function (a `use function` import under that name, or a same-named function declared in the current namespace).
  A surrounding `!` or comparison is preserved as is:
  `!in_array($id, array_keys($map))` → `!array_key_exists($id, $map)`.
- **F2 (kind C)** Replace `T` with a comparison:
  - operator: `==` for *exists*, `!=` for *missing*; append `=` when strict
    (`===` / `!==`).
  - left text `{n}`: verbatim text of `N`, wrapped in `(`…`)` when `N` is a
    binary expression (any binary operator, including `??`, `.`, `+`, `&&`,
    `instanceof`) or a ternary (including `?:`). Other expressions are used
    as is.
  - `{v}`: verbatim text of `V`, never wrapped.
  - Regular comparison style (default): `{n} {op} {v}`; Yoda style:
    `{v} {op} {n}`. Single spaces around the operator.
  Examples (regular): `in_array($c, ['x'], true)` → `$c === 'x'`;
  `false == in_array($c, ['x'])` → `$c != 'x'`;
  `in_array($a ?? 0, [7]) !== true` → `($a ?? 0) != 7`.

## Options
| Option | Type | Default | Effect |
|---|---|---|---|
| `FORCE_STRICT_COMPARISON` | bool | `false` | When on, pattern C always produces `===`/`!==`, whatever the 3rd argument. The EA fixture runs with it off. |

Comparison style (`regular`/`yoda`) is the global setting.

## PHP versions
No gating. The EA test runs at the PhpStorm default level (between 5.6 and 7.0).

## Examples

```php
<?php
function checks($role, $list, $map, $a, $b)
{
    $r = [];
    $r[] = <warning descr="Compare directly: '$role == 'admin''.">in_array($role, ['admin'])</warning>;
    $r[] = <warning descr="Compare directly: '$role === 'admin''.">in_array($role, array('admin'), TRUE)</warning>;
    $r[] = <warning descr="Compare directly: '$role == 'admin''.">in_array($role, ['k' => 'admin'], false)</warning>;
    $r[] = <warning descr="Compare directly: '$role != 'admin''.">!in_array($role, ['admin'])</warning>;
    $r[] = <warning descr="Compare directly: '$role !== 'admin''.">in_array($role, ['admin'], true) == false</warning>;
    $r[] = <warning descr="Compare directly: '$role == 'admin''.">false !== in_array($role, ['admin'])</warning>;
    $r[] = <warning descr="Compare directly: '$role != 'admin''.">true != in_array($role, ['admin'])</warning>;
    $r[] = $b || <warning descr="Compare directly: '$role == 'admin''.">in_array($role, ['admin'])</warning>;
    $r[] = <warning descr="Compare directly: '($a . $b) == 'ab''.">in_array($a . $b, ['ab'])</warning>;
    $r[] = <warning descr="Compare directly: '($a ? $b : 0) === 3''.">in_array($a ? $b : 0, [3], true)</warning>;
    $r[] = !<warning descr="Look the key up directly with 'array_key_exists($role, $map)'.">in_array($role, array_keys($map), true)</warning>;
    $r[] = <warning descr="Look the key up directly with 'array_key_exists('id', $map)'.">\in_array('id', array_keys($map))</warning>;

    $r[] = in_array($role, []);
    $r[] = in_array($role, ['admin', 'owner']);
    $r[] = in_array($role, $list);
    $r[] = in_array($role, array_keys($map, 1));
    return $r;
}
```

```php
<?php
function checks($role, $list, $map, $a, $b)
{
    $r = [];
    $r[] = $role == 'admin';
    $r[] = $role === 'admin';
    $r[] = $role == 'admin';
    $r[] = $role != 'admin';
    $r[] = $role !== 'admin';
    $r[] = $role == 'admin';
    $r[] = $role != 'admin';
    $r[] = $b || $role == 'admin';
    $r[] = ($a . $b) == 'ab';
    $r[] = ($a ? $b : 0) === 3;
    $r[] = !array_key_exists($role, $map);
    $r[] = array_key_exists('id', $map);

    $r[] = in_array($role, []);
    $r[] = in_array($role, ['admin', 'owner']);
    $r[] = in_array($role, $list);
    $r[] = in_array($role, array_keys($map, 1));
    return $r;
}
```

Yoda style: `in_array($role, ['admin'], true)` → `'admin' === $role`.

## Divergences
- **Precedence when `T` is the bare call inside an operator:** the call is
  replaced by an unparenthesised comparison. Fine under `&&`/`||`, but under
  a non-boolean binary parent (`in_array($x, [1]) == $y`, `$k . in_array(...)`)
  or a cast/unary other than `!` the result changes meaning or is invalid
  (`==` is non-associative). Recommendation: wrap the replacement in
  parentheses when the parent operator binds tighter than or equal to `==`;
  keep the bare form under `&&`, `||`, `and`, `or`, `xor`, `?:`, assignment,
  argument and statement contexts (matches all fixtures).
- **Needle/element precedence:** only binary/ternary needles are wrapped;
  assignment, `print`, `yield`, `include`, arrow-fn needles and any
  low-precedence element `V` (e.g. `[$a ?: 1]`, `[$x = 2]`) produce wrong
  code. Recommendation: also parenthesise those. Not covered upstream.
- **Spread element:** `in_array($x, [...$list])` would be treated as a
  one-element literal and rewritten to `$x == ...$list` (invalid).
  Recommendation: skip haystacks whose single element is a spread or a
  by-reference element.
- **Loose comparison semantics:** `in_array` with loose mode and a single
  element is equivalent to `==`, so the rewrite is exact; no divergence.
- **Function-name matching — custos diverges from upstream.** Upstream matches
  `in_array`/`array_keys` by the name as written (case-sensitive, any
  namespace qualifier, no resolution), so a differently cased call such as
  `In_Array($x, ['a'])` is missed while a namespaced or imported user function
  of the same name is reported (and rewritten) as if it were the builtin.
  custos matches case-insensitively and only calls that reach the global
  function.
- **Builtin spelling (custos diverges).** Upstream inserts a bare `array_key_exists(` in its fix, so a function of that name declared in (or imported into) the namespace captures the rewritten call. custos writes `\name(` in that case (F1).
- **Key lookups (custos diverges).** `in_array($x, array_keys($a))` and
  `array_key_exists($x, $a)` differ: `'5'` is stored as the int key 5 (a
  strict search for `'5'` misses it, array_key_exists() finds it), loose
  comparison matches numeric strings (`'1.0' == 1`) and, before PHP 8,
  `'abc' == 0`. Pattern K keeps the report but offers the fix only for an
  int-typed needle with strict comparison, a string literal that is not a
  canonical integer with strict comparison, or a non-numeric string
  literal with loose comparison on PHP 8+. EA case affected:
  `in-array-misuse.php` (listed divergence).
