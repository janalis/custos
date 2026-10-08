---
id: NullCoalescingOperatorCanBeUsed
group: Language level migration
kind: semantic
needs: [types]
php: { min: "7.0", max: "" }
---

# NullCoalescingOperatorCanBeUsed

## Summary

Ternaries and small `if` constructs that pick a value when it "exists"
(`isset`, `!== null`, `array_key_exists`, truthiness of the owning object) and a
fallback otherwise can be collapsed into a single `??` expression, which is
shorter and easier to read.

## Detection

### Terminology

- *Strip(e)*: `e` with any number of surrounding parentheses removed.
- *Equivalent*: same node kind and same token sequence ignoring
  whitespace/comments, or identical source text; for plain variables, same
  variable name. Node kinds must match: a parenthesised expression is never
  equivalent to the bare expression it wraps.
- *null literal*: the constant `null`, case-insensitive (`NULL`, `Null`),
  optionally `\`-qualified (`\null`).
- *Property access*: `$o->p`, `$o?->p` (non-static) or `X::$p` (static). Its
  *base* is the part before `->`/`?->`/`::`.
- *wrap(e)*: `(` + text + `)` when `e` is a ternary (full or short `?:`), an
  assignment (plain, by-reference or compound), or a binary expression whose
  operator is anything other than `??` (arithmetic, `.`, comparison, logical
  `&&`/`||`/`and`/`or`/`xor`, bitwise, `<=>`, `instanceof`); otherwise the
  text as is (already-parenthesised expressions are not wrapped again).

### Condition classification (shared)

Given a condition `C0`, let `C = Strip(C0)`. If `C` is a logical-not `!X`,
the condition is **negated** and the subject condition is `K = Strip(X)`;
otherwise it is **positive** and `K = C`. (Only one `!` is unwrapped; other
unary operators make it unclassifiable.) `K` is classified as:

- **K-truthy**: a variable, array access or property access.
- **K-isset**: `isset(S)` with exactly one argument `S`.
- **K-empty**: `empty(S)` (exactly one argument).
- **K-null**: a `===` or `!==` comparison where the right operand or the left
  operand is the null literal. Its *subject* is the left operand, unless the
  left operand is the null literal, in which case it is the right operand.
- **K-ake**: a call to the global function `array_key_exists` (name compared
  case-insensitively; unqualified or `\`-qualified, not shadowed by a function
  of that name declared in or imported into the current namespace; argument
  count checked later).
- anything else: no report.

### Replacement generation (shared)

Inputs: the classified condition, `T` = the value used when the condition is
true, `F` = the value used when it is false (`F` may be **absent** only in the
`if` forms, see below). Produces replacement `R` or nothing. "cand" is the
value that must equal the probed expression, "alt" the fallback.

- **G1 K-isset**: positive → cand = `T`, alt = `F`; negated → cand = `F`,
  alt = `T`. Requires cand present and equivalent to `S`.
  `R = wrap(cand) ?? wrap(alt)`, with alt text `null` when alt is absent.
- **G2 K-empty**: note the inversion — negated (`!empty(S)`) → cand = `T`,
  alt = `F`; positive (`empty(S)`) → cand = `F`, alt = `T`. Requires cand to
  be a property access (static or not) whose base is equivalent to `S`, and
  G6 to hold with probe `S` (null allowed in the probe only for a non-static
  cand). `R = cand ?? wrap(alt)` (alt absent → `null`).
- **G3 K-truthy** (`K` itself is the probed expression): positive → cand =
  `T`, alt = `F`; negated → cand = `F`, alt = `T`. Requires cand to be a
  **non-static** property access whose base is equivalent to `K`. Then the
  inferred type of cand must contain at least one known type, and G6 must
  hold with probe `K` (null allowed in the probe).
  `R = cand ?? wrap(alt)` (alt absent → `null`).
- **G4 K-ake** (requires `F` present): the call must have exactly two
  arguments `(key, arr)`. Positive → cand = `T`, alt = `F`; negated → swap.
  cand must be an array access `c[i]` with a non-empty index, `c` equivalent to
  `arr` and `i` equivalent to `key`; alt must be the null literal.
  `R = wrap(cand) ?? wrap(alt)`.
- **G5 K-null** (requires `F` present): the probed value "is set" when the
  operator is `!==` and the condition is positive, or the operator is `===`
  and the condition is negated. If set-branch is the true branch → cand = `T`,
  alt = `F`; otherwise cand = `F`, alt = `T`. cand must be equivalent to the
  subject. `R = wrap(cand) ?? wrap(alt)`.

- **G6 safe fallback** (G2/G3): when alt is absent or the null literal, it
  holds. Otherwise all of:
  - the inferred type of cand is known and does not contain `null`
    (a set object whose property is `null` produced `null`, while `??` would
    produce the fallback);
  - the inferred type of the probe is known and every component is a class
    type, `object` or `static` (or `null` where allowed). A probe holding a
    truthy scalar or array made the original read a property on a
    non-object and yield `null`; `??` would yield the fallback.
  With a null fallback both forms yield `null` in every case (the original
  may additionally emit a warning for a property read on a non-object).

Separator in `R` is exactly ` ?? ` (one space each side).

### Form A — ternaries (option SUGGEST_SIMPLIFYING_TERNARIES)

- **D1** A full ternary `C0 ? T : F` (short ternaries `?:` are ignored).
- **D2** Classify `C0`; generate `R` with `T`, `F` taken verbatim (including
  any parentheses around a branch). Report when `R` exists.

### Form B — `if` statements (option SUGGEST_SIMPLIFYING_IFS)

- **D3** An `if` statement without `elseif` branches (an `else` whose body is
  another `if`, i.e. `else if`, is allowed on the *outer* statement but then
  the outer statement does not match any shape below; the inner `if` is
  examined on its own).
- **D4** Classify its condition.
- **D5** The `if` body must be a braced block containing exactly one statement
  (comments ignored). That statement must be a *candidate*: either a `return`
  statement (with or without value), or an expression statement whose
  expression is a plain `=` assignment (not compound, not list destructuring)
  to a plain variable `$v`.
- **D6** Determine the shape (first matching rule wins; if the governing rule
  fails its inner conditions, there is no fallback to later rules):
  - **S1 if/else return** (has `else`): the `else` body is a braced block with
    exactly one statement, and both candidates are returns. `T` = the if
    return's value, `F` = the else return's value (either may be absent).
    Removal range: the whole `if` statement.
  - **S2 if/else assign** (has `else`): the `else` body is a braced block with
    exactly one statement; both candidates are assignments to plain variables
    with equivalent targets. `T`/`F` = the assigned values. Removal range: the
    whole `if` statement.
  - (has `else` but neither S1 nor S2 → no report.)
  - **S3 assign then if** (no `else`): the previous sibling statement in the
    same statement list (skipping comments/whitespace) is an expression
    statement holding a plain `=` assignment to a plain variable, and the if
    candidate is an assignment. Then all of: the two targets are equivalent;
    the previous assignment is not by reference (`=&`); its value is not
    itself an assignment (`$a = $b = …`); its value does not contain, anywhere
    inside it (as a descendant), a variable equivalent to the target
    (`$x = trim($x);` disqualifies); its value has no side effects (no
    function/method/static call, `|>`, `new`, `clone`, assignment,
    `++`/`--`, `include`/`require`, `eval`, `exit`, `print`, `throw`,
    `yield` or backtick command; closure bodies are not inspected), because
    in `R` it becomes the fallback and is only evaluated when the probe is
    unset. `T` = the if-body value, `F` = the
    previous statement's value. Removal range: from the start of the previous
    statement through the end of the `if`.
  - **S4 if-return then return** (no `else`): the if candidate is a return and
    the next sibling statement is a `return` statement. `T` = the if return's
    value, `F` = the next return's value (may be absent for bare `return;`).
    Removal range: from the start of the `if` through the end of that return.
  - **S5 if-return at end of body** (no `else`): the if candidate is a return,
    there is no next sibling statement at all, and the `if` sits directly in
    the body block of a function, method or closure (not inside a nested
    block; arrow functions have no statement body). `T` = the return value, `F` absent.
    Removal range: the whole `if`.
- **D7** `T` must be present (an `if (...) { return; }` never matches).
  Generate `R`. The statement text `Q` is:
  - when `T` comes from a return: `return R`;
  - when `T` comes from an assignment: `{target} = R` where `{target}` is the
    source text of the if-body assignment's variable.
- **D8** Report when `Q` exists.
- **Name case.** Wherever this rule compares two expressions for
  equivalence, the names PHP resolves case-insensitively — function and
  method names, class names in calls, `new`, `instanceof` and `::`
  accesses, and keywords — compare case-insensitively (`Cache::$map['k']`
  matches `cache::$map['k']`, `$o->Name()` matches `$o->name()`); variable,
  property and constant names stay case-sensitive.

## Exceptions (no report)

- **E1** Language level below 7.0, or the corresponding option is off.
- **E2** Short ternary `?:`.
- **E3** `isset()`/`empty()` with more than one argument; `array_key_exists`
  with an argument count other than 2.
- **E4** cand not equivalent to the probed expression (`isset($a[0]) ?
  $a[0]->x : …`, `trim($v)` instead of `$v`, a parenthesised branch).
- **E5** `array_key_exists` with a non-null fallback.
- **E6** K-truthy: base mismatch; static property (`$o ? $o::$p : null`);
  class constants (`$o::C`, `$o::class`) — not property accesses; cand type
  unknown/unresolvable; G6 failing with a non-null fallback (nullable cand,
  probe not known to be an object).
- **E9** K-empty with a non-null fallback when G6 fails:
  `!empty($o) ? $o->label : 'none'` with a nullable `label`, or a probe of
  unknown or scalar type.
- **E10** S3 when the previous value has side effects:
  `$v = load(); if (isset($w)) { $v = $w; }`.
- **E7** `if` shapes: `elseif` present; bodies not braced or with more than one
  statement; container mismatch; compound or by-reference previous assignment;
  chained previous assignment; previous value reading the container; inverted
  logic producing a non-matching cand (`$v = 'd'; if ($p === null) { $v = $p; }`).
- **E8** K-null and K-ake need a false-branch value: S5 (and bare `return;`
  as `F`) never produce a report for them.

## Report

- Form A range: the whole ternary expression (condition start to false-branch
  end, without parentheses wrapping the ternary).
- Form B range: the `if` keyword token only (for an `else if`, the `if`
  keyword of the inner statement).
- Severity: info (weak warning).
- Message: `Simplify to '{Q}' using the null coalescing operator.` where `{Q}`
  is `R` for Form A and the statement text `Q` for Form B.

## Fix

- **F1 Form A**: replace the ternary with `R` verbatim.
- **F2 Form B, S1/S2/S5**: replace the whole `if` statement (through the end of
  its `else` block, if any) with `Q;`. If that `if` is itself the body of an
  `else` (`else if`), replace it with `{ Q; }` instead, yielding
  `else { Q; }` (upstream reformats the block onto several lines; whitespace
  is irrelevant for comparison).
- **F3 Form B, S3**: replace the text range from the start of the previous
  assignment statement through the end of the `if` statement (including the
  whitespace/comments between them) with `Q;`.
- **F4 Form B, S4**: replace the text range from the start of the `if` through
  the end of the following `return` statement (including the `;`) with `Q;`.
- Whitespace before the replaced range and after it is kept untouched.

## Options

| Option | Type | Default | Effect |
|---|---|---|---|
| SUGGEST_SIMPLIFYING_TERNARIES | bool | true | Enables Form A (ternaries). |
| SUGGEST_SIMPLIFYING_IFS | bool | true | Enables Form B (`if` statements). |

Both upstream fixtures set only one option explicitly, but the other one keeps
its default `true`, so both forms are active in both runs.

## PHP versions

Nothing is reported below PHP 7.0. Upstream fixtures run at 7.1.

## Type inference requirements (G3 only)

The cand property's type comes from: declared property types, `@var` doc
comments on properties (`null|Foo`, `string`, …), `@var` inline doc comments
declaring a variable's class (`/** @var Foo $x */`), element types of
array-typed variables (`Foo[]` → element `Foo`), and chained accesses
(`$x->a->b` where `a` is typed `null|Foo` and `Foo::$b` is `string` → `string`;
nullability of the intermediate does not leak into the final type). When the
base's class or the property cannot be resolved, there is no known type and no
report.

## Examples

```php
<?php
class Box {
    /** @var int */
    public $size;
    /** @var Box|null */
    public $inner;
}
/** @var Box $box */
/** @var Box[] $boxes */

$a = <weak_warning descr="Simplify to '$opts['k'] ?? 5' using the null coalescing operator.">isset($opts['k']) ? $opts['k'] : 5</weak_warning>;
$b = <weak_warning descr="Simplify to '$opts['k'] ?? null' using the null coalescing operator.">!isset($opts['k']) ? null : $opts['k']</weak_warning>;
$c = <weak_warning descr="Simplify to '$row->id ?? ($fallback . 'x')' using the null coalescing operator.">null === $row->id ? $fallback . 'x' : $row->id</weak_warning>;
$d = <weak_warning descr="Simplify to '$map[$k] ?? null' using the null coalescing operator.">(array_key_exists($k, $map)) ? $map[$k] : NULL</weak_warning>;
$e = <weak_warning descr="Simplify to '$box->size ?? 0' using the null coalescing operator.">$box ? $box->size : 0</weak_warning>;
$f = <weak_warning descr="Simplify to '$box->inner ?? null' using the null coalescing operator.">!$box ? null : $box->inner</weak_warning>;
$g = <weak_warning descr="Simplify to '$boxes[1]->size ?? null' using the null coalescing operator.">!empty($boxes[1]) ? $boxes[1]->size : null</weak_warning>;
$h = $box ? $box->inner : -1;            // nullable with non-null fallback
$i = $mystery ? $mystery->size : null;   // unknown type
$j = array_key_exists($k, $map) ? $map[$k] : 0;
$k2 = isset($opts['k'], $opts['j']) ? $opts['k'] : 5;
$l = isset($opts['k']) ?: 5;

function pickA($in) {
    $out = 'none';
    <weak_warning descr="Simplify to '$out = $in['v'] ?? 'none'' using the null coalescing operator.">if</weak_warning> (isset($in['v'])) {
        $out = $in['v'];
    }
    $n = abs($n);
    if (isset($in[$n])) { $n = $in[$n]; }      // previous value reads the target
    <weak_warning descr="Simplify to 'return $in['w'] ?? 'none'' using the null coalescing operator.">if</weak_warning> ($in['w'] !== null) {
        return $in['w'];
    }
    return 'none';
}

function pickB($in) {
    if ($in) {
        $res = 1;
    } else <weak_warning descr="Simplify to '$res = $in ?? 2' using the null coalescing operator.">if</weak_warning> (isset($in)) {
        $res = $in;
    } else {
        $res = 2;
    }
    <weak_warning descr="Simplify to 'return $in ?? null' using the null coalescing operator.">if</weak_warning> (!(null === $in)) {
        return $in;
    } else {
        return null;
    }
}

$cb = function ($q) {
    <weak_warning descr="Simplify to 'return $q ?? null' using the null coalescing operator.">if</weak_warning> (isset($q)) {
        return $q;
    }
};
```

```php
<?php
class Box {
    /** @var int */
    public $size;
    /** @var Box|null */
    public $inner;
}
/** @var Box $box */
/** @var Box[] $boxes */

$a = $opts['k'] ?? 5;
$b = $opts['k'] ?? null;
$c = $row->id ?? ($fallback . 'x');
$d = $map[$k] ?? NULL;
$e = $box->size ?? 0;
$f = $box->inner ?? null;
$g = $boxes[1]->size ?? null;
$h = $box ? $box->inner : -1;            // nullable with non-null fallback
$i = $mystery ? $mystery->size : null;   // unknown type
$j = array_key_exists($k, $map) ? $map[$k] : 0;
$k2 = isset($opts['k'], $opts['j']) ? $opts['k'] : 5;
$l = isset($opts['k']) ?: 5;

function pickA($in) {
    $out = $in['v'] ?? 'none';
    $n = abs($n);
    if (isset($in[$n])) { $n = $in[$n]; }      // previous value reads the target
    return $in['w'] ?? 'none';
}

function pickB($in) {
    if ($in) {
        $res = 1;
    } else { $res = $in ?? 2; }
    return $in ?? null;
}

$cb = function ($q) {
    return $q ?? null;
};
```

Note for `pickB`: the outer `if ($in)` has an `else` whose body is an `if`, so
it matches no shape; only the inner `if` is reported. `if (!(null === $in))`
is negated K-null with `===`, so the true branch is the "set" branch.

## Divergences

- G2/G3 fallback (custos diverges from upstream). Upstream rewrites
  `!empty($o) ? $o->p : 'x'` without any type check, and the truthy form
  `$o ? $o->p : 'x'` only checks the property's nullability. Both change the
  result when the property is `null` on a set object (was `null`, becomes
  `'x'`) or when the probe holds a truthy non-object (was `null` with a
  warning, becomes `'x'`). custos requires G6 for any non-null fallback:
  a known non-nullable property type and a probe known to be an object (or
  null). With a null fallback the rewrite is value-preserving and is kept.
- G2 with a static candidate (custos diverges from upstream).
  `!empty($c) ? $c::$p : null` skips the class lookup when `$c` is `''`,
  `null`, `0` or `false`, but `$c::$p ?? null` performs it and throws
  (`Class "" not found`, "Class name must be a valid object or a string"):
  `??` does not guard the class part. custos rewrites the static form only
  when the probe's type is known and holds objects only (no `null`, no
  string), whatever the fallback.
- S3 (custos diverges from upstream). Upstream turns
  `$v = f(); if (isset($w)) { $v = $w; }` into `$v = $w ?? f();`, so `f()`
  is only evaluated when `$w` is unset and its side effects may be lost.
  custos skips S3 when the previous value has side effects (E10).
- Alternative-syntax `if (): … endif;` bodies: upstream treatment unverified;
  recommendation: do not report (only braced blocks qualify).
- **K-ake resolution and case (custos diverges):** upstream matches
  `array_key_exists` case-sensitively on the written last segment: it misses
  `Array_Key_Exists('k', $a) ? $a['k'] : null` and treats a user
  `Cfg\array_key_exists()` as the built-in. custos matches any case and
  requires the call to resolve to the global function; `\null` also counts
  as the null literal.
- **Name case (custos diverges).** Upstream compares the expressions
  textually, so operands that differ only in the case of a function, method
  or class name (`Stats::$n` vs `stats::$n`), which PHP treats as the same,
  are not recognised as equivalent. custos folds the case of those names
  (Detection, "Name case").
- **References (custos diverges).** An if/else (or S3) whose assignment
  binds a reference (`$item = &$this->menu[$k];`) is not rewritten: `??`
  yields a value, so later writes through `$item` would no longer reach the
  array (Matomo `MenuAbstract`). Neither is the if/return form inside a
  function declared to return by reference (`function &get()`).
- **S3 needs a probe independent of the target (custos diverges).** The
  merged form `T = V ?? X` no longer assigns `X` to `T` before the probe
  runs, so S3 does not apply when the condition or the assigned value
  reads `T`: `$type = $parts[1]; if (isset($mimeMap[$type])) { $type =
  $mimeMap[$type]; }` became `$type = $mimeMap[$type] ?? $parts[1]`, which
  looks up the previous `$type` (SuiteCRM `get_file_mime_type()`).
