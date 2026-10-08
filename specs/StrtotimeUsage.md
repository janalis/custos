---
id: StrtotimeUsage
group: Performance
kind: semantic
needs: [names, stubs]
php: { min: "", max: "" }
---

# StrtotimeUsage

## Summary

Two wasteful `strtotime()` idioms: parsing the literal `'now'` just to get the
current timestamp (`time()` does that directly), and passing `time()` as the
base timestamp, which is already the default.

## Detection

- **D1** A plain function call whose name, as written in its last segment, is
  `strtotime` (compared case-insensitively, as PHP does), **and** that call resolves to the
  global function `\strtotime` with the same rules as the `time()` call in
  D3 (a namespaced or imported user function `strtotime` → no report). Call
  it `C`. It must have 1 or 2 arguments (0 or
  more than 2 → nothing).
- **D2** *Literal now* — `C` has exactly 1 argument, which is a string literal
  (single- or double-quoted) whose raw contents between the quotes equal
  `now` ignoring letter case (`'now'`, `"NOW"`, `'Now'`). No trimming: `' now'`
  or `'now '` do not match; an interpolated string never matches.
- **D3** *Redundant base* — `C` has exactly 2 arguments, and the 2nd argument
  is a plain function call (not method/static, not wrapped in parentheses)
  whose name as written (last segment) is `time` (case-insensitive),
  **and** that call resolves to the global function `\time`:
  - `\time()` → global;
  - unqualified `time()` in the global namespace → global;
  - unqualified `time()` inside a namespace → global only when no function
    `time` exists in that namespace (and no `use function` import brings in a
    non-global `time`); if it resolves to a namespaced function → no report;
  - unresolvable → no report (`time` is a known built-in, so this only
    happens if it is shadowed).
  The arguments of the `time()` call are not inspected. The 1st argument of
  `C` can be anything.

## Exceptions (no report)

- **E1** `strtotime()` with 0 or ≥ 3 arguments.
- **E2** Single argument that is not exactly `now` (case-insensitive), e.g.
  `'today'`, `'+1 day'`, `'now '`, a variable, a constant.
- **E3** 2nd argument other than a call to the global `time()`: a variable,
  `microtime()`, `$clock->time()`, `Ns\time()` resolving to a namespaced
  function, `(time())`.
- **E4** Name with different case (`StrToTime`), method calls.

## Report

- Range: the whole call `C` (from the start of its name, including any
  namespace qualifier, to its closing `)`).
- Severity: warning for both cases (the D3 case is shown upstream with the
  "unused symbol" look, but its severity is still warning).
- Messages:
  - D2: `Call time() instead of parsing 'now'.`
  - D3: `The base timestamp already defaults to the current time; drop the time() argument.`

## Fix

- **F1** (D2) Replace `C` with a call to the global `time()`:
  - `\time()` when `C`'s name is fully qualified (`\strtotime('now')` →
    `\time()`);
  - `\time()` when an unqualified `time` would not reach the global function
    at `C`'s position (a function `time` declared in the current namespace,
    or a `use function` import of a non-global `time`);
  - `time()` otherwise (`strtotime('now')` → `time()`).
- **F2** (D3) Replace `C` with `{name}({arg1})`, where `{name}` is `C`'s
  function name exactly as written (qualifier included) and `{arg1}` is the
  verbatim text of `C`'s 1st argument (`\strtotime($s, time())` →
  `\strtotime($s)`, `strtotime($s, time())` → `strtotime($s)`).
- No parentheses are added around the replacement.

## Options

None.

## PHP versions

No gating. Upstream fixture runs at the test default level.

## Examples

```php
<?php
namespace Billing;

function due($spec) {
    $now   = <warning descr="Call time() instead of parsing 'now'.">strtotime("Now")</warning>;
    $other = <warning descr="Call time() instead of parsing 'now'.">\strtotime('NOW')</warning>;
    $next  = <warning descr="The base timestamp already defaults to the current time; drop the time() argument.">strtotime($spec, \time())</warning>;
    $last  = <warning descr="The base timestamp already defaults to the current time; drop the time() argument.">strtotime('last monday', time())</warning>;

    $a = strtotime('now ');
    $b = strtotime($spec, $now);
    $c = strtotime($spec, microtime(true));
    $d = strtotime('now', time());
    $e = strtotime();
    return [$now, $other, $next, $last, $a, $b, $c, $d, $e];
}
```

```php
<?php
namespace Billing;

function due($spec) {
    $now   = time();
    $other = \time();
    $next  = strtotime($spec);
    $last  = strtotime('last monday');

    $a = strtotime('now ');
    $b = strtotime($spec, $now);
    $c = strtotime($spec, microtime(true));
    $d = strtotime('now');
    $e = strtotime();
    return [$now, $other, $next, $last, $a, $b, $c, $d, $e];
}
```

Note the `$d` line: only D3 fires on `strtotime('now', time())` (D2 requires
exactly one argument); after the fix the result is `strtotime('now')`, which
D2 reports on a subsequent run.

A namespaced shadow is not reported:

```php
<?php
namespace Clock;
function time() { return 0; }
$t = strtotime('+1 hour', time());
```

## Divergences

- **Qualifier kept by the fix — custos diverges from upstream** (F1, F2).
  Upstream always emits a bare `time()` / `strtotime(...)`, dropping a
  leading `\`. Inside a namespace that declares (or imports) its own `time`
  or `strtotime`, the bare name then calls that function instead of the
  built-in. custos keeps the qualifier as written for F2 and emits `\time()`
  for F1 whenever the original was fully qualified or a bare `time` would not
  resolve to the global function.
- **`strtotime` resolved — custos diverges from upstream** (D1). Upstream
  matches the callee on its name only, so a user function `App\strtotime()`
  (called qualified, imported, or unqualified inside `namespace App`) is
  reported and "fixed" as if it were the built-in. custos reports only calls
  that reach the global `strtotime`.
- **Function-name case — custos diverges from upstream.** Upstream matches
  `strtotime`/`time` case-sensitively, so `StrToTime('now')` is missed
  although PHP calls the same function. custos compares them
  case-insensitively (the resolution requirement is unchanged).
