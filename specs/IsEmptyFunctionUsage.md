---
id: IsEmptyFunctionUsage
group: Code style
kind: semantic
needs: [types, index, hierarchy]
php: { min: "", max: "" }
---

# IsEmptyFunctionUsage

## Summary
`empty()` treats many unrelated values (`0`, `'0'`, `''`, `[]`, `null`,
`false`) as "empty". When the argument's type is known, a precise check is
clearer: `count(...) === 0` for countables, `... === null` for nullable
scalars/objects. Optionally, any other `empty()` use can be flagged.

## Detection
Node: an `empty(...)` construct.

Definitions:
- *S* (subject): the single argument with any number of wrapping parentheses
  removed.
- *inverted*: the **direct** parent of the `empty(...)` node is a logical-not
  unary expression `!empty(...)` (a parenthesized `!(empty($x))` is not
  inverted).
- *Types(S)*: the inferred type set of *S* (native declarations, `@param`,
  `@var`, `@return` docblocks, function return types, …), with unknown parts
  dropped, each member normalised: any `X[]` form → `array`;
  `boolean`/`true`/`false` → `bool`; `integer` → `int`; leading `\` on core
  type names removed; class types stay fully qualified with a leading `\`.

D0. Pre-checks: if the construct has exactly one argument and *S* is an array
    element access (`$a[0]`, `$a['k']`, `$o->list[1]`), stop — **no report of
    any kind**. If there is not exactly one argument, skip D1/D2 and go to D3.

D1. **Count suggestion** (only when `SUGGEST_TO_USE_COUNT_CHECK` is on):
    *Types(S)* is non-empty and **every** member is either `array` or a class
    type that resolves to a known class/interface which is, implements or
    extends `\Countable` (anywhere in its inheritance tree). Any `null`,
    scalar, unresolvable class or non-countable class in the set fails D1.
    → report with the count fix; stop.

D2. **Null-comparison suggestion** (only when `SUGGEST_TO_USE_NULL_COMPARISON`
    is on). Applies when either:
    - D2a. (`SUGGEST_NULL_COMPARISON_FOR_SCALARS` on) *Types(S)* has exactly two
      members, one is `null` and the other is one of `int`, `float`, `bool`,
      `resource`; or
    - D2b. after ignoring `null`, *Types(S)* has at least one member and **all**
      remaining members are class types (start with `\`). A non-nullable single
      class type qualifies too.
    When D2 applies:
    - Property-access guard: descend from *S* repeatedly to its first child
      that is a node (not a token); if *S* itself or any node on that descent is
      a property access (instance `->p` / `?->p`, or static `X::$p`), stop with
      **no report of any kind**. Examples suppressed: `$o->p`, `$o->p->q`,
      `$o->m()->p`, `$o->p->m()`, `$list[0]->p` (descent covers the object
      side of member accesses). The descent stops at an argument list: when
      the first child is a call's argument list (function call, static call
      on a class name, `new`), the guard does not apply, so
      `empty(make($o->p))` is not suppressed.
    - Otherwise → report with the null fix; stop.

D3. **Generic report** (only when `REPORT_EMPTY_USAGE` is on): reached when
    D1/D2 did not report and did not stop. Report the `empty(...)` construct
    with no fix. Typical: string/nullable-string subjects, mixed unions like
    `bool|null|Foo`, literals (`empty(0)`, `empty('')`, `empty(null)`),
    unknown types, and property accesses whose type does not qualify for D2.

## Exceptions (no report)
E1. Array element subjects (D0), regardless of options.
E2. D2-qualifying subjects that go through a property access (guard in D2),
    regardless of `REPORT_EMPTY_USAGE`.
E3. Everything when all four options are off.
E4. `?string`, `string|null`, `int|bool|null` (three members), `array|null`,
    `mixed`, `object`, `callable` do not qualify for D1/D2 (they may still get
    D3).

## Report
- Range:
  - D1/D2 not inverted: the `empty(...)` construct (`empty` through `)`).
  - D1/D2 inverted: the whole `!empty(...)` unary expression (including `!`).
  - D3: always only the `empty(...)` construct, even when preceded by `!`.
- Severity: info (fixtures tag it `weak_warning`).
- Messages:
  - D1/D2: `Replace with '{replacement}'.` (replacement = fix text below);
  - D3: `Prefer a type-specific check over empty().`

## Fix
The reported range (the `empty(...)` or the `!empty(...)`) is replaced by a
new expression; `{S}` is the verbatim source text of the subject with its
wrapping parentheses removed. `{op}` is `===` when not inverted, `!==` when
inverted. No parentheses are added around the result.

F1. (D1) regular comparison style (default): `count({S}) {op} 0`;
    yoda style: `0 {op} count({S})`.
F2. (D2) regular: `{S} {op} null`; yoda: `null {op} {S}`.
Builtin spelling: `count` is written `\count` when an unqualified call at
that position would not reach the global function (a `use function`
import under that name, or a same-named function declared in the current
namespace) (F1, message included): `\count($rows) === 0`.
D3 has no fix.

## Options
| Option | Type | Default | Effect |
|---|---|---|---|
| REPORT_EMPTY_USAGE | bool | false | Enables the generic report D3. |
| SUGGEST_TO_USE_COUNT_CHECK | bool | false | Enables D1 (count comparison for arrays/Countable). |
| SUGGEST_TO_USE_NULL_COMPARISON | bool | true | Enables D2 (null comparison) as a whole. |
| SUGGEST_NULL_COMPARISON_FOR_SCALARS | bool | true | Enables D2a (nullable int/float/bool/resource); D2b (objects) only needs the previous option. |

Global setting: comparison style (`regular` default / `yoda`) selects the
operand order of F1/F2.

## PHP versions
None.

## Examples
All four options on, regular style:

```php
<?php

function tally(array $rows, \SplObjectStorage $seen)
{
    if (<weak_warning descr="Replace with 'count($rows) === 0'.">empty($rows)</weak_warning>) {}
    if (<weak_warning descr="Replace with 'count($seen) !== 0'.">!empty($seen)</weak_warning>) {}
    if (<weak_warning descr="Replace with 'count($rows) === 0'.">empty(($rows))</weak_warning>) {}
}

/**
 * @param float|null $ratio
 * @param true|null  $flag
 * @param string|null $label
 */
function probe($ratio, $flag, $label, ?\DateTime $when, \stdClass $box, $map)
{
    return [
        <weak_warning descr="Replace with '$ratio === null'.">empty($ratio)</weak_warning>,
        <weak_warning descr="Replace with '$flag !== null'.">!empty($flag)</weak_warning>,
        <weak_warning descr="Replace with '$when === null'.">empty($when)</weak_warning>,
        <weak_warning descr="Replace with '$box === null'.">empty($box)</weak_warning>,
        <weak_warning descr="Prefer a type-specific check over empty().">empty($label)</weak_warning>,
        !<weak_warning descr="Prefer a type-specific check over empty().">empty($label)</weak_warning>,
        <weak_warning descr="Prefer a type-specific check over empty().">empty(0)</weak_warning>,
        empty($map['key']),
    ];
}

class Node
{
    /** @var Node|null */
    public $next;
    /** @var string */
    public $name;
}

function walk(Node $n)
{
    return [
        empty($n->next),
        empty($n->next->next),
        <weak_warning descr="Prefer a type-specific check over empty().">empty($n->name)</weak_warning>,
    ];
}
```

```php
<?php

function tally(array $rows, \SplObjectStorage $seen)
{
    if (count($rows) === 0) {}
    if (count($seen) !== 0) {}
    if (count($rows) === 0) {}
}

/**
 * @param float|null $ratio
 * @param true|null  $flag
 * @param string|null $label
 */
function probe($ratio, $flag, $label, ?\DateTime $when, \stdClass $box, $map)
{
    return [
        $ratio === null,
        $flag !== null,
        $when === null,
        $box === null,
        empty($label),
        !empty($label),
        empty(0),
        empty($map['key']),
    ];
}

class Node
{
    /** @var Node|null */
    public $next;
    /** @var string */
    public $name;
}

function walk(Node $n)
{
    return [
        empty($n->next),
        empty($n->next->next),
        empty($n->name),
    ];
}
```

Yoda style, default options: `empty($ratio)` → `null === $ratio`;
`!empty($when)` → `null !== $when`.

## Divergences
- The replacement is inserted without parentheses; in a higher-precedence
  context (`'x' . empty($n)`) upstream's output changes meaning. Recommendation:
  wrap the replacement in parentheses when the parent expression binds tighter
  than a comparison (conformance fixtures only use statement/array/argument
  contexts, so this is safe).
- **Guard stops at calls (custos diverges):** upstream's "first child"
  descent also walks into a function call's argument list, so
  `empty(make($o->p))` is silently skipped when D2 would apply, although the
  call result has nothing to do with property-access semantics. custos stops
  the descent at argument lists (D2).
- **Builtin spelling (custos diverges).** Upstream inserts a bare `count(`,
  which a function of that name declared in (or imported into) the current
  namespace captures, so the fix would call user code. custos writes
  `\count(` in that case.
- **Low-precedence subjects (custos diverges):** `{S}` is inserted verbatim,
  so `empty($a ?? $b)` became `$a ?? $b === null`, which compares only `$b`
  (and likewise for ternaries, assignments, logical/bitwise operators). In
  F2 custos wraps such a subject in parentheses: `($a ?? $b) === null`
  (message included). F1 is unaffected (the subject is a `count()`
  argument).
