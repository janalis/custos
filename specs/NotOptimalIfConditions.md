---
id: NotOptimalIfConditions
group: Control flow
kind: semantic
needs: [names, types, hierarchy, index, stubs]
php: { min: "", max: "" }
---

# NotOptimalIfConditions

## Summary

Looks at the conditions of `if`/`elseif` statements and flags four kinds of
problems: a cheap operand evaluated after an expensive one in an `&&`/`||`
chain (short-circuiting would save work if swapped), `and`/`or` keyword
operators (low precedence, error-prone), an equality comparison on a value
that the same `&&` chain also tests with `instanceof` (likely a logic bug),
and an `instanceof` check made redundant by another `instanceof` on the same
subject against a related type.

## Detection

Applies to each `if` statement: its own condition, then each of its `elseif`
conditions (in source order). An `else if` is a nested `if` visited on its
own. Conditions of loops, ternaries, `match` etc. are not inspected.

### Splitting a condition into operands

For a condition expression `K`:

1. Remove wrapping parentheses.
2. If the result is a prefix unary expression (any unary: `!`, `-`, `@`,
   cast, …), take its operand and remove wrapping parentheses again (only one
   unary level is unwrapped).
3. If the result is not a binary `&&` or `||` (symbolic only — keyword
   `and`/`or` are **not** split), the operand list is that single expression.
4. Otherwise, with `OP` = that operator, the list is obtained by flattening
   left-nested chains of the same `OP`: the right operand (parentheses
   stripped) is last; then repeatedly the left operand (parentheses
   stripped) is examined — while it is a binary with the same `OP`, its right
   operand (stripped) is prepended and its left operand continues. The final
   left operand (stripped) is first. A parenthesised right-nested group
   (`$a && ($b && $c)`) is **not** flattened: it is one operand. Mixed
   operators: `$a && $b || $c` → `[$a && $b, $c]` with `OP = ||`.
   `OP` is remembered as the chain operator for this condition.

### D1 — ordering (option SUGGEST_OPTIMIZING_CONDITIONS)

With the operand list `O1…On` (n ≥ 2), compute `cost(Oi)` (below). For each
`i ≥ 2`: report `Oi` when `cost(Oi) < cost(Oi-1)` and `Oi` is **not coupled**
with `Oi-1` (below). Comparison is only between neighbours. When either
neighbour's cost is *unknown* (an unknown part anywhere inside it, see
*Property fetches*), the pair is not compared and nothing is reported for
it.

**Cost** (`cost(e)`, `e` with parentheses stripped first):

- 0: missing expression, constant reference (`true`, `null`, `PHP_EOL`,
  magic constants), class reference, class constant access (`A::B`,
  `A::class`, `static::X`), numeric literal.
- variable, string literal (incl. interpolated/heredoc), static property
  access (`A::$p`), and an instance property access (`$o->p`, `$o?->p`)
  classified *stored* (below): sum of the costs of its embedded
  sub-expressions — e.g. the object expression of a property access, the
  expression inside `${…}` / `$o->{…}`, interpolated expressions in a string.
  Plain `$x`, `'text'`, a stored `$o->p` cost 0; `"{$o->m()}"` costs 5.
- instance property access classified *computed* (below): costs like a
  method call — cost of the object expression + 5.
- instance property access classified *unknown* (below): its cost is
  unknown, and so is the cost of every expression containing it.
- array access `b[i]`: `cost(b) + cost(i)`.
- `empty(...)` / `isset(...)`: sum of argument costs.
- function call: sum of argument costs, `+5` unless the call resolves to
  one of the global functions below (name compared case-insensitively like
  PHP function names; resolved like any builtin match: `\is_array()` or an
  unqualified call with PHP's global fallback count, while `Lib\is_array()`,
  a `use function` import of another function under that name, or an
  unqualified call in a namespace declaring that function costs 5 like any
  user call):
  `array_key_exists function_exists property_exists class_exists
  interface_exists trait_exists is_a is_subclass_of defined is_array is_bool
  is_callable is_countable is_float is_double is_real is_int is_integer
  is_long is_iterable is_null is_numeric is_object is_resource is_scalar
  is_string`. (Note `method_exists` is not in the list.)
- method call (instance or static, incl. nullsafe): sum of argument costs +
  cost of the object/class expression + 5.
- unary expression (`!`, `-`, casts, `@`, `++`/`--`): cost of its operand.
- binary expression (incl. comparisons, `instanceof`, `.`, `??`): left +
  right.
- array literal: sum of costs of all keys and values.
- assignment (incl. compound `.=`, `+=`…): cost of the assigned value only.
- ternary: `cost(cond) + max(cost(then), cost(else))` (missing middle part of
  `?:` costs 0).
- anything else (`new`, closures, `include`, `print`, `match`, `clone` if not
  modelled as unary, …): 10.

**Property fetches.** Reading a property is cheap only when it reads a
stored slot. Since PHP 8.4 a property may carry a `get` hook (arbitrary code,
which can e.g. initialise a lazy collection and run a database query), and on
any version a read of an undeclared property can land in `__get()`. Each
instance property access `R->name` / `R?->name` with an identifier name is
classified from the inferred type of `R`:

- *stored*: every class of `R`'s type set resolves in the index, and in each
  of them `name` resolves (own or inherited) to a declared property that is
  not virtual and has no `get` hook (a `set`-only hook on a backed property
  keeps it stored: reads do not run code); or `name` is not declared and
  neither the class nor any ancestor declares `__get()` (a dynamic property,
  e.g. on `\stdClass`).
- *computed*: in at least one class of the set, `name` resolves to a
  property with a `get` hook, a virtual property, or a property declared
  only by an interface or as `abstract` (an implementation may hook it); or
  `name` is not declared and the class or an ancestor declares `__get()`.
- *unknown*: anything else — `R`'s type is missing, contains an unknown or
  unresolved part, `mixed` or `object`, or the name is dynamic
  (`$o->{$k}`, `$o->$k`).

Static property access (`A::$p`) cannot be hooked nor reach `__get()` and is
always stored.

**Coupling** (`prev` = `Oi-1`, `cur` = `Oi`); coupled if any scenario holds:

- **S1 (mutation)**: collect the "mutated" expressions of `prev`:
  - the target of every assignment in `prev` (including `prev` itself when it
    is an assignment, nested assignments, compound assignments); for
    `list(...)`/`[...] =` destructuring, every target variable;
  - every argument that is a plain variable passed to a by-reference
    parameter (same position) of a call in `prev` (including `prev` itself),
    when the called function/method resolves (user code or stubs, e.g.
    `array_shift`, `preg_match`). Only calls having at least one plain
    variable argument are resolved.
  Coupled if `cur` itself, or any sub-expression of `cur`, is of the same node
  kind and equivalent (see *Equivalence*) to a mutated expression.
- **S2 (array access)**: collect from `cur` (itself and all sub-expressions)
  every array access that is not the base of an enclosing array access
  (outermost of a chain). For each such access, collect the base chain: for
  `$a['x']['y']` that is `$a['x']` and `$a` (not the full access itself).
  Coupled if some proper sub-expression of `prev` (not `prev` itself) of the
  same node kind is equivalent to one of these, ignoring candidates in `prev`
  that are themselves the base of an array access.
- **S3 (isset guard)**: for every `isset(...)` in `prev` (itself or nested),
  take each array access inside the isset arguments and its innermost base;
  if that base is a variable, remember its name. Coupled if `cur` itself or
  any sub-expression of `cur` is a variable with one of those names.
  `isset()` also guards the existence of the variable itself: moving `$x`
  before `isset($x[...])` can raise "Undefined variable".

- **S4 (side effects — custos divergence, see Divergences)**: coupled if
  `prev` or `cur` (itself or any sub-expression, not descending into closure
  / arrow-function bodies) contains an *impure* construct, because swapping
  would change whether that effect happens:
  - a function call that is not *pure* (below): every user-defined function,
    every unresolved or dynamic call (`$f()`, `call_user_func(...)`), and
    every built-in not in the purity list — in particular output, I/O,
    filesystem, network, process and global-state functions such as
    `fwrite fputs fread fgets fclose fopen fflush ftruncate flock
    file_put_contents file_get_contents readfile unlink rename copy mkdir
    rmdir chmod chown touch tempnam move_uploaded_file header setcookie
    http_response_code printf print_r var_dump error_log trigger_error mail
    curl_exec exec shell_exec system ini_set set_error_handler putenv
    setlocale define array_pop array_shift sort next reset`, and any
    `session_*`, `curl_*`, `ob_*`, `socket_*`, `ftp_*`, `stream_*`, `posix_*`,
    `pcntl_*`, `mysqli_*`, `pg_*` function;
  - a call that passes an argument at a by-reference parameter position of
    the resolved function, even when the function is in the purity list
    (`preg_match($re, $s, $m)` is impure, `preg_match($re, $s)` is pure);
  - `include` / `require` (`_once`), `eval`, `exit` / `die`, `print`;
  - every method call (`->`, `?->`), static method call (`A::m()`,
    `parent::m()`, …) and object creation (`new X(...)`, which runs a
    constructor): their targets run arbitrary user code (or magic
    `__call`/`__get`), and there is no safe subset that can be recognised
    from the call alone;
  - every instance property access classified *computed* (see *Property
    fetches*): a `get` hook or `__get()` runs user code just like a method.

  **Purity list** (function name matched case-insensitively, without
  namespace, and only when it resolves to the built-in — a same-named user
  function in the current namespace is impure):
  - every function of the cost-0 list above (`array_key_exists`, `is_*`,
    `defined`, `*_exists`, `is_a`, `is_subclass_of`, …) and `method_exists`,
    `key_exists`;
  - counting/strings: `count sizeof strlen mb_strlen trim ltrim rtrim chop
    strtolower strtoupper mb_strtolower mb_strtoupper ucfirst lcfirst ucwords
    substr mb_substr substr_count strpos stripos strrpos strripos mb_strpos
    mb_stripos str_contains str_starts_with str_ends_with strstr stristr
    strrchr str_pad str_repeat strrev str_split mb_str_split sprintf vsprintf
    implode join explode strcmp strcasecmp strncmp strncasecmp strnatcmp
    strnatcasecmp str_replace str_ireplace nl2br number_format wordwrap
    htmlspecialchars htmlentities html_entity_decode strip_tags addslashes
    stripslashes quotemeta preg_quote preg_match preg_match_all preg_replace
    preg_split preg_grep` and every `ctype_*` function;
  - arrays (no callbacks): `in_array array_search array_keys array_values
    array_merge array_merge_recursive array_replace array_slice array_flip
    array_unique array_reverse array_combine array_fill array_fill_keys
    array_pad array_chunk array_diff array_diff_key array_diff_assoc
    array_intersect array_intersect_key array_column array_sum array_product
    array_count_values array_key_first array_key_last array_is_list range`;
  - numbers/conversion: `abs min max floor ceil round intdiv fmod sqrt pow
    intval floatval doubleval boolval strval gettype get_debug_type
    is_nan is_finite is_infinite`;
  - classes/objects: `get_class get_parent_class get_object_vars
    get_class_methods class_implements class_parents constant`;
  - encoding/hashing: `md5 sha1 crc32 hash base64_encode base64_decode
    bin2hex hex2bin json_encode json_decode urlencode urldecode rawurlencode
    rawurldecode http_build_query parse_url filter_var version_compare`;
  - time/ids (no observable side effect): `time microtime hrtime date gmdate
    mktime strtotime checkdate uniqid`;
  - filesystem queries (read-only): `file_exists is_file is_dir is_link
    is_readable is_writable is_writeable is_executable filesize filemtime`.

**Equivalence**: same node kind and, for variables, same name (names
non-empty); otherwise same structure/token sequence ignoring whitespace and
comments, or identical source text.

### D2 — keyword logical operators (option REPORT_LITERAL_OPERATORS)

For the whole condition expression of the `if`/`elseif` (no splitting), visit
it and every nested binary expression at any depth (including inside calls,
closures, nested parentheses). Each binary whose operator token is `and`
(case-insensitive) is reported (suggest `&&`); `or` is reported (suggest
`||`). `xor` is not reported.

### D3 — equality next to instanceof (option REPORT_INSTANCE_OF_FLAWS)

Applies when the chain operator of this condition is `&&` and the operand list
has ≥ 2 entries. Let `subj` be the left operand of the **first** operand (in
list order) that is an `instanceof` binary. If there is one, report every
operand (anywhere in the list, before or after) that is a binary with
operator `==`, `!=`/`<>`, `===` or `!==` whose left or right operand is
equivalent to `subj`. The other operand must be a literal value (parentheses
stripped, optionally signed): a number or string literal (interpolated
strings included), `null`/`true`/`false`, or an array literal (custos
diverges: comparing `subj` with any other expression, e.g. `$a instanceof X
&& $a !== $b`, is an ordinary identity check and is not reported).

### D4 — redundant instanceof (option REPORT_INSTANCE_OF_FLAWS)

Applies when the operand list has ≥ 2 entries (chain operator `&&` or `||`).

1. Take every operand that is an `instanceof` binary whose right side is a
   class name that resolves to a class/interface/trait. Others are ignored.
2. Group them by subject (left operand), using *Equivalence*.
3. In each group with ≥ 2 entries, for each entry `A` (class `Ca`), compute
   the ancestry of `Ca`: `Ca` itself, all its interfaces (transitively),
   traits it uses (transitively), parent classes and their interfaces/traits;
   for an interface: itself and its parent interfaces. For every **other**
   entry `B` of the group whose class `Cb` is contained in that ancestry
   (`Ca` is a subtype of `Cb`), the redundant check is
   - under `||`: `A`, the more specific check (whenever it is true, `B` is
     true as well, so `A` adds nothing);
   - under `&&`: `B`, the broader check (whenever `A` holds, `B` holds too).
   Each redundant entry is reported once. Exception: a `Cb` equal to
   `\DateTimeInterface` is ignored when the configured PHP level is below
   5.5 (other entries still checked). Two checks against the same class are
   both reported.

- **Name case.** Wherever this rule compares two expressions for
  equivalence (coupling checks S1–S3, the D3 equality subject, the D4
  instanceof subjects), the names PHP resolves case-insensitively —
  function and method names, class names in calls, `new`, `instanceof` and
  `::` accesses, and keywords — compare case-insensitively
  (`Repo::$node` matches `repo::$node`); variable, property and constant
  names stay case-sensitive.

## Exceptions (no report)

- **E1** Neighbouring operands in non-increasing order, or equal cost.
- **E2** Coupled neighbours (S1–S3), e.g. `($n = $c->count()) && $n > 0`,
  `count($a) > 0 && $a[0]`, `($x = array_shift($q)) && $q`,
  `!isset($m[f($k)]) && !array_key_exists($k, $m)`,
  `isset($rows[md5($k)]) && $rows`, and neighbours with a side effect (S4):
  `fwrite($h, $line) && $ok`, `mkdir($dir) || $dryRun`,
  `rebuild($cache) && $fresh` (user function), `$repo->save($e) || $quiet`,
  `Cache::clear() || $quiet` (method calls).

- **E2a** Neighbours where one side reads a property whose cost cannot be
  established (*unknown*: untyped receiver, dynamic name), and neighbours
  where one side reads a hooked, virtual, interface/abstract or `__get()`
  property (*computed*, impure under S4), e.g. `'' === $slug ||
  filter_var($slug, FILTER_VALIDATE_URL) || !$page->hasContent` when
  `hasContent` has a `get` hook.
- **E3** Keyword `and`/`or` chains are a single operand for D1/D3/D4 (still
  reported by D2).
- **E4** D3 when the chain is `||`, or no `instanceof` operand exists.
- **E5** D4 when classes are unrelated or unresolvable, or right side is a
  variable/expression.

## Report

| Check | Range | Severity |
|-------|-------|----------|
| D1 | the reported operand, parentheses stripped (inner expression only) | info (weak warning) |
| D2 | the `and` / `or` operator token | info (weak warning) |
| D3 | the whole equality operand (parentheses stripped) | info (weak warning) |
| D4 | the whole `instanceof` operand (parentheses stripped) | warning (rule default) |

Messages (our wording):

- D1: `Cheaper check placed after a costlier one; evaluate it first.`
- D2: `Use '&&' instead of 'and'.` / `Use '||' instead of 'or'.`
- D3: `Equality check on a value also tested with instanceof; verify the logic.`
- D4: `Redundant instanceof: another check on the same value already covers this type.`

## Fix

None.

## Options

| Option | Type | Default | Effect |
|--------|------|---------|--------|
| REPORT_LITERAL_OPERATORS | bool | true | enables D2 |
| REPORT_INSTANCE_OF_FLAWS | bool | true | enables D3 and D4 |
| SUGGEST_OPTIMIZING_CONDITIONS | bool | true | enables D1 |

Upstream tests set one option to `true` without disabling the others, so each
fixture is effectively evaluated with all checks on.

## PHP versions

- D4: `\DateTimeInterface` as ancestor is ignored below PHP 5.5. Upstream
  tests run at the PhpStorm test default (7.x) except one fixture at 5.4.
- Otherwise no gating.

## Examples

```php
<?php
// D1 ordering
if (strlen($name) > 3 && <weak_warning descr="Cheaper check placed after a costlier one; evaluate it first.">$enabled</weak_warning>) {}
if (\json_encode($id) || <weak_warning descr="Cheaper check placed after a costlier one; evaluate it first.">empty($cache)</weak_warning>) {}
if (isset($rows[md5($k)]) && <weak_warning descr="Cheaper check placed after a costlier one; evaluate it first.">$limit</weak_warning>) {}
if (!(trim($v) && (<weak_warning descr="Cheaper check placed after a costlier one; evaluate it first.">$v</weak_warning>))) {}
if ($ok) {} elseif (json_decode($raw) || <weak_warning descr="Cheaper check placed after a costlier one; evaluate it first.">is_string($w)</weak_warning>) {}
if (preg_match('/^v\d/', $tag) && <weak_warning descr="Cheaper check placed after a costlier one; evaluate it first.">$strict</weak_warning>) {}
if ($enabled && strlen($name) > 3) {}
if (is_int($n) && $n) {}
if (($total = $cart->sum()) && $total > 10) {}
if (count($items) > 0 && $items[0]) {}
if (($head = array_pop($stack)) && $stack) {}
if (!isset($map[lower($key)]) && !array_key_exists($key, $map)) {}
// S4: side effects are never reordered
if (fwrite($log, $line) !== false && $verbose) {}
if (rename($tmp, $target) || $force) {}
if (session_start() && $user) {}
if (preg_match('/^v(\d+)/', $tag, $parts) && $strict) {}
if (recompute($totals) > 0 && $dirty) {}       // user function: impure
if (strlen(file_get_contents($path)) && $ok) {}
if ($repo->find($id) || empty($cache)) {}       // method call: impure
if (isset($rows[md5($k)]) && $rows) {}          // S3: $rows itself is guarded
if (fetch($a) <weak_warning descr="Use '&&' instead of 'and'.">and</weak_warning> $b) {} // one operand for D1

// property fetches
final class Folder {
    public bool $open = false;
    public array $files = [];
    public bool $isEmpty { get => $this->files === []; }
    public function __construct(public ?string $label = null) {}
}
final class Lazy {
    public function __get(string $n) { return load($n); }
}
function scan(Folder $f, Lazy $l, $any, string $p) {
    if (strlen($p) > 3 && <weak_warning descr="Cheaper check placed after a costlier one; evaluate it first.">$f->open</weak_warning>) {}
    if (trim($p) || <weak_warning descr="Cheaper check placed after a costlier one; evaluate it first.">$f->label</weak_warning>) {}
    if (strlen($p) > 3 && !$f->isEmpty) {}      // get hook: computed, impure
    if (is_dir($p) && $l->cached) {}            // __get(): computed, impure
    if (strlen($p) > 3 && $any->flag) {}        // untyped receiver: unknown
    if ($f->isEmpty || $f->open) {}             // computed neighbour: impure
}

// D2 keyword operators
if ($p <weak_warning descr="Use '&&' instead of 'and'.">AND</weak_warning> $q) {}
if ($p || ($q <weak_warning descr="Use '||' instead of 'or'.">or</weak_warning> $r)) {}
if ($p xor $q) {}

// D3 equality next to instanceof
if (<weak_warning descr="Equality check on a value also tested with instanceof; verify the logic.">$node != null</weak_warning> && $node instanceof Leaf && $node->ok) {}
if ($node instanceof Leaf || $node === null) {}

// D4 redundant instanceof
interface Shape {}
class Circle implements Shape {}
class Ring extends Circle {}
if ($s instanceof Shape || <warning descr="Redundant instanceof: another check on the same value already covers this type.">$s instanceof Ring</warning>) {}
if ($s instanceof Circle && <warning descr="Redundant instanceof: another check on the same value already covers this type.">$s instanceof Shape</warning>) {}
if ($s instanceof Circle || $t instanceof Shape) {}
```

(No fix output: the rule has no quick-fix.)

## Divergences

- **D3 identity checks (custos diverges from upstream).** Upstream reports
  any equality operand on the instanceof subject, so the common
  `$other instanceof User && $other !== $current` ("another user than this
  one") is flagged as a logic flaw. Only comparisons with a literal (null,
  booleans, numbers, strings, array literals) are redundant or suspicious
  next to `instanceof`; custos reports only those.
- **S4 side-effect coupling (custos diverges from upstream).** Upstream
  suggests moving a cheaper operand before any costlier one, even when the
  costlier one writes a file, closes a handle, sends a header or changes
  permissions (`fwrite(...) && $ok`, `chmod(...) || $quiet`); following the
  advice changes whether the effect happens (found on real code). custos only
  reorders across calls to functions from an explicit purity list and treats
  every other function call (including user functions), by-reference
  arguments, `include`/`eval`/`exit`/`print`, and every method call, static
  call and `new` as non-reorderable. Method calls were first kept reorderable
  only because an upstream fixture expects `$obj->count() > 0 || isset($b)`
  to be reported; a method may persist, log or mutate state just like a
  function, and nothing in the call identifies a side-effect-free subset, so
  that upstream case is now an intentional divergence. Recorded in
  `docs/internals/decisions.md` ("Spec-level false positives").
- **Redundant check under `&&` (custos diverges from upstream).** Upstream's
  D4 always reports the more specific `instanceof`, which is right for `||`
  but backwards for `&&`: in `$s instanceof Circle && $s instanceof Shape`
  the `Shape` check is the redundant one, and removing the reported
  `Circle` check (as upstream suggests) would change the condition. custos
  reports the broader check under `&&`.
- **isset guard on the variable itself (custos diverges from upstream).**
  Upstream's S3 ignores `cur` when it is exactly the guarded variable, so
  `isset($x[f()]) && $x` is told to evaluate `$x` first, which raises
  "Undefined variable" when `$x` is not set — the case the `isset()` guards
  against. custos counts `cur` itself (S3).
- **Case of function names in the cost model (custos diverges from
  upstream).** Upstream looks up cheap function names case-sensitively, so
  `IS_ARRAY($x)` costs 5 while `is_array($x)` costs 0: `IS_ARRAY($x) && $y`
  is reported and `strlen($s) > 3 && IS_ARRAY($x)` is not. PHP function names
  are case-insensitive; custos compares them case-insensitively.
- A rarely relevant quirk: the remembered chain operator is not reset between
  the `if` and its `elseif`s, but it is only consulted when a condition has
  ≥ 2 operands, which always sets it — no observable effect.
- **Resolution of cheap function names (custos diverges from upstream).**
  Upstream looks up the written name without its namespace prefix, so a
  user function such as `Lib\is_array()` (or an `is_array()` declared in the
  current namespace) is treated as a cheap builtin. custos looks up the
  resolved global function name; user functions cost like any call. This
  only affects the suggested evaluation order.
- **Property reads with hooks or `__get()` (custos diverges).** Upstream
  prices every `$o->p` like a variable (free), so a read that runs a PHP 8.4
  `get` hook or `__get()` — possibly a lazy-loaded Doctrine collection, i.e.
  a query — is called cheaper than a preceding `filter_var()` and is
  suggested to move first (found on real code). custos classifies property
  reads (*Property fetches*): only stored properties are free; hooked,
  virtual, interface/abstract and `__get()` reads cost like a method call
  and are not reordered (S4); reads whose receiver type is unknown are left
  out of the comparison rather than guessed at. Upstream's own edge case
  (`... || $object->field || Clazz::STATE`, untyped `$object`) stays silent
  in both.
- **Name case (custos diverges).** Upstream compares the expressions
  textually, so subjects that differ only in the case of a function, method
  or class name (`Repo::$node` vs `repo::$node`), which PHP treats as the
  same, are not recognised as equivalent. custos folds the case of those
  names (Detection, "Name case").
