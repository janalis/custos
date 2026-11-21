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
5.6. On the older versions the call throws an `Error` (a fatal error before
PHP 7.0). Raise the project's minimum PHP version, install a polyfill or
write the code without the function. The rule needs a complete view of the
code the project can load, since a polyfill may declare the function, so it
reports nothing unless every Composer dependency is installed.

## Detection

- D1. A call to a function written as a name (`foo()`, `\foo()`), resolved
  with the file's namespace and imports. An unqualified call in a namespace
  is concerned only when no function of the namespaced name exists, so that
  it reaches the global one.
- D2. PHP or a known extension declares the function, but only from a
  version later than the target: no declaration of it is available at the
  target version.
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
  else arm.
- E3. A function PHP removed before the target version (`each()` targeting
  PHP 8.0): an extension installed from PECL may declare it again.
- E4. A function of an extension that does not ship with PHP and whose
  versions follow the extension's own releases rather than PHP's (the New
  Relic agent's).
- E5. Dynamic calls (`$fn()`, `call_user_func('foo')`) and names in strings.

## Report

- Range: the function name.
- Severity: error.
- Message: `str_contains() needs PHP 8.0; the project targets PHP 5.6.`

## Fix

None: raising the minimum PHP version or adding a polyfill is a project
decision.

## Options

None.

## PHP versions

All versions: the target version decides which functions are reported.
Nothing is reported at the latest supported version.

## Examples

```php
<?php
namespace App;

if (<error descr="str_contains() needs PHP 8.0; the project targets PHP 5.6.">str_contains</error>($haystack, 'needle')) {
    echo <error descr="array_key_last() needs PHP 7.3; the project targets PHP 5.6.">array_key_last</error>($items);
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
