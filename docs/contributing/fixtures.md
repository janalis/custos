# Fixtures

Every rule has its own fixtures in `testdata/rules/<ID>/`. They are the CI
gate (`make fixtures`) and, with the rule packages' tests, must cover 100 %
of the rule's statements (`make coverage`).

## Files

```text
testdata/rules/OneTimeUseVariables/
├── basic.php              # findings, marked up
├── basic.fixed.php        # expected source after `custos fix`
├── false-positives.php    # code that must not be reported
├── options.php            # same rule with non-default options…
├── options.json           # …set by this sidecar
└── old.php / old.json     # behaviour on an older PHP version
```

- `<name>.php`: a fixture. Every expected finding is marked up (see below);
  anything not marked must not be reported.
- `<name>.fixed.php` (optional): the exact expected result of applying all
  quick-fixes to `<name>.php`, markup removed.
- `<name>.json` (optional): settings for that fixture.
- `*.inc`: extra project files whose symbols are indexed next to a fixture
  (see `companions`); they are not fixtures themselves.

A typical set is `basic.php`, `false-positives.php` and `basic.fixed.php`
when the rule has a fix, plus one fixture per option or PHP version that
changes the behaviour. Cover parser-recovery guards with broken PHP in
`broken*.php`.

## Markup

Wrap the highlighted code in a tag named after the severity, with the exact
message in `descr`:

```php
<?php
function total($a) {
    <warning descr="Variable $total is used only once; inline its value.">$total</warning> = $a + 1;
    return $total;
}
```

- Tags: `error`, `warning`, `weak_warning` (reported as `info`) and `info`.
- The range must match the finding exactly, and the message is compared
  exactly.
- Tags may nest. A `<caret>` marker is accepted and ignored.
- `descr` is HTML-unescaped: write `&quot;` for `"` and `&lt;` for `<`.

## Sidecar settings

```json
{
  "php": "7.4",
  "comparisonStyle": "yoda",
  "options": { "ALLOW_LONG_STATEMENTS": true, "configuration": ["\\App\\dump"] },
  "companions": ["Base.inc"]
}
```

| Key | Meaning |
|---|---|
| `php` | Target PHP version for this fixture. |
| `comparisonStyle` | `regular` or `yoda`. |
| `options` | Rule options (bool, number, string or list). |
| `calls` | Rarely needed: list options given as registration calls. |
| `companions` | Extra project files (relative to the rule directory) to index. |

## Rules on other files

Rules that read a manifest instead of PHP (such as `SecurityAdvisories` on
`composer.json`) use one directory per case:
`<case>/composer.json`, with optional `composer.fixed.json` and
`composer.config.json`.

## Running

```sh
make fixtures RULE=OneTimeUseVariables
make coverage            # lists uncovered blocks of internal/rules
```
