---
id: DeprecatedFeature
group: Language level migration
kind: semantic
needs: [names, index, hierarchy, types, stubs]
php: { min: "8.0", max: "" }
---

# DeprecatedFeature

## Summary

Report code that uses a feature PHP deprecates at the target version: PHP
emits a deprecation notice for it, and a later major version removes it.
The rule covers PHP's own functions, methods and constants deprecated from
PHP 8.0 on, and the language deprecations of PHP 8.6.

## Detection

- D0. A function, method or constant counts when PHP deprecates it at or
  before the target version: from PHP 8.0, for instance `strftime()` (8.1),
  `utf8_encode()` (8.2), `lcg_value()` (8.4), `curl_close()` and
  `ReflectionProperty::setAccessible()` (8.5), `FILTER_SANITIZE_STRING`
  (8.1), `E_STRICT` (8.4) or `PDO::MYSQL_ATTR_COMPRESS` (8.5). A call
  through a first-class callable or a partial application refers to it too.
  The rest of this list gives the deprecations of PHP 8.6.
- D1. A call to one of PHP's own functions deprecated in PHP 8.6, resolved
  through the file's namespace and imports: `spl_classes()`,
  `spl_object_hash()`, `metaphone()`, `strcoll()`, `mysqli_get_charset()`,
  `mysqli_stmt_init()` and the multibyte regular expression functions
  (`mb_ereg()`, `mb_eregi()`, `mb_ereg_replace()`, `mb_eregi_replace()`,
  `mb_ereg_replace_callback()`, `mb_ereg_match()`, `mb_split()`, the
  `mb_ereg_search` family, `mb_regex_encoding()` and
  `mb_regex_set_options()`).
- D2. A call to one of PHP's own methods deprecated in PHP 8.6, on a
  receiver whose type is known to be the class or a subclass that does not
  override the method: `getFlags()`, `setFlags()`, `asort()`, `ksort()`,
  `uasort()`, `uksort()`, `natsort()`, `natcasesort()`, `serialize()` and
  `unserialize()` of `ArrayIterator`; `fgetcsv()`, `fputcsv()`,
  `setCsvControl()` and `getCsvControl()` of `SplFileObject`;
  `get_charset()` and `stmt_init()` of `mysqli`. A call through the class,
  `self` or `parent` (`parent::__wakeup()`) counts when PHP's own
  declaration runs.
- D3. A `new mysqli_stmt(...)` that passes no query.
- D4. A use of the constants `SORT_LOCALE_STRING` and
  `MB_ONIGURUMA_VERSION`, and of a class constant PHP's own class declares
  deprecated.
- D5. A call to `define()` with a third argument.
- D6. A call to `is_a()` whose first argument is known to be a string and
  whose `allow_string` argument is absent or `false`, or to
  `is_subclass_of()` whose first argument is known to be a string and whose
  `allow_string` argument is `false`.
- D7. A call to `array_walk()`, `array_walk_recursive()` or
  `http_build_query()` whose first argument is known to be an object.
- D8. A `return` statement inside a `finally` block, in the same function.
- D9. A `return` with a value, or a `yield`, in a constructor or a
  destructor.
- D10. A `__debugInfo()` method declaring a nullable return type
  (`?array`, `array|null`, `null`). `mixed`, which PHP does not deprecate
  there, is not concerned.
- D11. A class constant named `namespace`.
- D12. A declaration named `let` or `is`, in any case: a class, interface,
  trait, enum, function or constant (`const`, or `define()` with that name
  as a literal, in any namespace), an alias given by `class_alias()` with a
  literal name, or an alias in a `use`, `use function` or `use const`
  declaration.
- D13. A function named `readonly`, in any case.
- D14. A constant named `_` (`const`, or `define()` with a literal name), or
  `_` as the alias of a class or a constant in a `use` declaration.

## Exceptions (no report)

- E1. Targets below the version that deprecates the feature; below PHP 8.0,
  nothing.
- E2. A function, method, class or constant the project declares in place of
  PHP's own (a polyfill, an override in a subclass, a namespaced function
  that shadows the global one).
- E2a. A use of a deprecated function, method or constant that a check of
  the PHP version guards (`PHP_VERSION_ID`, `version_compare()`, ...), or a
  check that the function exists, as for UnavailableFunction: the code may
  run only on versions where the feature is not deprecated.
- E3. `is_double()`, `is_long()`, `is_integer()`, `doubleval()`,
  `socket_set_timeout()` and `mysqli_execute()`, which AliasFunctionsUsage
  already reports with their replacements.
- E4. A receiver or argument whose type is unknown or only possibly the
  deprecated one; a name that is not a literal.
- E5. A `return` in a function declared inside the `finally` block; a
  `return` without a value in a constructor or destructor; a method, class
  constant or property named `let`, `is` or `readonly`; a function imported
  as `_`.

## Report

- Range: the function or method name of a call, the constant, the `define()`
  argument or the declared name; the `return` keyword through its value; the
  `yield` expression; the return type of `__debugInfo()`.
- Severity: warning.
- Message: what to change, with the deprecation and PHP's replacement when
  it names one, for instance `Replace spl_object_hash(), deprecated as of
  PHP 8.6: use spl_object_id().` or `Move this return out of the finally
  block: returning there is deprecated as of PHP 8.6.`

## Fix

None.

## Options

None.

## PHP versions

Active from PHP 8.0. Each function, method or constant is reported from the version that deprecates it; the other constructs listed, from PHP 8.6.

## Examples

```php
<?php
$id = <warning descr="Replace spl_object_hash(), deprecated as of PHP 8.6: use spl_object_id().">spl_object_hash</warning>($order);

function total(array $lines): int
{
    try {
        return array_sum($lines);
    } finally {
        <warning descr="Move this return out of the finally block: returning there is deprecated as of PHP 8.6.">return 0;</warning>
    }
}
```

## Divergences

Native custos inspection, independently specified; no EA counterpart or EA
conformance requirement.
