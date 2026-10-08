---
id: CascadingDirnameCalls
group: Language level migration
kind: syntax
needs: []
php: { min: "7.0", max: "" }
---

# CascadingDirnameCalls

## Summary

Since PHP 7.0 `dirname()` accepts a second "levels" argument, so a chain of
nested `dirname()` calls can be collapsed into a single call with the summed
level count: fewer calls, easier to read.

## Detection

Terminology: a *dirname call* is a function call (not a method/static call)
whose unqualified name is `dirname`, compared case-insensitively as PHP
resolves function names (custos diverges, see Divergences; any namespace
qualifier such as `\dirname(...)` is ignored when comparing the name) and
which resolves to the global `dirname()` (not a `use function` import from
another namespace, nor an unqualified call in a namespace declaring its own
`dirname`; a qualified `Foo\dirname(...)` never matches). An
*eligible* dirname call has exactly 1 or exactly 2 arguments.

- **D1** The visited node is an eligible dirname call `C` (the *top* call).
- **D2** Top-only: `C` is skipped when it is itself, directly (no parentheses
  in between), the **first** argument of another eligible dirname call (it
  is then a member of that call's chain). Only the outermost call of a chain
  is analysed. A dirname chain used as the *levels* argument
  (`dirname($p, dirname(dirname($q)))`) is not a chain member and is
  analysed on its own.
- **D3** Walk the chain starting at `C`. For the current call `K`:
  - if `K` is not named `dirname`, stop (without consuming `K`);
  - if `K` has 1 argument: record `arg = K.arg0`, add 1 to the *count*;
  - if `K` has 2 arguments: record `arg = K.arg0`, append `K.arg1` to the
    *level list* (in outer-to-inner order);
  - any other argument count: stop (without consuming `K`);
  - if `K.arg0` is a function call, continue with it as the next `K`;
    otherwise stop.
  `arg` ends up being the first argument of the innermost consumed dirname
  call (the *path*).
- **D4** Report only if at least one dirname call nested in `C` was consumed,
  i.e. the path `arg` is not `C`'s own first argument. Thus
  `dirname(dirname())` (inner call has 0 args → not consumed) and
  `dirname(realpath($p), 2)` (first argument is not a dirname call) are not
  reported.
- **D5** Level computation: for every entry of the level list, in order, try
  to parse its source text as a signed 32-bit decimal integer (optional
  leading `+`/`-`, decimal digits only; leading zeros allowed and read as
  decimal). Parsed values are added to the count; entries that do not parse
  (variables, expressions, hex/octal/binary literals, literals with `_`,
  out-of-range numbers) are kept, in order, as an *expression list* (their
  source text).
- **D6** If the resulting count is exactly 1 and the expression list is empty,
  do not report (nothing to collapse, e.g. `dirname(dirname($p, 0))`).
- **D7** Replacement text `R`:
  `dirname(` + path source text + `, ` + levels + `)`, where levels is the
  count (decimal, may be `0` or negative) followed by each expression-list
  entry, all joined with ` + `. The count is always emitted first, even when
  it is 0 (`0 + $a + $b`). The function name is always plain `dirname`, even
  if the original was written `\dirname`.

## Exceptions (no report)

- **E1** PHP language level below 7.0.
- **E2** A single dirname call (no nested dirname call consumed), with any
  argument: `dirname($p)`, `dirname($p, 1)`, `dirname($p, 4)`,
  `dirname(trim($p))`, `dirname(realpath($p), 2)`.
- **E3** Calls with 0 or 3+ arguments are never top calls and break a chain.
- **E4** Inner dirname calls of a reported chain are not reported themselves
  (D2).
- **E5** Count of exactly 1 with no expression levels (D6).

## Report

- Range: the whole top call `C`, from the start of its name (including any
  leading `\` / namespace qualifier) to its closing `)`.
- Severity: warning.
- Message: `Collapse the nested dirname() calls into '{R}'.`

## Fix

- **F1** Replace the top call `C` with `R` (D7) verbatim. Path and level
  expressions keep their original source text; separators are exactly `, `
  and ` + `.
  - `dirname(dirname(dirname($root)))` → `dirname($root, 3)`
  - `dirname(dirname(dirname(dirname(getcwd()))))` → `dirname(getcwd(), 4)`
  - `dirname(dirname($root, $up))` → `dirname($root, 1 + $up)`
  - `dirname(dirname($root, 2), 3)` → `dirname($root, 5)`
  - `dirname(dirname($root, $a), $b)` → `dirname($root, 0 + $b + $a)`

## Options

None.

## PHP versions

- Reported only when the configured PHP level is ≥ 7.0 (the version that
  introduced the `levels` parameter). The upstream fixture runs at 7.1.

## Examples

```php
<?php
$base = <warning descr="Collapse the nested dirname() calls into 'dirname($root, 3)'.">dirname(dirname(dirname($root)))</warning>;
$app  = <warning descr="Collapse the nested dirname() calls into 'dirname(getcwd(), 4)'.">dirname(dirname(dirname(dirname(getcwd()))))</warning>;
$var  = <warning descr="Collapse the nested dirname() calls into 'dirname($root, 1 + $up)'.">dirname(dirname($root, $up))</warning>;
$sum  = <warning descr="Collapse the nested dirname() calls into 'dirname($root, 5)'.">dirname(dirname($root, 2), 3)</warning>;
$two  = <warning descr="Collapse the nested dirname() calls into 'dirname($root, 0 + $b + $a)'.">dirname(dirname($root, $a), $b)</warning>;
$fq   = <warning descr="Collapse the nested dirname() calls into 'dirname($root, 2)'.">\dirname(\dirname($root))</warning>;

// not reported
$one  = dirname($root);
$lvl  = dirname($root, 1);
$far  = dirname($root, 4);
$wrap = dirname(getcwd());
$real = dirname(realpath($root), 2);
$zero = dirname(dirname($root, 0));
$bad  = dirname(dirname());
$none = dirname();
```

```php
<?php
$base = dirname($root, 3);
$app  = dirname(getcwd(), 4);
$var  = dirname($root, 1 + $up);
$sum  = dirname($root, 5);
$two  = dirname($root, 0 + $b + $a);
$fq   = dirname($root, 2);

// not reported
$one  = dirname($root);
$lvl  = dirname($root, 1);
$far  = dirname($root, 4);
$wrap = dirname(getcwd());
$real = dirname(realpath($root), 2);
$zero = dirname(dirname($root, 0));
$bad  = dirname(dirname());
$none = dirname();
```

## Divergences

- **Function-name case (custos diverges):** upstream matches `dirname`
  case-sensitively, so `Dirname(dirname($p))` or `DIRNAME(DIRNAME($p))` is
  neither reported nor walked as a chain, although PHP calls the same
  function. custos compares the name case-insensitively; the replacement
  still writes plain lower-case `dirname` (D7).
- Upstream's chain walk (D3) also continues into *method* calls named
  `dirname` (`dirname(dirname($fs->dirname($p)))` would treat the method call
  as part of the chain and drop the object). Recommendation: only plain
  function calls extend a chain; a method/static call stops the walk like any
  other non-dirname call. Not covered by fixtures.
- Level expressions are joined with ` + ` without parentheses, so a
  low-precedence level (`dirname(dirname($p, $deep ? 2 : 1))`) yields
  `1 + $deep ? 2 : 1` (wrong precedence). Recommendation: wrap any level
  expression that is not a primary expression (variable, constant, call,
  property/array access, literal) in parentheses. Not covered by fixtures.
- Integer literals with leading zeros are summed as decimal (`010` → 10) while
  PHP reads them as octal (8). Recommendation: treat a literal with a leading
  `0` followed by more digits as a non-parsable expression (kept as text).
  Not covered by fixtures.
- **Levels-argument chains (custos diverges):** upstream also skips a
  dirname call that is the *second* argument of an outer dirname call, so
  `dirname($p, dirname(dirname($q)))` misses the collapsible inner chain.
  That call is not part of the outer chain; custos skips only first-argument
  chain members (D2) and reports the inner chain.
- Named arguments (PHP 8, `dirname(path: $p, levels: 2)`) are not handled
  specially upstream. Recommendation: skip any call in the chain that uses
  named arguments or argument unpacking (`...`).
- **Shadowing namespaced function (custos diverges):** upstream also chains
  calls to a user `dirname()` declared in (or imported into) the current
  namespace. custos only chains calls that reach the global function, and
  writes the replacement as `\dirname(...)` when a bare `dirname` would
  resolve to such a user function at that position (D7).
