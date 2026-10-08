---
id: UnSafeIsSetOverArray
group: Control flow
kind: semantic
needs: [types, hierarchy, index]
php: { min: "", max: "" }
---

# UnSafeIsSetOverArray

## Summary

`isset()` mixes two questions: "does the key/variable exist" and "is it not
null". This rule points out single-argument `isset()` calls where a more
explicit construct says what is meant: a plain `!== null` comparison for
variables and declared properties, `array_key_exists()` for array keys, and
flags concatenated index expressions that should be computed beforehand.

## Detection

Visit every `isset(...)` construct.

### Context gathering

- **D1** The `isset` must have exactly one argument. `isset()` with two or more
  arguments is never reported.
- **D2** Determine the *context node* `P` and the *inverted* flag:
  - if the direct parent of `isset(...)` is a unary logical-not `!` (directly,
    no parentheses in between), `inverted = true` and `P` is the parent of that
    `!` expression; the *subject* is the `!isset(...)` expression;
  - otherwise `inverted = false`, `P` is the direct parent of the `isset`, and
    the subject is the `isset(...)` itself.
- **D3** `stored = true` when `P` is an assignment expression (plain `=` or any
  compound assignment such as `.=`, including `$list[] = …`) or a `return`
  statement.
- **D4** Ternary bail-out: if `P` is a ternary expression and the subject is
  exactly its condition node (not wrapped in parentheses), look at the branch
  taken when the key is missing — the *false* branch when not inverted, the
  *true* branch when inverted. If that branch is literally the constant `null`
  (case-insensitive, optionally `\null`; parentheses around it are *not*
  stripped), nothing at all is reported for this `isset`.
- **D5** Let `A` be the single argument with surrounding parentheses removed.

### Variables and properties (A is not an array access)

- **D6** If `A` is a plain variable (`$name`, also `$this` / variable-variables)
  and it is **not** inside any function, method, closure, arrow function or
  class body (i.e. it is in the global/template scope), nothing is reported.
- **D7** If `A` is a property access (`$obj->name`, `$this->name`,
  `Cls::$name`), it is only considered when the property resolves to a
  **declared** property (class body or constructor promotion, including
  inherited declarations).
  Not reported when the property cannot be resolved (dynamic property, `stdClass`,
  `$this->$name`, unknown receiver type), or when it resolves only to a
  docblock `@property` tag. A constructor-promoted property counts as
  declared (see Divergences).
- **D8** For any remaining non-array-access `A` (variables in a function-like
  scope, resolved declared properties, anything else that passed D6/D7):
  when option `SUGGEST_TO_USE_NULL_COMPARISON` is on **and** the `isset` is not
  located anywhere inside a `finally { … }` block (at any depth, closures
  included), report the subject with the *null-comparison* problem.
  Replacement text `R`:
  - regular comparison style: `A !== null` (not inverted) / `A === null`
    (inverted);
  - yoda comparison style: `null !== A` / `null === A`.
  `A` is the argument's source text verbatim; parts are joined with one space.
- In all non-array-access cases, processing stops here (no other check applies).

### Array accesses (A is `container[index]`)

- **D9** Concatenated index: when `REPORT_CONCATENATION_IN_INDEXES` is on and
  `stored` is false, and **any** index of the `[...]` chain of `A` (the
  last bracket, then the brackets of its container while that container is
  itself an array access) is directly a binary concatenation (`x . y`, not
  wrapped in parentheses), report `A` with the *concatenation* problem and
  stop. `$m['a' . $b]['c']` matches; the chain stops at anything that is not
  an array access (`$o->{'a' . $b}['c']`, `f('a' . $b)['c']` do not).
- **D10** Otherwise, when `SUGGEST_TO_USE_ARRAY_KEY_EXISTS` is on, report `A`
  with the *array_key_exists* problem **unless** the container (the expression
  before the last `[...]`) is known to be an object type. Decide from the set of
  resolved types of the container (unknown parts dropped; any type spelled with
  `[]` counts as `array`):
  - empty set (type cannot be inferred) → report;
  - any member that is a non-class type (`array`, `string`, `int`, `float`,
    `bool`, `callable`, `iterable`, `resource`, `void`, …) → report;
  - `null` members are ignored; if only `null` remains → report;
  - otherwise (at least one class/interface type, the `object` pseudo-type,
    or `mixed`, and no non-class type) → not reported (the object is assumed
    to implement `ArrayAccess`; `array_key_exists()` does not accept
    objects).
  For chained accesses the container is the inner access (`$m['a']` in
  `$m['a']['b']`); its type is the element type of `$m`, usually unknown
  (→ report) or `mixed` (→ no report). No upstream fixture covers this.
  The D9 `stored` restriction does not apply here: `$v = isset($a['k' . $n]);`
  gets the array_key_exists report.

## Exceptions (no report)

- **E1** `isset` with zero or several arguments.
- **E2** Ternary "isset-or-null" idiom: `isset($a[k]) ? $a[k] : null`,
  `!isset($a[k]) ? null : $a[k]` (D4).
- **E3** Plain variables at file/global level (D6).
- **E4** Unresolved, dynamic, magic (`@property`) properties (D7).
- **E5** Null-comparison suggestion inside `finally` blocks.
- **E6** Concatenated index when the isset result is assigned or returned
  (falls through to D10 instead).
- **E7** Array access whose container is an object type (ArrayAccess, stdClass,
  any class/interface, the `object` pseudo-type) or `mixed`.
- **E8** `!(isset(...))` — the `!` is not the direct parent, so the expression
  is treated as non-inverted; the subject is just `isset(...)`.

## Report

| Problem | Range | Severity |
|---|---|---|
| null-comparison (D8) | the subject: `isset(...)` or, when inverted, the whole `!isset(...)` from `!` to `)` | info (weak warning) |
| concatenation (D9) | the argument `A` (inside the parentheses) | warning (rule default) |
| array_key_exists (D10) | the argument `A` | info (weak warning) |

Messages (our wording):

- null-comparison: `Compare with null instead: '{R}'.`
- concatenation: `Compute the concatenated key in a variable before using it.`
- array_key_exists: `Use array_key_exists() to check for the key itself.`

## Fix

- **F1** null-comparison only: replace the subject (`isset(...)` or
  `!isset(...)`) with `R` verbatim (no parentheses added).
  `$ok = isset($node);` → `$ok = $node !== null;`;
  `$ok = !isset($this->cache);` → `$ok = $this->cache === null;`.
- The concatenation and array_key_exists problems have no fix.

## Options

| Option | Type | Default | Effect |
|---|---|---|---|
| SUGGEST_TO_USE_ARRAY_KEY_EXISTS | bool | false | Enables D10. |
| SUGGEST_TO_USE_NULL_COMPARISON | bool | false | Enables D8 and its fix. |
| REPORT_CONCATENATION_IN_INDEXES | bool | true | Enables D9. When off, such accesses go straight to D10. |

All upstream fixtures run with all three options **on** and regular comparison
style.

## PHP versions

No gating. Fixtures run at the IDE test default (PHP below 7.1).

## Examples

Options: all three `true`, regular comparison style.

```php
<?php
$grid = [];
if (isset(<warning descr="Compute the concatenated key in a variable before using it.">$grid['r' . $row]</warning>)) {
    $both = isset($grid['a'], $grid['b']);
    $cell = isset($grid['k']) ? $grid['k'] : NULL;
    $cell = !isset($grid['k']) ? null : $grid['k'];
    print isset(<weak_warning descr="Use array_key_exists() to check for the key itself.">$grid[$col]</weak_warning>);
    $seen = !isset(<weak_warning descr="Use array_key_exists() to check for the key itself.">$grid['r' . $row]</weak_warning>);
}
$top = isset($grid);

class Ledger {
    public $total;
    public function check(\ArrayAccess $bag, $field) {
        $tmp = new stdClass();
        $u = isset($bag['x']);
        $v = isset($tmp->anything);
        $w = isset($this->$field);
        $p = <weak_warning descr="Compare with null instead: '$this->total !== null'.">isset($this->total)</weak_warning>;
        $q = <weak_warning descr="Compare with null instead: '$tmp === null'.">!isset($tmp)</weak_warning>;
        try {
            return 1;
        } finally {
            $r = isset($tmp);
        }
    }
}
```

```php
<?php
$grid = [];
if (isset($grid['r' . $row])) {
    $both = isset($grid['a'], $grid['b']);
    $cell = isset($grid['k']) ? $grid['k'] : NULL;
    $cell = !isset($grid['k']) ? null : $grid['k'];
    print isset($grid[$col]);
    $seen = !isset($grid['r' . $row]);
}
$top = isset($grid);

class Ledger {
    public $total;
    public function check(\ArrayAccess $bag, $field) {
        $tmp = new stdClass();
        $u = isset($bag['x']);
        $v = isset($tmp->anything);
        $w = isset($this->$field);
        $p = $this->total !== null;
        $q = $tmp === null;
        try {
            return 1;
        } finally {
            $r = isset($tmp);
        }
    }
}
```

## Divergences

- **D9 every index level (custos diverges).** Upstream only inspects the
  last `[...]` of the access chain, so `$m['a' . $b]['c']` falls through to
  the array_key_exists check although the concatenated key is just as much
  worth precomputing. custos inspects every bracket of the chain.
- D8 replacement has no parentheses: the fix of `!isset($a) && $b` is fine, but
  an argument with lower-precedence parts cannot occur (isset only accepts
  variables/properties/accesses), so no precedence issue is expected.
- **Constructor-promoted properties (custos diverges).** Upstream resolves
  a promoted property (PHP 8.0+) to the constructor parameter rather than a
  class-body declaration, so `isset($this->promoted)` never gets the
  null-comparison suggestion. A promoted property is a declared property
  like any other, so custos treats it as one (D7).
- **D10 `object` — custos diverges from upstream.** Upstream classifies the
  `object` pseudo-type as a scalar, so `function f(object $o) { isset($o['k']); }`
  gets the array_key_exists suggestion, although `array_key_exists()` throws
  a TypeError on objects (PHP 8) and the access can only work through
  `ArrayAccess`. custos treats `object` like a class type (no report).
