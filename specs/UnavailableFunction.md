---
id: UnavailableFunction
group: Probable bugs
kind: semantic
needs: [names, index, stubs, composer]
php: { min: "", max: "" }
---

# UnavailableFunction

## Summary

Report a call to a function PHP only provides from a later version than the
project targets, such as `str_contains()` in a project that supports PHP
5.6, or that PHP removed before it, such as `each()` in a project on PHP
8.x. Where the function is missing the call throws an `Error` (a fatal
error before PHP 7.0). Change the project's PHP versions, install a
polyfill or write the code without the function. The rule needs a complete view of the
code the project can load, since a polyfill may declare the function, so it
reports nothing unless every Composer dependency is installed. A polyfill
Composer does not load (one a bootstrap script includes) counts only when
it lies in the analyzed paths.

## Detection

- D1. A call to a function written as a name (`foo()`, `\foo()`), resolved
  with the file's namespace and imports. An unqualified call in a namespace
  is concerned only when no function of the namespaced name exists, so that
  it reaches the global one.
- D2. An extension that ships with PHP declares the function, but no
  declaration of it is available at the target version: it comes in a
  later version, or was removed in an earlier one.
- D3. Neither the analyzed project nor its installed packages declare a
  function of that name (a polyfill).
- D4. The project's world is complete, as for undefined functions: its
  Composer metadata exists and every package and extension it requires is
  installed and was read.

## Exceptions (no report)

- E1. A project without Composer metadata, or whose installed packages do
  not match its requirements: a polyfill may come from code that was not
  seen.
- E2. A file that asks for the function by name with `function_exists()` or
  `is_callable()` (a literal name), or a call that a check about the PHP
  version or the function guards: a branch whose condition makes the
  check, or an earlier check that exits early, short-circuits or selects an
  else arm, or, in an instance method, a check in the class's constructor
  that throws or exits when it fails.
- E3. A function of an extension that does not ship with the latest PHP:
  from PECL or a vendor, such as the New Relic agent, whose versions may
  follow the extension's own releases rather than PHP's, or one PHP moved
  to PECL (`mysql_connect()`, `mcrypt_encrypt()`, `imap_open()`), whose
  PECL build may still declare what PHP removed.
- E4. A function of an extension PHP bundles that PECL also builds for
  older PHP versions (sodium, zip), when the project or an installed
  package requires that extension (`ext-sodium` in `composer.json`): the
  project may run that build.
- E5. Dynamic calls (`$fn()`, `call_user_func('foo')`) and names in strings.

## Report

- Range: the function name.
- Severity: error.
- Message: `str_contains() needs PHP 8.0; the project targets PHP 5.6.`, or
  for a removed function `each() was removed in PHP 8.0; the project
  targets PHP 8.4.`

## Fix

None: raising the minimum PHP version or adding a polyfill is a project
decision.

## Options

None.

## PHP versions

All versions: the target version decides which functions are reported.
At the latest supported version only removed functions are, and nothing
is reported when no target
version is configured (the command line, the custos configuration,
`composer.json` or `.php-version`): the fallback version says nothing about
where the code runs.

## Examples

```php
<?php
namespace App;

if (<error descr="str_contains() needs PHP 8.0; the project targets PHP 5.6.">str_contains</error>($haystack, 'needle')) {
    echo <error descr="array_key_last() needs PHP 7.3; the project targets PHP 5.6.">array_key_last</error>($items);
}
```

In a project on PHP 8.4:

```php
<?php
while (list($key, $value) = <error descr="each() was removed in PHP 8.0; the project targets PHP 8.4.">each</error>($items)) {
    echo $key;
}
```

Guarded calls and functions available at the target are not reported:

```php
<?php
if (function_exists('array_key_first')) {
    $first = array_key_first($items);
}
if (PHP_VERSION_ID >= 80000) {
    $found = str_contains($haystack, 'needle');
}
$length = strlen($haystack);
```

## Divergences

Native custos inspection, independently specified; no EA counterpart or EA conformance requirement.
