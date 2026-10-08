# Configuration

custos reads `custos.json` from the project root. The project root is the
directory given with `--config`, or else the first directory, going up from the
first path, that holds `custos.json` or `composer.json`. Every key is
optional.

```json
{
  "php": "8.3",
  "comparisonStyle": "yoda",
  "shortOpenTag": false,
  "paths": ["src", "tests"],
  "exclude": ["var", "public/build"],
  "baseline": "custos-baseline.json",
  "rules": {
    "MultipleReturnStatements": { "enabled": false },
    "OneTimeUseVariables": {
      "enabled": true,
      "severity": "warning",
      "options": { "ALLOW_LONG_STATEMENTS": false }
    },
    "ForgottenDebugOutput": {
      "options": { "configuration": ["\\App\\Debug::dump"] }
    }
  }
}
```

| Key | Meaning |
|---|---|
| `php` | Target PHP version (`"5.3"` to `"8.5"`). Rules that depend on the language level use it. |
| `comparisonStyle` | `"regular"` (`$x === null`, default) or `"yoda"` (`null === $x`): the operand order that comparison rules expect and that fixes write. |
| `shortOpenTag` | Treat `<?` as an opening tag, as with PHP's `short_open_tag` setting. Default `false`. |
| `paths` | What `custos analyse` and `custos fix` check when no path is given. Relative to the root; must stay inside it. |
| `exclude` | Extra directory names or root-relative paths to skip. |
| `baseline` | Baseline file used by `analyse` (see [Suppressing findings](./suppressing#baseline)). |
| `rules` | Per-rule settings, keyed by rule ID (or PhpStorm inspection name). |

Each `rules` entry accepts:

- `enabled`: `true` or `false`. It overrides the rule's default; the
  [rule reference](/rules/) shows which rules are on by default.
- `severity`: `"error"`, `"warning"` or `"info"`.
- `options`: the rule's options. Each rule page lists them with their
  defaults, as does `custos explain <rule>`.

An unknown rule ID or an invalid `severity` or `comparisonStyle` is a
configuration error (exit code 2), so typos do not go unnoticed.

## Target PHP version

The first of these that is set decides:

1. `--php` on the command line;
2. `php` in `custos.json`;
3. `config.platform.php` in `composer.json`;
4. the lowest version that `require.php` in `composer.json` allows;
5. PHP 8.4.

`custos analyse --stats` shows which version was used and where it came from.

## PHPUnit version

The PHPUnit rules adapt their advice to the PHPUnit version the project
uses: the installed `phpunit/phpunit` when the project's `vendor` is
available, else the lowest version `composer.json` allows (`require-dev`,
then `require`). See the `PHP_UNIT_VERSION` option of
[PhpUnitTests](/rules/phpunit/PhpUnitTests) and
[PhpUnitDeprecations](/rules/phpunit/PhpUnitDeprecations).

## In the language server

`custos lsp` reads the same `custos.json` from the workspace root. Editors can
override it with LSP `initializationOptions`, which take the same JSON shape.
See [Editors](./editors).
