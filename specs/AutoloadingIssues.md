---
id: AutoloadingIssues
group: Probable bugs
kind: syntax
needs: []
php: { min: "", max: "" }
---

# AutoloadingIssues

## Summary
PSR-0/PSR-4 autoloaders locate a class by its name, so a file that declares a
single class whose name does not match the file name will not be found by the
autoloader. Point at the class name when file and class names disagree.

## Detection
The rule works per file, using only the file's base name (no directory) and
the declarations in it.

- **D1** The file base name ends with `.php` (case-sensitive).
- **D2** The base name is not exactly `index.php` and not exactly
  `actions.class.php`.
- **D3** The base name does not fully match the "timestamped migration" shape:
  4 digits `_` 2 digits `_` 2 digits `_` 6 digits `_` then at least one
  character, then `.php` (regex `^\d{4}_\d{2}_\d{2}_\d{6}_.+\.php$`).
- **D4** Collect the class-like declarations of the file that are top-level
  definitions: classes (incl. abstract/final), interfaces, traits (and enums)
  declared at file level or directly inside a `namespace` statement/block.
  Functions, constants and `use` imports are ignored; anonymous classes are
  not declarations. Continue only when **exactly one** such declaration
  exists.
- **D5** Let `N` be that declaration's short name as written, and let
  `E` = the file base name up to (not including) its **first** `.`
  (`Foo.class.php` → `Foo`, `bar.php` → `bar`).
- **D6** PSR-0 extraction: when the class is in the global namespace (its FQN
  has exactly one backslash, the leading one) and `N` contains `_`, let
  `X` = the part of `N` after its last `_` (`Vendor_Pkg_Item` → `Item`).
  Otherwise `X` = `N`.
- **D7** Report when `E != X` **and** `E != N` (both case-sensitive), unless
  E1 applies.

## Exceptions (no report)
- **E1** WordPress naming: the base name ends with `class-` + `lower(N)` with
  every `_` replaced by `-`, + `.php` (e.g. class `Shop_Cart_Item` in
  `class-shop-cart-item.php`). The comparison is case-sensitive against the
  lowered name, so the file name part must be lowercase.
- **E2** Files not ending in `.php`, `index.php`, `actions.class.php`,
  timestamped-migration file names (D1–D3).
- **E3** Zero or two-plus class-like declarations in the file (e.g. an
  interface plus a class).
- **E4** Namespaced classes never get the PSR-0 `_` extraction: in namespace
  `App`, class `Pkg_Item` in `Item.php` is reported, in `Pkg_Item.php` it is
  not.

## Report
- Range: the class name identifier of the declaration (just `N`, not the
  `class` keyword or the body).
- Severity: warning.
- Message: `File name does not match the class name; autoloading may fail.`

## Fix
None.

## Options
None.

## PHP versions
None.

## Examples
Each block is a separate file; its base name is given in the comment line
above the block.

`Invoice.php` — no report:
```php
<?php
namespace Billing;

final class Invoice
{
}
```

`invoice.php` — reported (case differs):
```php
<?php
class <warning descr="File name does not match the class name; autoloading may fail.">Invoice</warning>
{
}
```

`Ledger.inc.php` — no report via PSR-0 extraction (`E` = `Ledger`):
```php
<?php
abstract class Accounts_Books_Ledger
{
}
```

`Journal.inc.php` — reported:
```php
<?php
trait <warning descr="File name does not match the class name; autoloading may fail.">Accounts_Books_Ledger</warning>
{
}
```

`Pair.php` — no report (two declarations):
```php
<?php
interface PairContract {}
class Pair implements PairContract {}
```

`2031_04_17_093015_add_orders_table.php` — no report (migration name):
```php
<?php
class AddOrdersTable {}
```

`class-shop-cart-item.php` — no report (WordPress naming):
```php
<?php
class Shop_Cart_Item {}
```

## Divergences
- Conditional declarations (a class declared inside an `if` at file level) and
  whether they count as top-level definitions are unverified upstream.
  Recommendation: count only declarations that are direct statements of the
  file or of a namespace body.
- Upstream compares only the first `.`-separated segment of the file name, so
  `Invoice.php.dist` style names are not .php files (D1 fails) and are never
  reported; keep this.
