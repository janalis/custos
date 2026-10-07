---
id: UnqualifiedReference
group: Performance
kind: semantic
needs: [names, index, stubs]
php: { min: "7.0", max: "" }
---

# UnqualifiedReference

## Summary
Inside a namespace, an unqualified call such as `strlen($s)` must be resolved
at run time (namespace first, then global). Writing `\strlen($s)` lets the
PHP 7+ compiler bind it at compile time and, for a set of special functions,
replace the call by a dedicated opcode. The same holds for global constants
and for function names passed as string callbacks.

## Detection
The whole rule is active only when the configured PHP level is **≥ 7.0**.

### Part A — function calls and constants
- **D1** Candidate node, one of:
  - a plain function call (not method/static/`new`) whose name as written
    (last segment, compared case-insensitively, as PHP compares function
    names) is in the *opcode list* below, or any
    function call at all when `REPORT_ALL_FUNCTIONS` is on;
  - a global constant reference (an identifier used as a constant, not a
    class constant) when `REPORT_CONSTANTS` is on.
- **D2** Constants named `true`, `false`, `null`, `__LINE__`, `__FILE__`,
  `__DIR__`, `__FUNCTION__`, `__CLASS__`, `__TRAIT__`, `__METHOD__`,
  `__NAMESPACE__` are never reported, in any letter case (PHP treats these
  names case-insensitively).
- **D3** The name is fully unqualified: no leading `\`, no `Ns\` prefix, no
  `namespace\` prefix. Any qualifier → no report.
- **D4** Governing namespace:
  - if the file declares no namespace at all → no report;
  - if the file declares exactly one namespace (braced or not) → that one;
  - if it declares several → the nearest enclosing namespace block of the
    node; if there is none → no report.
- **D5** The name resolves (with PHP's namespace fallback rules, using the
  project index and the built-in stubs) to a function (for calls) or a
  constant (for constant references) whose fully-qualified name is
  `\` + the name **as written** (compared case-insensitively for functions,
  exactly for constants). Hence:
  - a function/constant declared in the current namespace (or imported from
    another namespace) wins → FQN is namespaced → no report;
  - unknown/unresolvable names → no report;
  - a call written with different casing (`StrLen()`) resolves to the global
    `\strlen` → reported as `\StrLen(...)` (the fix keeps the written case).
- **D6** The governing namespace has a statement body, and among the `use`
  statements placed directly in that body there is no `use function X` /
  `use const X` import whose bound name (the alias when one is given,
  otherwise the last segment of the imported symbol) equals the reference
  name (case-insensitively for `use function`, exactly for `use const`):
  `use function strlen as sl;` does not hide an unqualified `strlen()`,
  while `use function Lib\fmt as strlen;` does. Class
  imports (`use Foo\Bar;`) are ignored. Names appearing inside a `use`
  statement itself are therefore never reported (they import themselves).

### Part B — string callbacks
Checked on every plain function call (qualified or not) that resolves to one
of these global functions (name compared case-insensitively; a same-named
function declared in the current namespace or imported from another
namespace does not count), at the given callback position (0-based):

| function | callback argument |
|---|---|
| `call_user_func` | 0 |
| `call_user_func_array` | 0 |
| `array_map` | 0 |
| `array_filter` | 1 |
| `array_walk` | 1 |
| `array_reduce` | 1 |

- **D7** The call has **at least 2** arguments.
- **D8** The argument at the callback position is a string literal (single or
  double quoted) without interpolation.
- **D9** Its raw contents do not start with `\` and do not contain `::`.
- **D10** `REPORT_ALL_FUNCTIONS` is on, or the contents are a name from the
  opcode list (compared case-insensitively).
- **D11** The callback contents name a **known global function**: a
  function of that fully-qualified name (no namespace) exists in the project
  index or the built-in stubs (function names compared case-insensitively).
  Unknown names are not reported. Part B does **not** require a namespace
  (D4 is not applied) — see Divergences.

### Opcode list
`array_slice`, `assert`, `boolval`, `call_user_func`, `call_user_func_array`,
`chr`, `count`, `defined`, `doubleval`, `floatval`, `func_get_args`,
`func_num_args`, `get_called_class`, `get_class`, `gettype`, `in_array`,
`intval`, `is_array`, `is_bool`, `is_double`, `is_float`, `is_int`,
`is_integer`, `is_long`, `is_null`, `is_object`, `is_real`, `is_resource`,
`is_string`, `ord`, `strlen`, `strval`, `function_exists`, `is_callable`,
`extension_loaded`, `dirname`, `constant`, `define`, `array_key_exists`,
`is_scalar`, `sizeof`, `ini_get`, `sprintf`, `printf`.

A call that is both in the opcode list and in the Part B table (e.g. an
unqualified `call_user_func('is_int', $v)` in a namespace) can produce two
reports: one on the call (Part A) and one on the callback string (Part B).

## Exceptions (no report)
- **E1** PHP level below 7.0.
- **E2** Qualified names (`\strlen()`, `Sub\strlen()`, `namespace\strlen()`).
- **E3** File without any namespace declaration (Part A only).
- **E4** Name resolving to a non-global function/constant, or unresolved.
- **E5** Name imported with `use function` / `use const` in the namespace body.
- **E6** Calls outside the opcode list while `REPORT_ALL_FUNCTIONS` is off;
  constants while `REPORT_CONSTANTS` is off.
- **E7** Callback strings already qualified (`'\trim'`), static-method
  callbacks (`'Cls::m'`), interpolated strings, non-string callbacks
  (closures, arrays, variables), calls with fewer than 2 arguments, callback
  names that are not a known global function (`'no_such_fn'`).
- **E8** `true`/`false`/`null` and magic constants (D2).

## Report
- Range: Part A — the whole call (from the name to the closing `)`) or the
  constant identifier; Part B — the callback string literal including its
  quotes.
- Severity: info (weak warning).
- Message: `Write '\{name}' to allow compile-time binding.` where `{name}` is
  the function name followed by `(...)` for calls, the constant name for
  constants, or the callback string contents (no `(...)`) for Part B.

## Fix
- **F1** Part A: insert a single `\` immediately before the name
  (`strlen($s)` → `\strlen($s)`, `PHP_EOL` → `\PHP_EOL`). Nothing else changes.
- **F2** Part B, single-quoted literal: `'name'` → `'\name'`.
- **F3** Part B, double-quoted literal: `"name"` → `"\\name"` (escaped
  backslash). Contents are otherwise kept verbatim.

## Options
| Option | Type | Default | Effect |
|---|---|---|---|
| `REPORT_ALL_FUNCTIONS` | bool | false | Report every unqualified global function call (and every callback name), not only those in the opcode list. |
| `REPORT_CONSTANTS` | bool | false | Also report unqualified references to global constants. |

The rule is disabled by default.

## PHP versions
- Requires level ≥ 7.0 for everything (both parts). Upstream fixtures run at
  7.1.

## Examples
With `REPORT_ALL_FUNCTIONS = true`, `REPORT_CONSTANTS = true`, level 7.1:

```php
<?php
namespace {
    function legacy_hook() {}
}

namespace Shop\Cart {
    use function array_sum;
    use const E_ALL;

    function money($v) { return $v; }
    const SCALE = 2;

    $n = <weak_warning descr="Write '\count(...)' to allow compile-time binding.">count([1])</weak_warning>;
    $u = <weak_warning descr="Write '\ucfirst(...)' to allow compile-time binding.">ucfirst('x')</weak_warning>;
    $e = <weak_warning descr="Write '\PHP_EOL' to allow compile-time binding.">PHP_EOL</weak_warning>;
    \array_map(<weak_warning descr="Write '\legacy_hook' to allow compile-time binding.">'legacy_hook'</weak_warning>, []);
    \array_walk($list, <weak_warning descr="Write '\legacy_hook' to allow compile-time binding.">"legacy_hook"</weak_warning>);

    $ok = \count([1]) + money(1) + SCALE + array_sum([]) + E_ALL;
    $ok2 = true && null === __LINE__;
    \array_map('\legacy_hook', []);
    \array_map('Helper::run', []);
    \usort($list, 'legacy_hook');
}
```

```php
<?php
namespace {
    function legacy_hook() {}
}

namespace Shop\Cart {
    use function array_sum;
    use const E_ALL;

    function money($v) { return $v; }
    const SCALE = 2;

    $n = \count([1]);
    $u = \ucfirst('x');
    $e = \PHP_EOL;
    \array_map('\legacy_hook', []);
    \array_walk($list, "\\legacy_hook");

    $ok = \count([1]) + money(1) + SCALE + array_sum([]) + E_ALL;
    $ok2 = true && null === __LINE__;
    \array_map('\legacy_hook', []);
    \array_map('Helper::run', []);
    \usort($list, 'legacy_hook');
}
```

Defaults (both options off), level 7.1, no namespace — nothing is reported:

```php
<?php
$size = strlen($s) + \strlen($s);
$max  = PHP_INT_MAX;
define('LIMIT', 3);
```

## Divergences
- **Part B ignores the namespace context (upstream):** callback strings are
  reported even in files without a namespace, where qualifying makes no
  difference. Recommendation: apply D4 to Part B too (skip when the file has no
  namespace). Not covered by upstream fixtures.
- **D11 — custos diverges from upstream.** Upstream checks that the *outer*
  function (`array_map`, …) exists instead of the callback, so with
  `REPORT_ALL_FUNCTIONS` any callback string is reported, including names
  of functions that do not exist (typos, functions defined only at run
  time); the claimed compile-time binding cannot apply to them. custos
  looks the callback name up and reports only known global functions.
- **Empty callback strings** (`array_map('', $a)` with
  `REPORT_ALL_FUNCTIONS`) would be reported as `'\'`. Recommendation: skip
  empty contents.
- **Global namespace block:** in a file with several namespaces, a call inside
  `namespace { … }` (the global block) is reported although it already binds
  globally. Recommendation: skip when the governing namespace is the global
  one. Not covered by fixtures.
- **Letter case and resolution — custos diverges from upstream.** Upstream
  compares function names case-sensitively: `StrLen($s)` in a namespace is
  not reported although PHP resolves it to `\strlen` at run time just the
  same, `'Is_Int'` callbacks are skipped, `use function COUNT;` does not
  count as an import of `count`, and mixed-case `True`/`Null` are only
  skipped by accident. Part B also matches the outer function by its written
  name, so a namespaced user `array_map()` was treated as the builtin. custos
  compares function names (and `true`/`false`/`null`/magic constants)
  case-insensitively and requires Part B's outer call to reach the global
  function; constant names stay case-sensitive.
- **Import aliases — custos diverges from upstream.** Upstream compares the
  reference with the last segment of the imported symbol and ignores the
  alias, so `use function strlen as sl;` silences every unqualified
  `strlen()` (which still goes through the runtime fallback), while
  `use function Lib\fmt as strlen;` lets `strlen()` be reported although it
  calls `Lib\fmt`. custos compares with the name the import binds (D6).
