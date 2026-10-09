---
id: PhpUnitDeprecations
group: PHPUnit
kind: syntax
needs: []
php: { min: "", max: "" }
---

# PhpUnitDeprecations

## Summary

Flags PHPUnit assertion APIs deprecated in the configured PHPUnit version:
the optional trailing parameters of `assertEquals()` / `assertNotEquals()`
(deprecated in 8.0 in favour of dedicated assertions) and the
`assertFileNotExists()` / `assertDirectoryNotExists()` assertions (deprecated
in 9.1 in favour of their `...DoesNotExist()` counterparts).

## Detection

Visit every method call (`->`, `?->` or `::`; any receiver — `$this`, `self`,
`static`, `parent`, a variable). The call is **not** resolved: matching is by
the method name, compared case-insensitively as PHP compares method names
(`AssertEquals`, `assertfilenotexists` match). Messages and replacement names
use the declared spelling (`assertEquals…`, `assertFileDoesNotExist`). No
test-context check.

Let `V` be the `PHP_UNIT_VERSION` option (unset → inferred from the indexed
PHPUnit, else `PHPUNIT80`; see Divergences). Versions are
ordered `PHPUNIT70 < 71 < 72 < 73 < 74 < 75 < 80 < 81 < 82 < 83 < 84 < 85 <
90 < 91 < 92 < 93 < 94 < 95`.

### Trailing `assertEquals` parameters (V ≥ PHPUNIT80)

- **D1** Method name is `assertEquals` or `assertNotEquals` and the call has
  **more than 3** arguments (spread `...$x` counts as one argument).
- **D2** For each 0-based argument position that exists, report that
  argument (each independently; up to four findings per call):
  - index 3 (`$delta`) → finding **DELTA**, replacement API
    `<name>WithDelta` (`assertEqualsWithDelta` / `assertNotEqualsWithDelta`);
  - index 4 (`$maxDepth`) → finding **DEPTH**, no replacement (parameter
    simply dropped in PHPUnit 8);
  - index 5 (`$canonicalize`) → finding **CANON**, replacement
    `<name>Canonicalizing`;
  - index 6 (`$ignoreCase`) → finding **CASE**, replacement
    `<name>IgnoringCase`.
  Arguments beyond index 6 are ignored. The argument's value is irrelevant
  (`0.0`, `false`, a variable, an assignment expression — all reported).

### Renamed file/directory assertions (V ≥ PHPUNIT91)

- **D3** Method name is exactly `assertFileNotExists` or
  `assertDirectoryNotExists` (any number of arguments). Report the name
  identifier (finding **RENAMED**). Replacement name: see F1.

## Exceptions (no report)

- **E1** V below `PHPUNIT80`: nothing at all is reported.
- **E2** V below `PHPUNIT91`: D3 is not reported (D1/D2 still apply).
- **E3** `assertEquals`/`assertNotEquals` with 3 or fewer arguments.
- **E4** Plain function calls (`assertEquals(...)` without receiver), and other
  deprecated PHPUnit APIs (not covered by this rule).

## Report

- Range:
  - DELTA / DEPTH / CANON / CASE: the whole argument expression at that
    position (e.g. `0.01`, `$tolerance`, `$depth = 5`).
  - RENAMED: the method-name identifier only (e.g. `assertFileNotExists`).
- Severity: info (rule default; upstream renders it as "deprecated"
  strike-through, which maps to info).
- Messages (our wording; `{name}` is the called method):
  - DELTA: `PHPUnit 8.0 deprecated the delta argument; call {name}WithDelta() instead.`
  - DEPTH: `PHPUnit 8.0 deprecated the maxDepth argument; drop it.`
  - CANON: `PHPUnit 8.0 deprecated the canonicalize argument; call {name}Canonicalizing() instead.`
  - CASE: `PHPUnit 8.0 deprecated the ignoreCase argument; call {name}IgnoringCase() instead.`
  - RENAMED: `{name}() was deprecated by PHPUnit 9.1; call {replacement}() instead.`

## Fix

- DELTA / DEPTH / CANON / CASE: no fix.
- **F1** RENAMED: rename the method identifier, leaving receiver, arrow,
  arguments, whitespace and comments untouched:
  - `assertFileNotExists` → `assertFileDoesNotExist`
  - `assertDirectoryNotExists` → `assertDirectoryDoesNotExist`
  (These are the real PHPUnit ≥ 9.1 names. Upstream produces a different,
  non-existent name — see Divergences.)

`internal/inspection/meta/rules.json` lists `hasFix: false` for this rule although the
upstream inspection offers F1; treat the rule as fixable.

## Options

| Option | Type | Default | Effect |
|---|---|---|---|
| `PHP_UNIT_VERSION` | enum (`PHPUNIT70` … `PHPUNIT95`, see ordering above) | `PHPUNIT80` | PHPUnit version targeted by the project. D1/D2 need ≥ `PHPUNIT80`; D3 needs ≥ `PHPUNIT91`. Unset → inferred from the indexed PHPUnit (see Divergences). |

## PHP versions

No PHP-level gating; only the PHPUnit version option matters.

## Examples

With `PHP_UNIT_VERSION = PHPUNIT80`:

```php
<?php

final class InvoiceTest
{
    public function testTotals()
    {
        $this->assertEquals(10.5, $sum, 'sum differs', <weak_warning descr="PHPUnit 8.0 deprecated the delta argument; call assertEqualsWithDelta() instead.">0.01</weak_warning>);
        self::assertNotEquals(
            [1, 2], $ids, '',
            <weak_warning descr="PHPUnit 8.0 deprecated the delta argument; call assertNotEqualsWithDelta() instead.">0</weak_warning>,
            <weak_warning descr="PHPUnit 8.0 deprecated the maxDepth argument; drop it.">$depth</weak_warning>,
            <weak_warning descr="PHPUnit 8.0 deprecated the canonicalize argument; call assertNotEqualsCanonicalizing() instead.">true</weak_warning>,
            <weak_warning descr="PHPUnit 8.0 deprecated the ignoreCase argument; call assertNotEqualsIgnoringCase() instead.">$fold = true</weak_warning>
        );
        $this->assertEquals('a', $label, 'label differs');
        $this->assertFileNotExists('/tmp/out.csv');
    }
}
```

With `PHP_UNIT_VERSION = PHPUNIT91`:

```php
<?php

final class ExportTest
{
    public function testCleanup()
    {
        $this-><weak_warning descr="assertFileNotExists() was deprecated by PHPUnit 9.1; call assertFileDoesNotExist() instead.">assertFileNotExists</weak_warning>($this->path);
        static::<weak_warning descr="assertDirectoryNotExists() was deprecated by PHPUnit 9.1; call assertDirectoryDoesNotExist() instead.">assertDirectoryNotExists</weak_warning>(dirname($this->path), 'still there');
        $this->assertFileDoesNotExist($this->path);
    }
}
```

```php
<?php

final class ExportTest
{
    public function testCleanup()
    {
        $this->assertFileDoesNotExist($this->path);
        static::assertDirectoryDoesNotExist(dirname($this->path), 'still there');
        $this->assertFileDoesNotExist($this->path);
    }
}
```

## Divergences

- **PHPUnit version when `PHP_UNIT_VERSION` is unset (custos diverges):**
  upstream assumes `PHPUNIT80`, so a PHPUnit 7 project is told that
  `assertEquals`' delta argument is deprecated. When the option is not
  configured, custos infers the version from the indexed PHPUnit (vendor
  included): `PHPUnit\Framework\Assert` declares `assertMatchesRegularExpression` →
  `PHPUNIT91`; else it declares `assertIsInt` but not `assertInternalType` →
  `PHPUNIT90`; else `assertIsInt` → `PHPUNIT80`; else `PHPUNIT70`. When
  PHPUnit is not indexed either, `PHPUNIT80`.
- **Case of method names (custos diverges from upstream).** Upstream
  matches the method name case-sensitively, so `$this->AssertEquals(1, 2,
  '', 0.1)` is not reported. PHP method names are case-insensitive, so
  custos matches any case and words messages and fixes with the declared
  PHPUnit spelling.
- **Replacement name (F1 / RENAMED message).** Upstream derives the new name
  by substituting `NotExist` with `DoesNotExist`, which yields
  `assertFileDoesNotExists` / `assertDirectoryDoesNotExists` (trailing `s`) —
  methods that do not exist in PHPUnit, so applying the upstream fix breaks
  the test. custos emits the real names `assertFileDoesNotExist` /
  `assertDirectoryDoesNotExist`. Consequence: the EA conformance fixture for
  PHPUnit 9.1 will differ in fix output by exactly that trailing `s`
  (ranges and severities still match). Recommendation: keep the correct
  names and record the case as an accepted conformance divergence.
- PHP 8 named arguments (`assertEquals(1.0, $x, delta: 0.1)`): upstream works
  purely on argument positions; behaviour with named arguments is unverified.
  Recommendation: for named arguments, map by parameter name (`delta`,
  `maxDepth`, `canonicalize`, `ignoreCase`) and report the whole named
  argument (name label included); positional arguments as specified. No
  fixture covers it.
- **Foreign `assertEquals` methods (custos diverges).** Upstream matches
  `assertEquals`/`assertNotEquals` by name on any receiver, so a comparator
  library's own `$comparator->assertEquals($a, $b, $delta, $canonicalize)`
  (another signature, not deprecated) is reported. custos skips the call
  when the receiver's type (or the static class) resolves to classes whose
  method of that name is declared outside the `PHPUnit\` namespace, and an
  instance call on a receiver other than `$this` whose type is unknown
  (PHPUnit assertions are called on the test case, or statically). `$this`,
  `self::`/`static::`, and receivers resolving to PHPUnit's declaration are
  still checked.
- **PHPUnit version from composer.json (custos).** When PHPUnit itself is
  not indexed (no vendor directory, or one installed without dev
  dependencies) and `PHP_UNIT_VERSION` is not configured, the version is
  the lowest one allowed by composer.json's `phpunit/phpunit` constraint
  (`require-dev`, else `require`) before falling back to the option
  default: EspoCRM (`^11.5`) was told to use `assertContains()` (strict
  from 9.0, unlike `in_array()`) and Grav `assertRegExp()` (removed in 10).
