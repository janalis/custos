---
id: FopenBinaryUnsafeUsage
group: Compatibility
kind: semantic
needs: [names, index]
php: { min: "", max: "" }
---

# FopenBinaryUnsafeUsage

## Summary

The PHP manual recommends always opening files with the `b` (binary) flag
for portability; the `t` (text translation) flag is Windows-only and
mangles line endings. Also, `b` must come after the access letter (`rb`,
`wb+`, `r+b`) — a leading or middle `b` (`bw`) is wrong.

## Detection

Visit every plain function call.

- **D1** The call resolves to the global function `fopen`: the name compares
  case-insensitively (`FOpen`), and a call resolving to a same-named
  namespaced function (declared in the current namespace, imported with
  `use function`, or written with a non-global qualifier) does not count.
- **D2** At least two arguments. Let `A` be the second argument.
- **D3** Obtain the mode literal `L`:
  - if `A` is itself a string literal, `L` is `A`;
  - otherwise run *value discovery* (as defined in the
    `CallableMethodValidity` spec) on `A`; among the discovered values keep
    only string literals; if exactly one remains it is `L`;
  - otherwise stop.
- **D4** Let `m` be the raw content of `L` (between the quotes, escape
  sequences not decoded). Stop when `m` is empty.
- **D5** Classify, first match wins:
  - **K1 misplaced `b`**: `m` contains `b` and does **not** end with `b` and
    does **not** end with `b+`. Reported regardless of the option.
  - (`m` contains `b` and ends with `b` or `b+` → fine, no report, even if it
    also contains `t`, e.g. `wtb`.)
  - **K2 text mode**: `m` contains no `b` but contains `t`. Reported only
    when `ENFORCE_BINARY_MODIFIER_USAGE` is on.
  - **K3 missing `b`**: `m` contains neither `b` nor `t`. Reported only when
    `ENFORCE_BINARY_MODIFIER_USAGE` is on.

All character tests are case-sensitive (`B`, `T` are ordinary characters).

## Exceptions (no report)

- **E1** Fewer than two arguments.
- **E2** Mode not traceable to exactly one string literal (unknown variable,
  call result, two candidate literals, top-level variable).
- **E3** Empty mode `''`.
- **E4** Correct binary modes: `b`, `rb`, `wb+`, `r+b`, `xtb`.
- **E5** K2/K3 when the option is off.

## Report

- Range: the second argument `A` as written (the variable, the literal with
  its quotes, or whatever expression) — not the resolved literal.
- Severity: K1 → error; K2, K3 → warning.
- Messages:
  - K1: `Move the 'b' flag to the end of the mode (e.g. 'rb', 'rb+').`
  - K2: `Use the 'b' flag instead of 't' for binary-safe file access.`
  - K3: `Add the 'b' flag to the mode for binary-safe file access.`

## Fix

All three kinds share one fix, which rewrites the **resolved literal `L`**
(possibly located elsewhere, e.g. in an earlier `$mode = 'w';` assignment),
not the argument:

- **F1** Starting from `m`:
  1. remove every `b`;
  2. replace every `t` with `b`;
  3. if the result contains no `b`: if it contains `+`, replace every `+`
     with `b+`; otherwise append `b`.
- **F2** Replace `L` with a **single-quoted** literal of the result
  (`'…'`), whatever quoting `L` had.
  - `'w'` → `'wb'`; `'a+'` → `'ab+'`; `"r"` → `'rb'`; `'wt'` → `'wb'`;
    `'wt+'` → `'wb+'`; `'bw+'` → `'wb+'`; `'bx'` → `'xb'`; `'rbt'` → `'rb'`.
- If `m` is empty the fix does nothing (cannot happen given D4).

## Options

| Option | Type | Default | Effect |
|---|---|---|---|
| `ENFORCE_BINARY_MODIFIER_USAGE` | bool | `true` | When on, modes lacking `b` (with or without `t`) are reported; when off only misplaced `b` is reported. |

## PHP versions

None.

## Examples

```php
<?php
function open_files($path)
{
    $access = 'a';
    $h1 = fopen($path, <warning descr="Add the 'b' flag to the mode for binary-safe file access.">$access</warning>);
    $h2 = fopen($path, <warning descr="Add the 'b' flag to the mode for binary-safe file access.">"r+"</warning>);
    $h3 = fopen($path, <warning descr="Use the 'b' flag instead of 't' for binary-safe file access.">'at'</warning>);
    $h4 = \fopen($path, <error descr="Move the 'b' flag to the end of the mode (e.g. 'rb', 'rb+').">'bx'</error>);
    $h5 = fopen($path, 'r+b');
    $h6 = fopen($path, 'xtb');
    $h7 = fopen($path, '');
    $h8 = fopen($path);
    $h9 = fopen($path, $_GET['m']);
}
```

```php
<?php
function open_files($path)
{
    $access = 'ab';
    $h1 = fopen($path, $access);
    $h2 = fopen($path, 'rb+');
    $h3 = fopen($path, 'ab');
    $h4 = \fopen($path, 'xb');
    $h5 = fopen($path, 'r+b');
    $h6 = fopen($path, 'xtb');
    $h7 = fopen($path, '');
    $h8 = fopen($path);
    $h9 = fopen($path, $_GET['m']);
}
```

## Divergences

- **`fopen` matching (custos diverges from upstream).** Upstream matches
  the written last segment case-sensitively without resolution, so
  `FOPEN($p, 'r')` is missed and a namespace's own `fopen()` is checked (and
  "fixed"). custos resolves the call to the global function, in any case.
- Interpolated double-quoted modes (`"{$m}"`) are processed on their raw
  text upstream. Recommendation: skip literals containing interpolation.
- When several `fopen()` calls share one resolved literal, upstream applies
  the rewrite once (later fixes find the literal already replaced).
  Recommendation: deduplicate edits on the same literal.
