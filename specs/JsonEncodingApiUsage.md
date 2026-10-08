---
id: JsonEncodingApiUsage
group: Type compatibility
kind: semantic
needs: [names, index, stubs]
php: { min: "", max: "" }
---

# JsonEncodingApiUsage

## Summary

`json_decode()` without its second argument silently returns `stdClass`
objects, which readers often confuse with arrays; and both `json_decode()` and
`json_encode()` report failures only through `json_last_error()` unless the
`JSON_THROW_ON_ERROR` flag (PHP 7.3+) is passed. This rule asks for an explicit
decoding target and for exception-based error handling.

## Detection

Visit every function call whose name (last segment, compared
case-insensitively as PHP does) is `json_decode` or `json_encode`.

- **D0** The call must resolve to the global built-in function
  (`\json_decode` / `\json_encode`): a fully-qualified call, an unqualified
  call in the global namespace, or an unqualified call in a namespace where no
  function of that name is declared in the namespace (fallback to global).
  A call resolving to a user function `Some\Ns\json_decode` is ignored.
- Argument lookup below ("argument *name* at position *i*") means: the
  argument passed by name `name:` if present, otherwise the positional
  (unnamed) argument at zero-based index `i`. PHP parameter names are
  `json_decode(json, associative, depth, flags)` and
  `json_encode(value, flags, depth)`.

### Decoding target (`json_decode` only)

- **D1** Option `HARDEN_DECODING_RESULT_TYPE` is on, the call has at least one
  argument, and there is no argument `associative` at position 1 (neither
  positional second argument nor `associative:` named argument). Report
  kind T.

### Error handling (`json_decode` and `json_encode`)

- **D2** Option `HARDEN_ERRORS_HANDLING` is on, the project PHP level is
  **≥ 7.3**, and the call has at least one argument.
- **D3** The subject argument exists: `json` at position 0 for
  `json_decode`, `value` at position 0 for `json_encode`.
- **D4** Let `F` be the flags argument: `flags` at position 3 for
  `json_decode`, `flags` at position 1 for `json_encode`. Report kind E unless
  `F` exists and *carries a strict flag* (D5).
- **D5a** If value discovery on `F` (D5) gives an *unknown* result (a
  variable on the path is also incremented/decremented or compound-assigned
  in its scope, e.g. `$flags = 0; $flags |= JSON_PRETTY_PRINT;` — custos
  refinement, see Divergences), kind E is **not** reported for that call.
- **D5** `F` carries a strict flag when *value discovery* (as defined in the
  `CallableMethodValidity` spec: parentheses stripped, ternary / `??`
  branches, function-local assignments and parameter defaults, property
  defaults, class constants, global constants resolved to their `define()`
  value including stub constants; top-level variables yield nothing) on `F`
  yields **exactly one** value `V`, and:
  - `V` is a number literal (or `-` + number literal) whose source text is
    exactly `4194304` (`JSON_THROW_ON_ERROR`) or `512`
    (`JSON_PARTIAL_OUTPUT_ON_ERROR`); or
  - `V` is a bare constant reference named exactly (case-sensitive)
    `JSON_THROW_ON_ERROR` or `JSON_PARTIAL_OUTPUT_ON_ERROR`; or
  - `V` is any other expression (e.g. `A | B`, a call) that contains, at any
    depth, a constant reference with one of those two names.
  Zero values (unknown parameter, unresolvable constant, top-level variable)
  or several values (ternary with different branches) → not strict → report.

  Consequences: `JSON_THROW_ON_ERROR` written directly resolves through the
  stubs to `4194304` → strict; `$opts | JSON_THROW_ON_ERROR` → strict;
  a local `$f = JSON_PARTIAL_OUTPUT_ON_ERROR;` then `$f` → strict;
  `define('MY_FLAGS', JSON_THROW_ON_ERROR | JSON_HEX_TAG)` then `MY_FLAGS` →
  strict; a class constant with such a value → strict; `JSON_PRETTY_PRINT`
  alone → not strict.

Both kinds may be reported on the same call (T first, then E), each with its
own fix.

## Exceptions (no report)

- **E1** Calls with no arguments.
- **E2** Calls resolving to a non-global function of the same name.
- **E3** `json_decode` with an explicit second / `associative:` argument (any
  value) — no kind T.
- **E4** Kind E is never reported below PHP 7.3, or when the flags argument
  carries a strict flag (D5).
- **E5** Kind E: if `json`/`value` cannot be found (e.g. only named arguments
  with other names), nothing is reported.

## Report

- Range: the whole call expression, from the function name (including any
  leading `\`) to the closing `)`.
- Severity: info (weak warning), both kinds.
- Messages:
  - T: `Pass the second argument to state whether JSON decodes to arrays or objects.`
  - E: `Pass JSON_THROW_ON_ERROR in the flags of this call.`

## Fix

The fix replaces the whole call with newly built text. `NS` is the namespace
qualifier exactly as written before the function name (`\` for
`\json_decode(...)`, empty for an unqualified call). Argument texts are the
argument expressions' source text (without any `name:` label). Separators are
exactly `, `. `C` is the spelling of `JSON_THROW_ON_ERROR` chosen by F6. F1b,
F4 and F5 edit the call in place instead of replacing it.

- **F1** (kind T) `NS json_decode(<arg0>, <bool>)` where `<arg0>` is the first
  argument's text and `<bool>` is `true` when `DECODE_AS_ARRAY` is on, else
  `false`. All other arguments are dropped (see Divergences).
  `json_decode($raw)` → `json_decode($raw, true)` (with DECODE_AS_ARRAY on).
- **F1b** (kind T, custos) When the call passes an argument by name, the
  fix instead appends `, associative: <bool>` after the last argument and
  keeps every argument: `json_decode($raw, depth: 8)` →
  `json_decode($raw, depth: 8, associative: true)`.
- **F2** (kind E, `json_decode`) only when **no** argument is passed by
  name; result
  `NS json_decode(<json>, <assoc>, <depth>, <flags>)`:
  - `<json>`: text of argument `json`/position 0;
  - `<assoc>`: text of argument `associative`/position 1 if present, else
    `true` when `HARDEN_DECODING_RESULT_TYPE` and `DECODE_AS_ARRAY` are both
    on, else `false`;
  - `<depth>`: text of argument `depth`/position 2 if present, else `512`;
  - `<flags>`: `C` if there is no flags argument, else `C | ` + flags
    text, the flags text in parentheses when it binds more loosely than
    `|` (logical operators, `??`, ternary, assignment, `yield`, `print`,
    `include`, `throw`, arrow function).
  `json_decode($raw, true, 64, $mode)` →
  `json_decode($raw, true, 64, JSON_THROW_ON_ERROR | $mode)` (with a bare
  `C`).
- **F3** (kind E, `json_encode`) only when no argument is passed by name;
  result `NS json_encode(<value>, <flags>)` when there is no
  `depth` argument (position 2), else `NS json_encode(<value>, <flags>, <depth>)`,
  with `<flags>` built as in F2.
  `json_encode($row, $mode, 8)` → `json_encode($row, JSON_THROW_ON_ERROR | $mode, 8)`.
- **F4** (custos) With a named `flags:` argument, `C | ` is prepended to
  its value (parenthesised as in F2): `json_encode($v, flags: $m)` →
  `json_encode($v, flags: C | $m)`.
- **F5** (custos) With another argument passed by name, `, flags: C` is
  appended after the last argument and every argument is kept:
  `\json_decode($x, associative: true)` →
  `\json_decode($x, associative: true, flags: \JSON_THROW_ON_ERROR)`.
- **F6** `C` is `\JSON_THROW_ON_ERROR` when the file already writes a global
  constant with a leading backslash (`\PHP_EOL`; `\true`, `\false` and
  `\null` do not count), or when a bare name at the call would not reach
  the global constant (a `use const` import or a constant of that name in
  the current namespace); otherwise `JSON_THROW_ON_ERROR`.

## Options

| Option | Type | Default | Effect |
|---|---|---|---|
| HARDEN_DECODING_RESULT_TYPE | bool | true | Enables kind T (D1). Also affects `<assoc>` in F2. |
| DECODE_AS_ARRAY | bool | false | Fixes insert `true` (decode to arrays) instead of `false`. |
| DECODE_AS_OBJECT | bool | true | Radio-button companion of DECODE_AS_ARRAY; has no effect on detection or fixes by itself. |
| HARDEN_ERRORS_HANDLING | bool | true | Enables kind E (D2–D5). |

Upstream fixture configurations: (a) T only — HARDEN_DECODING_RESULT_TYPE on,
DECODE_AS_ARRAY on, HARDEN_ERRORS_HANDLING off, default PHP level;
(b) E only — HARDEN_DECODING_RESULT_TYPE off, HARDEN_ERRORS_HANDLING on, PHP 7.3.

## PHP versions

- Kind T: no gating.
- Kind E: project level ≥ 7.3 (`JSON_THROW_ON_ERROR` introduced in 7.3). Test
  cases without an explicit level run below 7.3, so kind E never appears
  there.
- Named arguments (PHP 8.0 syntax) are parsed at any level and honoured in
  argument lookup.

## Examples

Configuration (a): T only, decode as array.

```php
<?php
namespace Feed;

function load(string $body, string $alt) {
    $a = <weak_warning descr="Pass the second argument to state whether JSON decodes to arrays or objects.">json_decode($body)</weak_warning>;
    $b = <weak_warning descr="Pass the second argument to state whether JSON decodes to arrays or objects.">\json_decode(trim($alt))</weak_warning>;
    $c = json_decode($body, false);
    $d = json_decode($body, associative: true);
    $e = json_decode();
    return [$a, $b, $c, $d, $e];
}
```

```php
<?php
namespace Feed;

function load(string $body, string $alt) {
    $a = json_decode($body, true);
    $b = \json_decode(trim($alt), true);
    $c = json_decode($body, false);
    $d = json_decode($body, associative: true);
    $e = json_decode();
    return [$a, $b, $c, $d, $e];
}
```

Configuration (b): E only, PHP 7.3.

```php
<?php
define('EXPORT_FLAGS', JSON_PARTIAL_OUTPUT_ON_ERROR | JSON_UNESCAPED_SLASHES);

class Exporter {
    const MODE = JSON_THROW_ON_ERROR | JSON_PRETTY_PRINT;

    function run(array $rows, int $mode, int $max) {
        $strict = JSON_THROW_ON_ERROR;
        return [
            <weak_warning descr="Pass JSON_THROW_ON_ERROR in the flags of this call.">json_decode($rows[0])</weak_warning>,
            <weak_warning descr="Pass JSON_THROW_ON_ERROR in the flags of this call.">json_decode($rows[1], true, $max)</weak_warning>,
            <weak_warning descr="Pass JSON_THROW_ON_ERROR in the flags of this call.">json_decode($rows[2], null, 16, JSON_BIGINT_AS_STRING)</weak_warning>,
            <weak_warning descr="Pass JSON_THROW_ON_ERROR in the flags of this call.">json_decode($rows[3], flags: $mode)</weak_warning>,
            <weak_warning descr="Pass JSON_THROW_ON_ERROR in the flags of this call.">json_encode($rows)</weak_warning>,
            <weak_warning descr="Pass JSON_THROW_ON_ERROR in the flags of this call.">json_encode($rows, $mode | JSON_HEX_TAG)</weak_warning>,
            <weak_warning descr="Pass JSON_THROW_ON_ERROR in the flags of this call.">json_encode($rows, $mode, $max)</weak_warning>,
            <weak_warning descr="Pass JSON_THROW_ON_ERROR in the flags of this call.">json_encode($rows, flags: $mode)</weak_warning>,

            json_decode($rows[0], true, 32, JSON_THROW_ON_ERROR),
            json_decode($rows[0], true, 32, $mode | JSON_PARTIAL_OUTPUT_ON_ERROR),
            json_decode($rows[0], flags: $strict),
            json_encode($rows, EXPORT_FLAGS),
            json_encode($rows, self::MODE),
            json_encode($rows, 4194304),
            json_encode(),
        ];
    }
}
```

```php
<?php
define('EXPORT_FLAGS', JSON_PARTIAL_OUTPUT_ON_ERROR | JSON_UNESCAPED_SLASHES);

class Exporter {
    const MODE = JSON_THROW_ON_ERROR | JSON_PRETTY_PRINT;

    function run(array $rows, int $mode, int $max) {
        $strict = JSON_THROW_ON_ERROR;
        return [
            json_decode($rows[0], false, 512, JSON_THROW_ON_ERROR),
            json_decode($rows[1], true, $max, JSON_THROW_ON_ERROR),
            json_decode($rows[2], null, 16, JSON_THROW_ON_ERROR | JSON_BIGINT_AS_STRING),
            json_decode($rows[3], flags: $mode),
            json_encode($rows, JSON_THROW_ON_ERROR),
            json_encode($rows, JSON_THROW_ON_ERROR | $mode | JSON_HEX_TAG),
            json_encode($rows, JSON_THROW_ON_ERROR | $mode, $max),
            json_encode($rows, flags: $mode),

            json_decode($rows[0], true, 32, JSON_THROW_ON_ERROR),
            json_decode($rows[0], true, 32, $mode | JSON_PARTIAL_OUTPUT_ON_ERROR),
            json_decode($rows[0], flags: $strict),
            json_encode($rows, EXPORT_FLAGS),
            json_encode($rows, self::MODE),
            json_encode($rows, 4194304),
            json_encode(),
        ];
    }
}
```

## Divergences

- **Case of the name (custos diverges from upstream).** Upstream compares
  the written name case-sensitively, so `JSON_DECODE($s)` is not reported.
  custos matches any case; the fix keeps the name and qualifier as written.
- **Unstable variables — custos refinement, not upstream.** Value discovery ignores `++`/`--` and compound assignments upstream, so a variable later incremented or extended is analysed with its initial value only. custos makes the result unknown (kind E is not reported, since the flags may well include a strict flag added through `|=`), as in the shared value discovery of `CallableMethodValidity`. No upstream fixture relies on such a variable; recorded in `docs/internals/decisions.md` ("Spec-level false positives").
- F1 keeps only the first argument; with named trailing arguments
  (`json_decode($s, depth: 4)`, `json_decode($s, flags: X)`) upstream drops
  them. Recommendation: insert the boolean as the second positional argument
  and keep the remaining arguments when they are named; produce no fix when
  positional and named arguments make insertion ambiguous. No fixture covers
  it.
- **Low-precedence flags are parenthesised (custos diverges, F2–F4).**
  Prepending `JSON_THROW_ON_ERROR | ` to `$pretty ? JSON_PRETTY_PRINT : 0`
  gives `(JSON_THROW_ON_ERROR | $pretty) ? JSON_PRETTY_PRINT : 0`, which
  drops the strict flag and changes the flags; custos writes
  `JSON_THROW_ON_ERROR | ($pretty ? JSON_PRETTY_PRINT : 0)`.
- **Named arguments are kept (custos diverges, F1b, F4, F5).** Upstream
  rebuilds the call positionally, turning `json_decode($x, associative: true)`
  into `json_decode($x, true, 512, JSON_THROW_ON_ERROR)`, and its result-type
  fix drops named trailing arguments (`json_decode($s, flags: X)` lost its
  flags). When the call already uses named arguments, custos appends the
  missing one by name (or extends the named `flags:` value) and leaves the
  others untouched.
- **Constant spelling follows the file (custos diverges, F6).** Upstream
  always inserts the bare `JSON_THROW_ON_ERROR`, out of place in a file that
  writes `\JSON_PRETTY_PRINT` or `\PHP_EOL`, and wrong where a namespaced
  constant of that name would capture the bare name.
- The numeric check treats a literal `512` as a strict flag; the value equals
  `JSON_PARTIAL_OUTPUT_ON_ERROR`, so it is correct for the flags slot. Keep.
- With both kinds enabled on the same `json_decode` call, applying both fixes
  in sequence is order-dependent; upstream fixtures never enable both.
  Recommendation: when both apply, offer them independently; conformance only
  checks configurations (a) and (b).
- Value discovery treats an unresolvable constant as "no value" → reported.
  An implementation without full stubs should at least resolve
  `JSON_THROW_ON_ERROR` (4194304) and `JSON_PARTIAL_OUTPUT_ON_ERROR` (512), or
  equivalently accept those bare names directly.
