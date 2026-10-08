---
id: UnnecessaryCasting
group: Code style
kind: semantic
needs: [names, types, hierarchy, index, stubs]
php: { min: "", max: "" }
---

# UnnecessaryCasting

## Summary
A type cast whose operand already has exactly the target type, or a string cast
used as an operand of concatenation (which converts to string anyway), is noise.
Dropping it simplifies the expression.

## Detection
Visit every cast expression `(T) operand`. Let *A* be the operand with all
enclosing parentheses removed (`(int) (($x))` → `$x`).

Cast kinds and their target type:

| cast tokens (case-insensitive, inner spaces allowed) | target |
|---|---|
| `(int)`, `(integer)` | int |
| `(float)`, `(double)`, `(real)` | float |
| `(bool)`, `(boolean)` | bool |
| `(string)`, `(binary)` | string |
| `(array)` | array |

`(object)` and `(unset)` are never reported.

### Concatenation (string casts only)
- **D1** A `(string)` cast whose direct parent is a binary `.` expression
  (either operand) → report (message M2). No type check.
- **D2** A `(string)` cast whose direct parent is a `.=` compound assignment →
  report (message M2). No type check.
- When D1/D2 fires, D3 is not evaluated for that cast.

### Already of the target type
- **D3** The *strict type* of *A* (below) consists of exactly one type after
  normalisation, and that type equals the cast's target. Then:
  - **D3a** if *A* is a variable whose name matches a parameter of the nearest
    enclosing function-like scope (function, method, closure, arrow function)
    and that parameter has **no declared type** (docblock types do not count),
    do not report;
  - **D3b** if the *possible-values* set of *A* (below) has exactly one element
    and that element is a direct operand of a `??` expression, do not report;
  - otherwise report (message M1).

### Strict type of *A*
- **T1** *A* is a property fetch (`$obj->p`, `$this->p`, `$obj?->p`): resolve
  the property; only a **private** property yields a type, namely the
  property's own type (declared type; if none, union of its `@var` docblock
  types and the type of its default value — `null` when it has no default,
  unless it is promoted or this class's constructor assigns it a non-null
  value in a top-level statement; custos diverges, see Divergences).
  Non-private or unresolved → empty.
- **T2** *A* is a function / method / static-method call: resolve the callee;
  only if the callee has a **declared** return type (in source or stubs) does
  it yield a type, namely the call's type per T-rules below (R-function
  overrides apply for plain functions). No declared return type → empty.
- **T3** Anything else: the expression type per T-rules below.
- **T4** Normalisation: drop unknown/unresolved types. If *A* is a member
  access (property fetch or method call) using `?->` and the set does not
  already contain `null`, add `null`. Then normalise each name: lowercase,
  strip a leading `\`; `integer`→`int`, `boolean`/`true`/`false`→`bool`,
  `double` is not an alias (keep as resolved), anything containing `[]`→`array`.
  Count distinct normalised types.

### Expression type rules (T-rules)
- String literal (any quoting, heredoc, interpolated) → string. Integer literal
  → int. Float literal (`0.0`, `.5`, `1e3`) → float. `true`/`false` → bool.
  `null` → null. Array literal → array.
- Global constant: if defined via `define('N', value)`/`const N = value`, the type of
  the value; built-in constants per stubs.
- Class constant `X::C`: `::class` → string; otherwise the type of the constant's
  value.
- Parenthesized expression → type of its content.
- Assignment `$a = v` (plain `=`) → type of `v`.
- Unary `-` or `~` → type of its operand. Unary `+`, `!`, casts: default rules
  (`!` → bool, `(T)` → T).
- Arithmetic binary `+`, `-`, `*`, `/`, `**` (not `%`): compute as follows.
  Let *L* be the left operand's normalised type set (unknowns removed).
  `float-ish(S)` = S is empty, or contains float, or contains `number`, or
  contains string but not int.
  1. `isFloat = float-ish(L)`; `isArray = L contains array`.
  2. If `!isFloat`, or (`!isArray` and operator is `+`): let *R* be the right
     operand's set; `isFloat = isFloat || float-ish(R)`;
     `isArray = (isArray && right operand is not a numeric literal) || R contains array`.
  3. Result: array if `isArray`; else float if `isFloat`; else, for `/`,
     `{int, float}` (an integer division returns a float unless it is
     exact; this step is specific to this rule — specs referencing these
     T-rules, such as CallableParameterUseCaseInTypeContext, keep int);
     else int.
  **custos (this rule only, see Divergences):** arithmetic is typed by
  PHP's actual result instead: an operand with no known type → empty; any
  float-only operand → float; int with int → int (`/` → `{int, float}`;
  `**` with a non-literal exponent → `{int, float}`); arrays only for
  `array + array`; anything else (numeric strings, null, bool) →
  `{int, float}`.
  Consequences (upstream heuristic): `int op int` → int for `+`, `-`, `*`, `**`; `int / int` →
  `{int, float}` (two types, so a cast of it is never reported, and a
  variable assigned from it is not single-typed either); any float operand →
  float; an unresolvable operand → float; a string operand (without int) →
  float.
- `??`: if both operands have a non-empty type (unknowns removed): union of
  (left minus null) and right; otherwise **empty**.
- Ternary `c ? a : b`: if both branches have non-empty types: union of both
  (for `c ?: b`, the left part minus null and `false` — it is only the
  result when truthy; custos diverges, see Divergences — plus `b`);
  otherwise default rules.
- Array access on a superglobal (`$_GET[...]`, `$_POST`, `$_COOKIE`,
  `$_REQUEST`, `$_SERVER`, `$_ENV`, `$_FILES`, `$_SESSION`, `$GLOBALS`):
  `{string, array}`; except `$_SERVER['<literal>']`: `argv` → array; `argc`,
  `REQUEST_TIME` → int; `REQUEST_TIME_FLOAT` → float; any other literal key
  (`REMOTE_PORT`, `SERVER_PORT` included, see Divergences) → string.
- `$this->p` property fetch: declared type of the property, if any; otherwise
  default rules.
- Function/method call: union of the callee's declared return type and its
  `@return` docblock types (stubs for built-ins); unresolved → empty. For plain
  (non-method) functions, these overrides replace the result (R-functions):
  - `str_replace`, `str_ireplace`, `preg_replace`, `preg_replace_callback`,
    `substr_replace`, `preg_filter`, `preg_replace_callback_array` →
    `{string, array}`, then narrowed by the subject argument (index 2 for the
    first four and `preg_filter`, index 0 for `substr_replace`, index 1 for
    `preg_replace_callback_array`): if that argument exists and has a fully
    known non-empty type, drop `array` when the argument type has no array and
    drop `string` when it has no string; finally, for the `preg_*` names,
    add `null` (they return null when the regular expression fails — custos
    diverges, see Divergences);
  - `strstr` → `{string, bool}`; `get_class` → `{string}`;
  - `explode` → `{array, bool}`, but with ≥2 arguments and a string-literal
    first argument: `{bool}` if that literal is empty, else `{array}`;
  - `parse_url` → `{array, bool}`, but with exactly 2 arguments whose second is
    a constant: `PHP_URL_PORT` → `{int, null}`, any other constant →
    `{string, null}`;
  - `current`, `reset`, `next`, `prev`, `end` → `{mixed}`;
  - `microtime`: exactly one argument that is not the constant `false` →
    `{float}`; otherwise (no argument, or `false`) → `{int}`;
  - `abs` with one argument whose type is non-empty and only int/float → that
    argument type.
- Variadic parameter variable → array.
- Variable (default): the types of the values assigned to it in the same scope
  (union over assignments visible at that point; at top level, the file's
  top-level assignments), plus — for a parameter — its declared type and
  `@param` docblock types. Undefined → empty. The union is then narrowed by
  the guards on the path to the use (`null !== $v`, `is_*($v)`,
  `instanceof`, truthiness, early exits — the engine's narrowing; custos
  diverges, see Divergences).
- `new X` → X; other expressions → whatever the general type engine provides,
  unknown otherwise.

### Possible-values set (for D3b)
Collect candidate value expressions of *A*, recursively, each expression
visited at most once, parentheses removed first:
- ternary → values of both branches; `??` → values of both operands;
- variable → the default value of a same-named parameter of the enclosing
  function-like scope (if any), plus, for every plain assignment to that variable
  in the scope body, the values of the right-most assigned value (through
  chains `$a = $b = v`); a variable outside any function scope yields nothing.
  **Unstable variable** (custos refinement, see Divergences): if `$v` is also the operand of `++`/`--` or the target of a compound assignment (`+=`, `.=`, `??=`, …) anywhere in that scope body, the whole result is *unknown*: then D3b applies as "do not report" (message M1 is not
  reported);
- property fetch → the property's default value (unless the default's text ends
  with the property name), plus assignments to the same property expression in
  the current scope and in the class constructor;
- class constant → values of its value; global constant (other than
  `true`/`false`/`null`) → its defined value;
- anything else → the expression itself.

## Exceptions (no report)
- **E1** More than one type, or zero types (unknown) for *A*.
- **E2** Untyped-parameter variable (D3a), even if a docblock declares the type.
- **E3** Non-private properties; functions/methods without a declared return
  type (docblock-only `@return`).
- **E4** Single possible value coming from a `??` operand (D3b), e.g.
  `(int) ($undefined ?? 0)`.
- **E5** Null-safe member access (adds `null`).
- **E6** `(object)` / `(unset)` casts.

## Report
- Range: the cast token only, from `(` to `)` inclusive (e.g. `(int)`), not
  the operand.
- Severity: info.
- Messages: **M1** `Operand already has the target type; remove the cast.`
  **M2** `Concatenation converts to string anyway; remove the cast.`

## Fix
- **F1** (both messages) Replace the whole cast expression by its operand as
  written, keeping the operand's own parentheses and dropping the cast token
  and the whitespace between cast and operand:
  `(int) ($n + 1)` → `($n + 1)`; `(string)$s . 'x'` → `$s . 'x'`;
  `(float)-1.5` → `-1.5`.

## Options
None.

## PHP versions
None beyond syntax availability (`?->` from PHP 8.0).

## Examples

```php
<?php
$text  = 'abc';
$list  = [1];
$flag  = false;
$ratio = 1.5;
$count = 4;

return [
    <weak_warning descr="Operand already has the target type; remove the cast.">(string)</weak_warning> $text,
    <weak_warning descr="Operand already has the target type; remove the cast.">(array)</weak_warning> $list,
    <weak_warning descr="Operand already has the target type; remove the cast.">(bool)</weak_warning> $flag,
    <weak_warning descr="Operand already has the target type; remove the cast.">(double)</weak_warning> $ratio,
    <weak_warning descr="Operand already has the target type; remove the cast.">(integer)</weak_warning> ($count * 3),
    <weak_warning descr="Operand already has the target type; remove the cast.">(float)</weak_warning> ($count - 0.5),
    (int) ($count / 2),
    <weak_warning descr="Operand already has the target type; remove the cast.">(float)</weak_warning>-2.25,
    (int) ($count * 0.5),
    (int) ($count + $missing),
    (string) $missing,
];

function strictArg(int $n, $loose) {
    return [<weak_warning descr="Operand already has the target type; remove the cast.">(int)</weak_warning> $n, (int) $loose];
}

class Meter {
    /** @var int */
    private $hidden;
    /** @var int */
    public $shown;
    public function size(): int { return 1; }
    /** @return int */
    public function legacy() { return 1; }

    public function read(?Meter $other) {
        return [
            <weak_warning descr="Operand already has the target type; remove the cast.">(int)</weak_warning> $this->hidden,
            (int) $this->shown,
            <weak_warning descr="Operand already has the target type; remove the cast.">(int)</weak_warning> $this->size(),
            (int) $this->legacy(),
            (int) $other?->size(),
            (int) ($this->shown ?? 0),
        ];
    }
}

function joins($a, $b) {
    $a .= <weak_warning descr="Concatenation converts to string anyway; remove the cast.">(string)</weak_warning> $b;
    return 'n=' . <weak_warning descr="Concatenation converts to string anyway; remove the cast.">(string)</weak_warning>$b;
}

function stamps() {
    return [
        <weak_warning descr="Operand already has the target type; remove the cast.">(int)</weak_warning> (microtime() * 10),
        (int) (microtime(true) * 10),
        (bool) end($GLOBALS),
    ];
}
```

After fix (changed lines):

```php
<?php
    $text,
    $list,
    $flag,
    $ratio,
    ($count * 3),
    ($count - 0.5),
    -2.25,
    return [$n, (int) $loose];
            $this->hidden,
            $this->size(),
    $a .= $b;
    return 'n=' . $b;
        (microtime() * 10),
```

## Divergences
- **`$_SERVER` ports are strings (custos diverges):** upstream types
  `$_SERVER['REMOTE_PORT']` and `['SERVER_PORT']` as int, so `(int)
  $_SERVER['SERVER_PORT']` is reported and the fix drops the cast. Web SAPIs
  (FPM, Apache, the built-in server) store them as strings, so removing the
  cast changes the value's type; custos types them as string.
- **Flow-aware variables and `?:` (custos diverges):** upstream's T-rules
  union every assignment of a variable regardless of guards, and keep
  `false` in the left part of `?:`. So `if (null !== $next) { $doc = $next; }`
  types `$next` as nullable, and `preg_split(…) ?: []` as `bool|array`
  (false positives in CallableParameterUseCaseInTypeContext, which shares
  the T-rules). custos narrows variable types by the guards on the path and
  drops `false` (with `null`) from the left part of `?:`.
- **`preg_*` replacement results may be null (custos diverges):** upstream
  types `preg_replace`, `preg_replace_callback`, `preg_filter` and
  `preg_replace_callback_array` as `string`/`array` only, so the common
  `(string) preg_replace(…)` guard against the null failure result is
  reported as redundant (and the fix removes it, letting null through).
  custos adds `null` to their R-function result, so the cast is kept. This
  also applies to CallableParameterUseCaseInTypeContext, which shares the
  T-rules.
- **Unstable variables — custos refinement, not upstream.** Value discovery ignores `++`/`--` and compound assignments upstream, so a variable later incremented or extended is analysed with its initial value only. custos makes the result unknown (M1 not reported; M2 does not use discovery and is unaffected), as in the shared value discovery of `CallableMethodValidity`. No upstream fixture relies on such a variable; recorded in `docs/internals/decisions.md` ("Spec-level false positives").
- **Integer division (custos diverges):** upstream types `int / int` as int,
  so `(int) ($a / $b)` with two ints is reported and the fix drops the cast,
  turning `7 / 2` from `3` into `3.5`. custos types the quotient as
  `{int, float}` and does not report it.
- The remaining arithmetic typing (unknown operand → float, string operand →
  float) and `microtime()` → int are upstream heuristics, not PHP semantics.
  They are reproduced because EA fixtures depend on them.
- The default variable/property typing depends on PhpStorm's inference engine;
  custos' engine only needs to match it on the patterns above (literal
  assignments, declared parameter/return/property types, `@var` on private
  properties, `new`). Where custos cannot infer a single type it must stay
  silent (empty type ⇒ no report).
- `(string)` casts as concatenation operands are reported even when the operand
  is an object without `__toString` or an array; upstream does not check.
- **Sound arithmetic (custos diverges).** Upstream's heuristic types an
  arithmetic result as float when an operand is unresolvable or a string,
  so `(float) ($cell * 100)` with `$cell` a `string|null` is reported —
  yet `"41" * 100` is the int `4100` and the cast is not redundant. custos
  types arithmetic by PHP's result rules (T-rules note above).
- **Implicitly null private properties (custos diverges).** Upstream types
  an untyped private property by `@var` plus its default; with no default
  the property still holds null until written, so `(float) $this->cout`
  over `/** @var float */ private $cout;` is not redundant (null → 0.0).
  custos adds `null` unless the constructor assigns the property (T1).
  EA cases `unnecessary-casting.php` and `unnecessary-casting.php8.php` are
  listed divergences.
- **`mb_convert_encoding()` (custos, shared T-rules).** Typed by its first
  argument: `{string, bool}` for a scalar input, `{array, bool}` for an
  array, instead of the stub's `array|string|false` (which made
  `$text = mb_convert_encoding($text, 'UTF-8')` look like an array
  assignment to CallableParameterUseCaseInTypeContext).
- **Classes named like scalar aliases (custos diverges):** T4 maps the
  normalised names `integer`/`boolean` to `int`/`bool`, so an object of a
  user class `Integer` or `Boolean` counts as already int/bool and
  `(int) new Integer()` is reported (the fix drops a cast that converts an
  object). custos resolves scalar aliases when parsing types; a remaining
  `\Integer` or `\Boolean` is a class and never matches a scalar target.
- **PHPDoc-only types (custos diverges).** A cast is reported only when the
  operand has the target type without user PHPDoc: the operand is typed a
  second time ignoring the project's `@param`, `@var` (inline and on
  properties), `@return`, templates and assertions (builtin stub types
  still count). `foreach ($ids as $id) { (int) $id; }` with `@param
  array<int, int> $ids`, or `(float) $this->rate` on a `@var float`
  property, is not reported: nothing enforces the docblock, and such casts
  guard SQL and output building (phpMyAdmin, PrestaShop review).
- **Unknown reaching definitions (custos diverges, T4).** Outside the
  spec-only typer, a variable with a reaching definition of unknown type
  (`$f = $o->x; if ($c) { $f = 'x'; } return (string) $f;`) is unknown
  instead of the union of its known definitions: the partial set made the
  cast look redundant and the fix changed behaviour (Magento config
  element and fixtures).
