---
id: NonSecureUniqidUsage
group: Security
kind: semantic
needs: [names, index, stubs]
php: { min: "", max: "" }
---

# NonSecureUniqidUsage

## Summary
Without its `more_entropy` argument, `uniqid()` is based only on the current
time in microseconds and collides easily. Pass `true` as second argument,
also when `uniqid` is used as a callback.

## Detection
### Direct calls
- **D1** A plain function call whose name (compared case-insensitively) is
  `uniqid` and which resolves to the global function `\uniqid` (an
  unqualified call in a namespace falls back to the global function unless
  that namespace declares its own `uniqid`; `use function uniqid;` keeps it
  global).
- **D2** No `more_entropy` argument is given: there is no positional
  argument at index 1 and no named argument `more_entropy:`.

### String callbacks
- **D3** A plain function call resolving to one of the global functions
  below (names compare case-insensitively; a same-named namespaced function
  does not count) with **at least 2**
  arguments; the callback argument `A` is taken at the listed index:

  | Function | Callback index |
  |---|---|
  | `call_user_func` | 0 |
  | `call_user_func_array` | 0 |
  | `array_map` | 0 |
  | `array_filter` | 1 |
  | `array_reduce` | 1 |
  | `array_walk` | 1 |
  | `array_walk_recursive` | 1 |

- **D4** `A` resolves to a string literal: `A` itself if it is a string
  literal, otherwise value discovery (as defined in the
  `CallableMethodValidity` spec) on `A` must yield exactly one string
  literal among its values.
- **D5** The literal's content, with string escapes decoded per its quote
  style, and with one leading `\` removed if present, equals `uniqid`
  compared case-insensitively (`'UNIQID'` calls the same function).

## Exceptions (no report)
- **E1** `uniqid('', true)`, `uniqid($p, false)` (any second argument),
  `uniqid(more_entropy: true)`.
- **E2** A namespaced call that resolves to a user-defined `uniqid` (not
  the global one).
- **E3** Callback functions called with fewer than 2 arguments
  (`call_user_func('uniqid')`); callbacks given as closures, arrays or
  other names.

## Report
- Range: D1 — the whole call, from the name (including a leading `\` if
  written) to the closing `)`; D3 — the callback argument `A` as written
  (a string literal including its quotes, or the variable that was
  resolved).
- Severity: error.
- Message: `Pass more_entropy = true to uniqid() to reduce collisions.`

## Fix
- **F1** Direct call: replace the argument list (from `(` to `)`
  inclusive) with a new list built from the template `('', true)`: each
  existing argument replaces the template argument at the same position
  (source text copied verbatim). Name and qualifier are kept.
  - `uniqid()` → `uniqid('', true)`
  - `uniqid('ord_')` → `uniqid('ord_', true)`
  - `\uniqid($p)` → `\uniqid($p, true)`
- **F2** Callback, only for the callback-index-0 functions
  (`call_user_func`, `call_user_func_array`, `array_map`), where the
  callback's first parameter is what `uniqid()` would receive as its prefix:
  replace `A` with the closure
  `function ($value) { return uniqid($value, true); }`. (Upstream's
  formatter outputs it on three lines; comparison is whitespace-collapsed,
  so keep exactly one space between `function` and `(`, between `)` and
  `{`, after `{` and before `}`.)
  Builtin spelling: `uniqid` is written `\uniqid` when an unqualified call at
  that position would not reach the global function (a `use function`
  import under that name, or a same-named function declared in the current
  namespace) inside the closure: in `namespace Ids; function uniqid() {…}`, `array_map('uniqid', $l)` → `array_map(function ($value) { return \uniqid($value, true); }, $l)` (the string callable always names the global function).
  For `array_filter`, `array_reduce`, `array_walk` and
  `array_walk_recursive` the report is made without a fix: these pass other
  arguments (`$carry, $item`, `$value, $key`, …), so replacing the callback
  with a one-parameter closure would change the values `uniqid()` is called
  with.

## Options
None.

## PHP versions
No gating. (Named arguments are parsed regardless of the configured level.)

## Examples

```php
<?php
namespace Shop {
    $id   = <error descr="Pass more_entropy = true to uniqid() to reduce collisions.">uniqid()</error>;
    $ref  = <error descr="Pass more_entropy = true to uniqid() to reduce collisions.">\uniqid('ord_')</error>;
    $tags = array_map(<error descr="Pass more_entropy = true to uniqid() to reduce collisions.">"uniqid"</error>, ['a', 'b']);
    $one  = call_user_func_array(<error descr="Pass more_entropy = true to uniqid() to reduce collisions.">'\\uniqid'</error>, ['x']);

    $ok1 = uniqid('', true);
    $ok2 = uniqid(more_entropy: true, prefix: 'z');
    $ok3 = call_user_func('uniqid');
    $ok4 = array_map('strtoupper', ['a']);
}
```

```php
<?php
namespace Shop {
    $id   = uniqid('', true);
    $ref  = \uniqid('ord_', true);
    $tags = array_map(function ($value) { return uniqid($value, true); }, ['a', 'b']);
    $one  = call_user_func_array(function ($value) { return uniqid($value, true); }, ['x']);

    $ok1 = uniqid('', true);
    $ok2 = uniqid(more_entropy: true, prefix: 'z');
    $ok3 = call_user_func('uniqid');
    $ok4 = array_map('strtoupper', ['a']);
}
```

## Divergences
- **Name matching (custos diverges from upstream).** Upstream compares the
  direct call name, the callback-taking function name and the callback
  string case-sensitively, and does not resolve the callback-taking
  function. custos matches all three case-insensitively (as PHP resolves
  function names and string callables) and requires the callback-taking
  call to resolve to the global function, so `Array_Map('UNIQID', $l)` is
  reported and a namespace's own `array_map('uniqid', $l)` is not.
- F1 with a named or spread argument (`uniqid(prefix: 'q')`,
  `uniqid(...$args)`) would yield `uniqid(prefix: 'q', true)` /
  `uniqid(...$args, true)`, which is invalid. Recommendation: still report,
  but for named arguments append `more_entropy: true`, and offer no fix for
  spread arguments. Not covered by fixtures.
- F2: custos diverges from upstream for the callback-index-1 functions.
  Upstream replaces the string callback with the one-parameter closure for
  every listed function, but `array_reduce` calls its callback with
  `($carry, $item)` and `array_walk`/`array_walk_recursive` with
  `($value, $key)`, so the original `uniqid` received a second argument the
  closure discards; the rewrite changes the computation rather than only
  adding entropy. custos still reports these callbacks but offers no fix.
- **Builtin spelling (custos diverges).** Upstream inserts a bare `uniqid(`,
  which a function of that name declared in (or imported into) the current
  namespace captures, so the fix would call user code. custos writes
  `\uniqid(` in that case.
