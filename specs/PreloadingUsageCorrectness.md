---
id: PreloadingUsageCorrectness
group: Probable bugs
kind: syntax
needs: []
php: { min: "", max: "" }
---

# PreloadingUsageCorrectness

## Summary
In an opcache preload script, including files with `require`/`include`
executes them, which can fail on unresolved dependencies (PHP bug #78918).
`opcache_compile_file()` only compiles and caches the file, which is what a
preload script usually wants.

## Detection
- **D1** The file's base name is exactly `preload.php` (case-sensitive; any
  directory).
- **D2** An inclusion expression: `require`, `require_once`, `include` or
  `include_once`, with an argument.
- **D3** The inclusion expression is directly an expression statement
  (`require 'x.php';`). Inclusions used as values (assigned, returned,
  parenthesised, part of a condition) are not reported.

## Exceptions (no report)
- **E1** Any other file name (`Preload.php`, `preload.inc.php`, `boot.php`).
- **E2** `$cfg = require 'config.php';`, `$cfg = (require 'config.php');`,
  `return include 'x.php';`, `if (include 'x.php') {}`.
- **E3** (custos) The argument contains a string literal (or literal part of
  an interpolated string) mentioning `autoload` or `preload`
  (case-insensitive): `require __DIR__.'/vendor/autoload.php';`,
  `require dirname(__DIR__).'/var/cache/prod/App_KernelProdContainer.preload.php';`.
  Such files must run (register the autoloader, execute the generated
  preload list); compiling them does nothing useful.

## Report
- Range: the inclusion expression, from the keyword through the end of its
  argument; the terminating `;` is not included.
- Severity: warning.
- Message: `Use opcache_compile_file() in a preload script instead of {keyword}.`
  (`{keyword}` = `require`, `require_once`, `include` or `include_once`.)

## Fix
- **F1** Replace the inclusion expression with
  `opcache_compile_file(ARG)`, where `ARG` is the source text of the argument
  with any enclosing parentheses removed (`require('a.php')` →
  `opcache_compile_file('a.php')`; `include __DIR__ . '/b.php'` →
  `opcache_compile_file(__DIR__ . '/b.php')`). The statement's `;` and the
  surrounding whitespace are kept.
  Builtin spelling: `opcache_compile_file` is written `\opcache_compile_file` when an unqualified call at
  that position would not reach the global function (a `use function`
  import under that name, or a same-named function declared in the current
  namespace).

## Options
None.

## PHP versions
None (preloading exists from PHP 7.4, but no gating is applied).

## Examples

File `preload.php`:

```php
<?php
<warning descr="Use opcache_compile_file() in a preload script instead of require.">require __DIR__ . '/src/Kernel.php'</warning>;
<warning descr="Use opcache_compile_file() in a preload script instead of include_once.">include_once($base . 'helpers.php')</warning>;
<warning descr="Use opcache_compile_file() in a preload script instead of require_once.">require_once ( 'vendor/autoload.php' )</warning>;
$map = require 'classmap.php';
return include 'tail.php';
```

```php
<?php
opcache_compile_file(__DIR__ . '/src/Kernel.php');
opcache_compile_file($base . 'helpers.php');
opcache_compile_file('vendor/autoload.php');
$map = require 'classmap.php';
return include 'tail.php';
```

## Divergences
None.
- **Builtin spelling (custos diverges).** Upstream inserts a bare `opcache_compile_file(`,
  which a function of that name declared in (or imported into) the current
  namespace captures, so the fix would call user code. custos writes
  `\opcache_compile_file(` in that case.
- **Scripts that must run (custos diverges).** Upstream reports every
  statement-level inclusion, including the Composer autoloader and the
  preload script Symfony generates (`config/preload.php` requires
  `var/cache/prod/*.preload.php`, which itself loads classes and calls
  `Preloader::preload()`). Replacing those with `opcache_compile_file()`
  silently disables preloading (and autoloading). custos skips inclusions
  whose path literal mentions `autoload` or `preload` (E3).
