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
  arguments; the callback argument `A` is the positional argument at the
  listed index, or the argument named `callback:`:

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
- **E4** Callbacks through which `more_entropy` is already forwarded (the
  counterpart of E1): `call_user_func` with three or more positional
  arguments or a named `more_entropy:`; `array_map` with two or more arrays
  (the second array's values are the second argument); `call_user_func_array`
  whose argument array is a literal with two or more non-spread elements or
  with a `'more_entropy'` string key.

## Report

- Range: D1 — the whole call, from the name (including a leading `\` if
  written) to the closing `)`; D3 — the callback argument `A` as written
  (a string literal including its quotes, or the variable that was
  resolved).
- Severity: error.
- Message: `Pass more_entropy = true to uniqid() to reduce collisions.`

## Fix

- **F1** Direct call: insert the missing arguments and leave everything
  else between the parentheses untouched (comments, spacing, trailing
  comma). An empty list gets `'', true` right after `(`; otherwise `, true`
  is inserted right after the last argument, or `, more_entropy: true` when
  an argument is named. No fix when an argument is unpacked (`...$args`).
  - `uniqid()` → `uniqid('', true)`
  - `uniqid('ord_')` → `uniqid('ord_', true)`
  - `\uniqid($p)` → `\uniqid($p, true)`
  - `uniqid(prefix: 'q')` → `uniqid(prefix: 'q', more_entropy: true)`
  - `uniqid('a' /* why */ ,)` → `uniqid('a', true /* why */ ,)`
- **F2** Callback: replace `A` with the closure
  `function ($value = '') { return uniqid((string) $value, true); }`.
  The default keeps a call without arguments working, and the cast keeps
  non-string prefixes working under `declare(strict_types=1)`, where the
  call inside the closure is type-checked strictly while the engine's own
  call of the string callable was not.
  The fix is offered only when all of these hold; otherwise the report is
  made without a fix:
  - the function is `call_user_func`, `call_user_func_array` or `array_map`
    (callback index 0), where the callback's first parameter is what
    `uniqid()` would receive as its prefix. `array_filter`, `array_reduce`,
    `array_walk` and `array_walk_recursive` pass other arguments
    (`$carry, $item`, `$value, $key`, …), so a one-parameter closure would
    change the values `uniqid()` is called with;
  - `A` is itself a string literal. A variable that value discovery resolves
    to `'uniqid'` may hold another callable on a path discovery does not see
    (a parameter default, a fallback, a conditional reassignment), and the
    closure would silently replace it;
  - the call has exactly two arguments, both positional and neither
    unpacked: a named or unpacked argument could reach the closure under a
    name it does not declare;
  - for `call_user_func_array`, the second argument is an array literal with
    exactly one element that has no key, is not unpacked and is not taken by
    reference. Any other array (a variable, an empty or keyed literal) may
    forward zero, named or several values.

  Builtin spelling: `uniqid` is written `\uniqid` when an unqualified call at
  that position would not reach the global function (a `use function`
  import under that name, or a same-named function declared in the current
  namespace) inside the closure: in `namespace Ids; function uniqid() {…}`,
  `array_map('uniqid', $l)` →
  `array_map(function ($value = '') { return \uniqid((string) $value, true); }, $l)`
  (the string callable always names the global function).

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
    $tags = array_map(function ($value = '') { return uniqid((string) $value, true); }, ['a', 'b']);
    $one  = call_user_func_array(function ($value = '') { return uniqid((string) $value, true); }, ['x']);

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
- **F1 (custos diverges).** A named argument gets `more_entropy: true`
  appended and an unpacked argument gets no fix, because appending a
  positional `true` after either is invalid PHP. The fix is an insertion,
  so comments inside the argument list are kept instead of being dropped by
  a rebuilt list.
- **Forwarded `more_entropy` (custos diverges).** Upstream reports a string
  callback whenever the callback-taking call has two arguments or more.
  custos does not report when the call already forwards a second argument
  to `uniqid()` (E4), for the same reason E1 exempts a direct call.
- **F2 closure and its conditions (custos diverges).** Upstream replaces
  the callback with a closure that has one required, uncast parameter, for
  any resolved callback and any argument list. That closure throws where
  the string callable worked: no forwarded value, a forwarded named or
  unpacked value, a non-string value under `declare(strict_types=1)`; and
  replacing a variable discards whatever other callable it may hold. custos
  emits a closure with a defaulted, cast parameter and withholds the fix
  outside the conditions listed in F2.
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
