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
- **E4** (custos) The inclusion is followed, in the same statement list, by
  code that may depend on the included file having run. Function and class
  declarations after the inclusion are not looked into. Such code is:
  - `new`, a method call or a static call;
  - a call to a function that is not a PHP builtin, or a dynamic call
    (`$fn()`);
  - a read of a variable that the script has not assigned yet at that point
    (`require 'paths.php'; foreach ($paths as $p) {}`). The read may be a
    plain use, a compound update (`$n++`, `$s .= 'x'`), an element or
    property write (`$config['debug'] = true`), a closure `use ($v)` or a
    use inside an arrow function. An assignment counts only once its
    right-hand side has been evaluated, so
    `$paths = array_merge($paths, $more)` still reads the included file's
    `$paths`. Variables are looked up in the variable scope holding the
    inclusion (top-level code or the enclosing function); parameters,
    `$this`, the superglobals, `$argv` and `$argc` are always available.
    Unlike the other items, a variable read counts anywhere after the
    inclusion up to the end of the enclosing namespace block, function or
    file, not only in the same statement list
    (`if ($warm) { require 'paths.php'; } foreach ($paths as $p) {}`);
  - a read of a constant that is neither `true`/`false`/`null`, nor a PHP
    builtin, nor declared earlier in the script with `const` or
    `define('NAME', …)`. `defined('NAME')` and `constant('NAME')` with a
    string literal count as reads of that constant. Class constants are not
    considered.

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
  Comments: when a comment sits between the keyword and the argument, or
  inside the parentheses that would be removed, only the keyword (and the
  whitespace after it) is replaced, so the comment is kept: a fully
  parenthesised argument keeps its parentheses as the call's
  (`require( // why` + newline + `'a.php')` →
  `opcache_compile_file( // why` + newline + `'a.php')`), otherwise `(` is
  written after the function name and `)` after the argument
  (`require /* first */ 'a.php'` →
  `opcache_compile_file(/* first */ 'a.php')`).
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
$base = __DIR__ . '/lib/';
<warning descr="Use opcache_compile_file() in a preload script instead of require.">require __DIR__ . '/src/Kernel.php'</warning>;
<warning descr="Use opcache_compile_file() in a preload script instead of include_once.">include_once($base . 'helpers.php')</warning>;
<warning descr="Use opcache_compile_file() in a preload script instead of require_once.">require_once ( 'lib/functions.php' )</warning>;
$map = require 'classmap.php';
return include 'tail.php';
```

```php
<?php
$base = __DIR__ . '/lib/';
opcache_compile_file(__DIR__ . '/src/Kernel.php');
opcache_compile_file($base . 'helpers.php');
opcache_compile_file('lib/functions.php');
$map = require 'classmap.php';
return include 'tail.php';
```

## Divergences

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
- **Inclusions whose effects later code needs (custos diverges).** Upstream
  reports an inclusion whatever follows it. custos does not when the
  following statements instantiate a class, call a method or a function
  that is not a PHP builtin, or read a variable or constant the script has
  not defined itself (E4): the included file may set up what that code
  needs (EspoCRM's `preload.php`: `include "bootstrap.php";` registers the
  autoloader, then `(new Application())->run(Preload::class);`; or
  `require 'paths.php'; foreach ($paths as $p) { … }`). Compiling the file
  instead of running it breaks the script.
- **Comments (custos diverges).** Upstream rebuilds the statement from the
  argument alone, which drops comments written between the keyword and the
  argument. custos keeps them (F1).
