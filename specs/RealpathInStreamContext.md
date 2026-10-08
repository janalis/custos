---
id: RealpathInStreamContext
group: Probable bugs
kind: syntax
needs: []
php: { min: "", max: "" }
---

# RealpathInStreamContext

## Summary

`realpath()` returns `false` for paths inside stream wrappers such as
`phar://`, so code that works on disk breaks once packaged. Climbing
directories with `realpath(__DIR__ . '/../')` or wrapping include paths in
`realpath()` should use `dirname()` / the plain path instead.

## Detection

- **D1** A plain function call that resolves to the global function
  `realpath` (name compared case-insensitively, as PHP does: `RealPath(...)`
  matches; `\realpath(...)` and an unqualified call in the global namespace
  match; an unqualified call in a namespace matches only when no `realpath`
  function is declared in that namespace or imported with `use function`;
  `Foo\realpath(...)` does not; custos diverges), with exactly **one**
  argument `S`.
- **D2** Not a test context: the file path does not end with `Test.php`,
  `Spec.php` or `.phpt` and does not contain `/Fixtures/`, and the
  enclosing class FQN (if any) does not end with `Test` nor contain `\Tests\`
  or `\Test\`.
- **D3** (include context) After skipping any parentheses around the call,
  its parent is an inclusion expression (`include`, `include_once`,
  `require`, `require_once`) → report.
- **D4** (parent-directory climbing) Otherwise, report when any string literal
  anywhere inside the call (any depth, including inside nested calls or
  concatenations) has raw contents containing `..`.
  At most one report per call.

### Replacement text `R` (computed from `S`)

- **R1** `S` is a concatenation `L . Q` whose left operand `L` is **not**
  itself a concatenation and whose right operand `Q` is a string literal
  whose raw contents start with `/..`: let `rest` = the contents; while
  `rest` starts with `/..`, remove that leading `/..` and wrap `L` once more
  as `dirname(…)`. `R` = `<wrapped L> . <quote><rest><quote>`, using the
  literal's original quote character (`'` or `"`), with exactly one space on
  each side of the `.`.
  - `__DIR__ . '/../'` → `dirname(__DIR__) . '/'`
  - `$base . "/../../cfg"` → `dirname(dirname($base)) . "/cfg"`
  - `dirname` is written `\dirname` when an unqualified call at that
    position would not reach the global function: a `use function` import
    under that name, or a `dirname` function declared in the current
    namespace (`namespace App; function dirname($p) {…}` →
    `\dirname(__DIR__) . '/'`).
- **R2** `S` is a single- or double-quoted string literal whose contents are
  an absolute path → `R` = `S`'s full source text (quotes included). Absolute
  means: starts with `/` or `\`, starts with a drive letter followed by `:\`
  or `:/` (`C:\x`, `c:/x`), or contains `://` (stream or URL path).
- **R3** Anything else → no `R`: relative literals (`'../x'`, `'cfg.php'`,
  `'./a'`), heredoc/nowdoc literals, `__DIR__ . '/..' . '/..'` (a left-nested
  concatenation), variables, calls.

## Exceptions (no report)

- **E1** Zero or several arguments.
- **E2** Calls outside include context without any `..` in their literals:
  `realpath(__DIR__ . '/')`, `realpath($path)`.
- **E3** Test contexts.

## Report

- Range: the whole call, from `realpath` through its closing `)`.
- Severity: warning.
- Message:
  - with `R`: `Use '{R}' instead: realpath() fails inside stream wrappers.`
  - without `R`: `realpath() fails inside stream wrappers such as phar://; prefer dirname().`

## Fix

- **F1** Only when `R` exists: replace the whole call with `R`, verbatim.
  Surrounding code (an `include` keyword, parentheses around the call, the
  statement's `;`) stays untouched: `include (realpath('/a.php'));` →
  `include ('/a.php');`.
- No fix when `R` does not exist (the report remains).

## Options

None.

## PHP versions

None.

## Examples

```php
<?php
$root = <warning descr="Use 'dirname(dirname(__DIR__, 1)) . '/var'' instead: realpath() fails inside stream wrappers.">realpath(dirname(__DIR__, 1) . '/../var')</warning>;
$cfg  = <warning descr="Use 'dirname(dirname($app)) . &quot;/etc&quot;' instead: realpath() fails inside stream wrappers.">realpath($app . "/../../etc")</warning>;
$odd  = <warning descr="realpath() fails inside stream wrappers such as phar://; prefer dirname().">realpath($app . '/lib' . '/..')</warning>;
$rel  = <warning descr="realpath() fails inside stream wrappers such as phar://; prefer dirname().">realpath('../shared')</warning>;
require <warning descr="Use ''/opt/app/boot.php'' instead: realpath() fails inside stream wrappers.">realpath('/opt/app/boot.php')</warning>;
include_once (<warning descr="realpath() fails inside stream wrappers such as phar://; prefer dirname().">realpath($entry)</warning>);
$ok   = realpath($app . '/cache');
$two  = realpath('/tmp', 'x');
```

```php
<?php
$root = dirname(dirname(__DIR__, 1)) . '/var';
$cfg  = dirname(dirname($app)) . "/etc";
$odd  = realpath($app . '/lib' . '/..');
$rel  = realpath('../shared');
require '/opt/app/boot.php';
include_once (realpath($entry));
$ok   = realpath($app . '/cache');
$two  = realpath('/tmp', 'x');
```

## Divergences

- **Function name (custos diverges).** Upstream matches the written name
  `realpath` case-sensitively and without resolution, so `RealPath(...)` is
  missed while a namespaced user function named `realpath` is reported (and
  "fixed" away). custos compares the name case-insensitively and requires
  the call to reach the builtin (D1).
- R1 strips a leading `/..` even when it is part of a longer name
  (`'/..hidden'` → `dirname(L) . 'hidden'`). Recommendation: only strip
  `/..` when followed by `/` or the end of the literal; no fixture covers
  this.
- R1's result is a concatenation; inserting it where a tighter-binding
  operator surrounds the call (`!realpath(...)`, `realpath(...)[0]`) changes
  precedence. Recommendation: wrap `R` in parentheses in such contexts.
- **custos diverges from upstream** on relative literals. Upstream replaces
  `realpath('../x')` with `'../x'`. That is not equivalent: `realpath()`
  resolves against the current working directory, while the bare string is
  later resolved by whatever consumes it (for `include`/`require`, the
  include path first, then the script's directory). custos still reports
  such calls but with the generic message and no fix (R2/R3); only absolute
  literals get the literal replacement.
- **Builtin spelling (custos diverges).** Upstream always inserts a bare
  `dirname(`, which a namespaced or imported function of that name captures
  (the fix then calls user code). custos writes `\dirname(` in that case
  (R1).
- **Whole segments (custos diverges).** R1 only climbs for `/..` followed
  by `/` or the end of the literal: `realpath(__DIR__ . '/..cache')` names a
  directory, not the parent, and is reported without a fix. When nothing
  remains after the climbs the replacement is `dirname(L)` without a
  `. ''` tail.
- **Trailing separators (custos diverges).** realpath() returns the path
  without a trailing `/`: `realpath(__DIR__ . '/../')` is
  `dirname(__DIR__)`, not `dirname(__DIR__) . '/'` (which made Magento's
  `$root . '/app/bootstrap.php'` read `//app`). R1 drops trailing `/` from
  the remaining tail (no tail left → no concatenation); R2 offers no fix
  for an absolute literal ending in `/` (other than `/` itself).
- **Tested results keep realpath() (custos diverges, fix).** `realpath()`
  returns `false` for a missing path; the `dirname()` replacement never
  does. When the call's value is tested for failure — negated, cast to
  bool, compared, used as a condition, a logical or `?:` operand, or
  assigned to a target that a later statement of the same list tests that
  way (`$this->root = realpath($r . '/../public'); if (!$this->root) {
  throw …; }`, Akeneo) — the finding is reported without a fix: the
  existence check would silently go dead.
- **Symlinked bases (custos diverges from F1).** `realpath($base . '/..')`
  resolves `$base` first, so through a symlink it names the parent of the
  link's *target*, while `dirname($base)` names the parent of the link
  (TYPO3's `typo3_src` symlink in `CoreUpdateService`). The R1 rewrite is
  offered as a fix only when the base is `__DIR__`, `__FILE__` or
  `dirname()` of them (PHP resolves symlinks in those); other bases are
  reported without a fix.
