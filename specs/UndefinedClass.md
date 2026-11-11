---
id: UndefinedClass
group: Probable bugs
kind: semantic
needs: [names, index, stubs, composer]
php: { min: "", max: "" }
---

# UndefinedClass

## Summary

Report a class, interface, trait or enum name that nothing declares, used
where PHP must load it: not the project, not an installed Composer package
and not PHP or a known extension. Such code throws an `Error` when it runs;
the name usually has a typo or a missing `use` import. The rule needs a
complete view of the code the project can load, so it reports nothing
unless every Composer dependency is installed.

## Detection

- D1. A name PHP loads when the code runs or the class is declared: the
  class of a `new`, of a static method call, of a static property access or
  of a class constant access other than `::class`; a parent class, an
  implemented or extended interface; a used trait. Names resolve with the
  file's namespace and imports; `self`, `static` and `parent` are skipped.
- D2. No class-like of that name is declared by the analysed project, its
  installed packages, PHP or the extensions the stubs describe, at any PHP
  version, directly or through a `class_alias()` call with literal names.
- D3. The project's world is complete, as for `UndefinedFunction`:
  `composer.json` and the vendor directory's `composer/installed.json`
  exist; every package the project requires, suggests or requires for
  development is installed (or replaced or provided by an installed
  package); every installed package that autoloads code was indexed; and
  every classmap and files entry of the project's own autoload
  configuration lies under the analysed paths.

## Exceptions (no report)

- E1. A project without Composer metadata, or whose installed packages do
  not match its requirements (D3): the class may come from code that was
  not seen.
- E2. A file that asks for the class by name with `class_exists()`,
  `interface_exists()`, `trait_exists()` or `enum_exists()` (a literal name
  or a `::class` fetch), or a name inside a branch whose condition checks
  existence, a loaded extension, a constant's definition or the PHP version
  (a conditional declaration or guarded code).
- E3. A namespaced name in a namespace that a PHP extension uses (such as
  `MongoDB\` or `Swoole\`), and a name under a PSR-4 or PSR-0 prefix that
  the project autoloads from a directory outside the analysed paths.
- E4. A global name when an indexed file calls `class_alias()` with computed
  names, which may declare any global alias (framework facades).
- E5. Names PHP does not load: `instanceof` and `catch` (they compare
  names; libraries use them for optional dependencies), parameter, return
  and property types (checked only against a value, and used for optional
  dependencies), `::class` (it only spells the name), attributes (loaded
  only when read through reflection), imports and PHPDoc.
- E6. Dynamic class names (`new $class`, `$class::create()`) and names in
  strings.

## Report

- Range: the class name.
- Severity: error.
- Message: `Fix the name or install the package declaring App\Model\Usr.`

## Fix

None: the intended class cannot be guessed.

## Options

None.

## PHP versions

All versions. A class PHP declares in another version than the configured
one counts as declared; other rules report version mismatches.

## Examples

```php
<?php
namespace App;

use App\Model\User;

$user = new <error descr="Fix the name or install the package declaring App\Model\Usr.">Model\Usr</error>();
class Admin extends <error descr="Fix the name or install the package declaring App\BaseUser.">BaseUser</error> {}
```

Optional dependencies named where PHP loads nothing, or guarded, are not
reported:

```php
<?php
function handle(\Throwable $e) {
    if ($e instanceof \Sentry\EventException) {
        return;
    }
    if (class_exists(\Monolog\Logger::class)) {
        new \Monolog\Logger('app');
    }
}
```

## Divergences

Native custos rule. No upstream inspection or fixture comparison applies.
These conditions are independently specified and examples are original.
