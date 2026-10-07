---
id: ArraySearchUsedAsInArray
group: Control flow
kind: syntax
needs: []
php: { min: "", max: "" }
---

# ArraySearchUsedAsInArray

## Summary
`array_search()` returns a key (which may be `0` or `''`) or `false`. When its
result is only used as a yes/no answer, `in_array()` expresses the intent
directly and avoids the classic "key 0 is falsy" trap. Comparing its result
with `true` is always pointless, since the function never returns `true`.

## Detection
D1. Node: a plain function call (not a method/static call) whose name part is
    `array_search`, compared case-insensitively as PHP compares function names
    (`ARRAY_SEARCH`, `Array_Search` match) that resolves to the global
    built-in: unqualified (in a namespace only when no function of that name
    is declared there and no `use function` imports another one) or
    `\array_search`; `Ns\array_search(...)` and shadowing user functions are
    not reported. At least two arguments are required. The fix
    writes the lower-case `in_array`.

D2. *Logical-operand case.* Starting from the call, skip outwards through any
    number of enclosing parentheses; let *P* be the first non-parenthesis
    ancestor and *S* the node directly below it (the call or its outermost
    wrapping parentheses). The call is a logical operand when:
    - *P* is an `if` or `elseif` (the call is its condition), a `while`
      condition, or a `do … while` condition; or
    - *P* is a logical-not unary `!`; or
    - *P* is a binary expression with operator `&&`, `||`, `and` or `or`
      (keyword operators case-insensitive; the call may be either operand;
      `xor` does **not** count); or
    - *P* is a full ternary `c ? a : b` and *S* is its condition `c`.
    Report the **call** (D2 report, fixable, default severity).

D3. *Comparison case* — only evaluated when D2 does not apply. The call's
    **direct** parent (no parentheses in between) is a binary expression with
    operator `===` or `!==`, and the *other* operand is the constant `true` or
    `false` (case-insensitive name; a `\`-qualified constant counts too). The
    call may be left or right operand.
    - D3a. Other operand is `false`: report the **whole binary expression**
      (fixable, default severity).
    - D3b. Other operand is `true`: report the **`true` constant** only, with
      severity *error*, no fix.

## Exceptions (no report)
E1. Fewer than two arguments (`array_search($needle)`, `array_search(...$args)`).
E2. Short ternary / null-coalescing: `array_search(...) ?: $d`,
    `$a ?: array_search(...)`, `$a ?? array_search(...)`, and the call used as
    a ternary branch (`$c ? array_search(...) : 1`).
E3. Loose comparisons (`==`, `!=`, `<>`) with booleans, comparisons with
    anything other than a boolean constant (`=== 0`, `!== null`, `=== $f`).
E4. Comparison with parentheses around the call: `(array_search($a, $b)) === false`
    (the binary is not the direct parent). Unless D2 applies through other
    parents, nothing is reported.
E5. `for (...; array_search(...); ...)` conditions, assignments, arguments,
    return values, `xor` operands — not logical-operand contexts.
E6. Methods and static methods named `array_search`.

## Report
- D2: range = the function call, from the start of its name (including a
  leading `\` / namespace qualifier) to its closing `)`. Severity: warning.
  Message: `Use 'in_array(...)' to test membership.`
- D3a: range = the whole `===`/`!==` binary expression (left operand start to
  right operand end). Severity: warning. Same message as D2.
- D3b: range = the `true` constant token. Severity: **error** (fixture markup
  `<error>`). Message: `array_search() cannot return true; this comparison never changes.`

## Fix
F1. (D2) Rename the call: replace only the function-name identifier
    `array_search` with `in_array`. Namespace qualifier, argument list (all
    arguments, including a third "strict" argument), surrounding parentheses
    and operators are kept verbatim.
    `if (!array_search($k, $map))` → `if (!in_array($k, $map))`.

F2. (D3a) Replace the whole binary expression:
    - operator `!==` → the renamed call alone: `in_array(<args>)`;
    - operator `===` → `!` immediately followed by the renamed call:
      `!in_array(<args>)` (no space, no extra parentheses).
    The call text (with F1 renaming applied, qualifier preserved) is taken from
    the operand that is not the constant, so `false === array_search($a, $b)`
    → `!in_array($a, $b)`. Whitespace/comments inside the comparison are lost.

F3. D3b has no fix.

Builtin spelling (F1, F2): the emitted name is `\in_array` when the call
was written `\array_search`, or when an unqualified `in_array` at that
position would not reach the global function (a `use function` import under that name, or a same-named function declared in the current namespace); otherwise `in_array`.
In `namespace Acl; function in_array() {…}`, `array_search($k, $m) !== false`
→ `\in_array($k, $m)`.

## Options
None.

## PHP versions
No gating.

## Examples

```php
<?php
function audit(array $roles, $who, $flag) {
    if (<warning descr="Use 'in_array(...)' to test membership.">array_search($who, $roles)</warning>) { log_hit(); }
    while (!<warning descr="Use 'in_array(...)' to test membership.">\array_search($who, $roles, true)</warning>) { $who = next_user(); }
    $ok = ((<warning descr="Use 'in_array(...)' to test membership.">array_search('root', $roles)</warning>)) && $flag;
    $tag = <warning descr="Use 'in_array(...)' to test membership.">array_search($who, $roles)</warning> ? 'known' : 'stranger';

    $miss = <warning descr="Use 'in_array(...)' to test membership.">array_search($who, $roles) === FALSE</warning>;
    $hit  = <warning descr="Use 'in_array(...)' to test membership.">false !== array_search($who, $roles)</warning>;
    $odd  = array_search($who, $roles) === <error descr="array_search() cannot return true; this comparison never changes.">true</error>;

    $key   = array_search($who, $roles) ?: 'none';
    $other = $flag ?? array_search($who, $roles);
    $loose = array_search($who, $roles) == false;
    $wrap  = (array_search($who, $roles)) !== false;
    $one   = array_search($who);
    return $flag xor array_search($who, $roles);
}
```

```php
<?php
function audit(array $roles, $who, $flag) {
    if (in_array($who, $roles)) { log_hit(); }
    while (!\in_array($who, $roles, true)) { $who = next_user(); }
    $ok = ((in_array('root', $roles))) && $flag;
    $tag = in_array($who, $roles) ? 'known' : 'stranger';

    $miss = !in_array($who, $roles);
    $hit  = in_array($who, $roles);
    $odd  = array_search($who, $roles) === true;

    $key   = array_search($who, $roles) ?: 'none';
    $other = $flag ?? array_search($who, $roles);
    $loose = array_search($who, $roles) == false;
    $wrap  = (array_search($who, $roles)) !== false;
    $one   = array_search($who);
    return $flag xor array_search($who, $roles);
}
```

## Divergences
- custos diverges: upstream matches the function name case-sensitively, so
  `ARRAY_SEARCH(...)` is missed. PHP function names are case-insensitive;
  custos matches any casing (D1).
- custos diverges: upstream matches the name regardless of namespace, so a
  namespaced user function `array_search` (`Lookup\array_search(...)`, or an
  unqualified call where `namespace Inventory` declares one) was reported and
  rewritten to the unrelated global `in_array`. custos requires the call to
  resolve to the global function (D1).
- The D3a `===` fix produces `!in_array(...)` without extra parentheses.
  Because the call must be a direct operand of `===`, the result keeps the
  same meaning in every position where the comparison could appear. No action.
- D2 ignores the call's semantics when the found key could legitimately be
  used as a truthy value; that is the point of the rule (key `0` is a bug
  source), so report it as upstream does.
- **Builtin spelling (custos diverges).** Upstream renames only the
  identifier, so a namespaced or imported `in_array` captures the rewritten
  call. custos qualifies the name with `\` in that case (Fix, "Builtin
  spelling").
