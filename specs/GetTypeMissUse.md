---
id: GetTypeMissUse
group: Control flow
kind: semantic
needs: [names, index]
php: { min: "", max: "" }
---

# GetTypeMissUse

## Summary
Comparing the result of `gettype($v)` with a type-name string is a roundabout
(and typo-prone) way of calling the dedicated `is_*()` predicate. The rule
suggests the predicate, and flags strings that `gettype()` can never return.

## Detection
- **D1** A call to the global function `gettype` with exactly one argument:
  written unqualified or as `\gettype(...)`, name compared case-insensitively
  as PHP compares function names (`GetType`, `\GETTYPE` match). A call that
  resolves to a user function — `Ns\gettype(...)`, or an unqualified call in
  a namespace that declares or imports its own `gettype` — is not matched.
- **D2** The call's **direct** parent is a binary expression (no parentheses
  in between) whose operator is one of `==`, `!=`, `===`, `!==` (`<>` is the
  same token as `!=`). The call may be either operand.
- **D3** The *other* operand is resolved to a single string literal `L`:
  - if it is itself a string literal, `L` is that literal;
  - otherwise, compute the set of its possible values (see "Value
    resolution" below), keep only the elements that are string literals, and
    if exactly one remains, that is `L` (even when non-literal values are also
    possible). Otherwise there is no report.
- **D4** Let `T` be `L`'s contents (between the quotes, as written; no
  escape decoding). The type map (case-sensitive):

  | `T`        | predicate     |
  |------------|---------------|
  | `boolean`  | `is_bool`     |
  | `integer`  | `is_int`      |
  | `double`   | `is_float`    |
  | `string`   | `is_string`   |
  | `array`    | `is_array`    |
  | `object`   | `is_object`   |
  | `resource` | `is_resource` |
  | `NULL`     | `is_null`     |

  - **D4a** `T` is in the map → report the binary expression (M1, with fix).
  - **D4b** `T` is not in the map and is neither `unknown type` nor
    `resource (closed)` → report `L` as an invalid type name (M2, no fix).
    Note this includes wrong-case spellings such as `null`, `bool`, `int`,
    `float`, `Integer`.

### Value resolution (D3, non-literal operand)
Parentheses around the expression are stripped first. Each expression is
visited at most once (cycle guard).
- Ternary `a ? b : c` (and short `a ?: c`): union of the possibilities of the
  true and false branches.
- `a ?? b`: union of both operands.
- Variable `$v`: only inside a function, method or closure (at file/global
  scope nothing is resolved → no report). Union of: the parameter `$v`'s
  default value (if it is a parameter with a default), and the value of every
  plain `=` assignment to `$v` anywhere in that function body (including
  nested blocks and closures); for chained assignments `$v = $w = X` the
  innermost value `X` is used. **Unstable variable** (custos refinement, see Divergences): if `$v` is also the operand of `++`/`--` or the target of a compound assignment (`+=`, `.=`, `??=`, …) anywhere in that scope body, the whole result is *unknown* → no report.
- Property fetch `$o->p` / static property: if it resolves to a declared
  property, its default value (unless the default's text ends with the
  property name), plus plain assignments whose target has the same text
  inside the current function and inside the class constructor.
- Class constant `C::K`: the resolved constant's value.
- Global constant `K` (not `true`/`false`/`null`): the value of the resolved
  `define()` / `const` declaration.
- Anything else: the expression itself.
Possibilities are resolved recursively with these same rules.

## Exceptions (no report)
- **E1** `unknown type` and `resource (closed)` (valid `gettype()` results
  without an `is_*` counterpart).
- **E2** `gettype()` with zero or several arguments.
- **E3** The call wrapped in parentheses (`(gettype($v)) === 'array'`) or
  compared with another operator (`<`, `<=>`, `instanceof`, ...).
- **E4** The other operand resolves to zero or to several distinct string
  literals (e.g. a ternary over two strings, a `switch (gettype(...))`).

## Report
- **M1** (D4a)
  - Range: the whole binary comparison expression (both operands and the
    operator), excluding surrounding parentheses.
  - Severity: warning (rule default).
  - Message (our wording): `Use '{replacement}' instead.` with the F1 text.
- **M2** (D4b)
  - Range: the string literal `L` including its quotes. When `L` was found by
    value resolution it is the literal at its own location (e.g. inside an
    assignment or a constant declaration), possibly elsewhere in the file.
  - Severity: **error**, regardless of the rule's configured severity.
  - Message (our wording): `gettype() never returns '{T}'.`

## Fix
- **F1** (M1 only) Replace the binary expression with
  `[!]predicate(ARG)`: a leading `!` when the operator is `!=` or `!==`,
  the predicate from the D4 map, and `ARG` the verbatim source text of the
  `gettype()` argument. No parentheses are added.
  `gettype($v) === 'double'` → `is_float($v)`;
  `'NULL' != gettype($v)` → `!is_null($v)`.
  The predicate is written `\predicate` when an unqualified call to it at
  the reported position would not reach the global function (a `use function` import under that name, or a same-named function declared in the current namespace):
  with `use function Polyfill\is_int;`, `gettype($x) === 'integer'` →
  `\is_int($x)`.
- M2 has no fix.

## Options
None.

## PHP versions
No gating.

## Examples

```php
<?php
function kinds($item, $mode = 'array') {
    $a = <warning descr="Use 'is_int($item->qty)' instead.">gettype($item->qty) === 'integer'</warning>;
    $b = <warning descr="Use '!is_bool($item)' instead.">'boolean' !== gettype($item)</warning>;
    $c = <warning descr="Use 'is_null($item)' instead.">gettype($item) == 'NULL'</warning>;
    $d = <warning descr="Use '!is_float($item)' instead.">gettype($item) != 'double'</warning>;
    $e = <warning descr="Use 'is_array($item)' instead.">gettype($item) === $mode</warning>;
    $f = gettype($item) === <error descr="gettype() never returns 'int'.">'int'</error>;
    $g = gettype($item) === <error descr="gettype() never returns 'null'.">'null'</error>;

    $h = gettype($item) === 'unknown type';
    $i = gettype($item) !== 'resource (closed)';
    $j = (gettype($item)) === 'string';
    $k = gettype($item) === ($item ? 'string' : 'object');
    return [$a, $b, $c, $d, $e, $f, $g, $h, $i, $j, $k];
}
```

```php
<?php
function kinds($item, $mode = 'array') {
    $a = is_int($item->qty);
    $b = !is_bool($item);
    $c = is_null($item);
    $d = !is_float($item);
    $e = is_array($item);
    $f = gettype($item) === 'int';
    $g = gettype($item) === 'null';

    $h = gettype($item) === 'unknown type';
    $i = gettype($item) !== 'resource (closed)';
    $j = (gettype($item)) === 'string';
    $k = gettype($item) === ($item ? 'string' : 'object');
    return [$a, $b, $c, $d, $e, $f, $g, $h, $i, $j, $k];
}
```

## Divergences
- **Unstable variables — custos refinement, not upstream.** Value discovery ignores `++`/`--` and compound assignments upstream, so a variable later incremented or extended is analysed with its initial value only. custos makes the result unknown (no report), as in the shared value discovery of `CallableMethodValidity`. No upstream fixture relies on such a variable; recorded in `docs/decisions.md` ("Spec-level false positives").
- **Resolved values (upstream quirk).** When the compared operand is a
  variable/constant, the fix still replaces the comparison with a fixed
  predicate derived from one possible value, and "exactly one string literal
  among the possibilities" is accepted even if other non-string values are
  possible (e.g. `$t = 'array'; if ($x) { $t = getKind(); }`). The M2 error
  may then land on a literal far from the comparison. Recommendation: for
  conformance only literal operands are exercised by upstream fixtures; an
  implementation may start with literal operands plus local-variable
  resolution, and should only report through resolution when **all**
  possibilities are the same string literal. Document any narrowing as a
  deliberate divergence.
- **Function name (custos diverges):** upstream matches the last name
  segment case-sensitively and ignores the namespace, so `GETTYPE($v)` was
  missed while a namespaced user `gettype` (`App\gettype($v)`, or an
  unqualified call where the namespace declares one) was reported and
  rewritten to `is_*()`. custos matches case-insensitively and only calls
  resolving to the global function (D1).
- **Builtin spelling (custos diverges).** Upstream always inserts the bare
  predicate name, which a namespaced or imported function of that name
  captures. custos writes `\predicate(` in that case (F1).
