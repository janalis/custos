# Configuration

custos reads `custos.json` from the project root. Every key is optional.

```json
{
  "php": "8.3",
  "comparisonStyle": "yoda",
  "shortOpenTag": false,
  "flow": { "assertions": false },
  "paths": ["src", "tests"],
  "exclude": ["generated", "public/build"],
  "baseline": "custos-baseline.json",
  "rules": {
    "StaticClosureCanBeUsed": { "enabled": false },
    "OneTimeUseVariables": {
      "severity": "info",
      "options": { "ALLOW_LONG_STATEMENTS": true }
    },
    "ForgottenDebugOutput": {
      "options": { "configuration": ["\\App\\Debug::dump"] }
    }
  }
}
```

| Key | Meaning |
|---|---|
| `php` | Target PHP version: 5.3–5.6, 7.0–7.4, or 8.0–8.5. See [Target PHP version](#target-php-version). |
| `comparisonStyle` | `"regular"` (`$x === null`, default) or `"yoda"` (`null === $x`): the operand order that comparison rules expect and that quick-fixes write. |
| `shortOpenTag` | Treat `<?` as an opening tag, as with PHP's `short_open_tag` setting. Default `false`. |
| `flow.assertions` | Trust `assert()` calls as guards. Default `false`. See [Assertions](#assertions). |
| `paths` | What `custos analyse` and `custos fix` check when no path is given. Relative to the root; must stay inside it. Default: the root. |
| `exclude` | Extra directories to skip. See [Excluded directories](#excluded-directories). |
| `baseline` | Baseline file used by `analyse`, relative to the root (see [Baseline](./suppressing#baseline)). |
| `rules` | Per-rule settings, keyed by rule ID or PhpStorm name. |

Each `rules` entry accepts:

- `enabled`: `true` or `false`. It overrides the rule's default; the
  [rule reference](/rules/) shows which rules are on by default.
- `severity`: `"error"`, `"warning"` or `"info"`, replacing the rule's default
  severity.
- `options`: the rule's options. Each rule page lists them with their
  defaults, as does `custos explain <rule>`.

## Project root

The project root is the first directory, going up from the start directory,
that holds `custos.json` or `composer.json`; if none does, it is the start
directory itself. The start directory is the `--config` directory if given,
else the first path on the command line (its directory, for a file), else
the current directory.

## Validation

custos stops with a configuration error ([exit code](./cli#exit-codes) `2`)
when `custos.json` is not valid JSON, or when it has:

- an unknown rule ID;
- a `severity` other than `error`, `warning` or `info`;
- a `comparisonStyle` other than `regular` or `yoda`;
- an unsupported `php` version;
- a `paths` entry or `baseline` that is absolute or leaves the root.

Unknown top-level keys and unknown rule options are ignored.

## Excluded directories

Directories named `vendor` and `node_modules`, hidden directories (whose name
starts with a dot, such as `.git`, `.idea`, `.custos` or `.phpunit.cache`), and
`var/cache` below each scanned path, are always skipped; a scanned path may
still be a hidden directory itself. `exclude` (or `--exclude a,b` on the command
line) adds more. Each entry is a directory name, skipped at any depth, or a
path relative to each scanned path (not to the root): with `"paths": ["src"]`,
write `"gen"` or `"sub/gen"` to skip `src/sub/gen`, not `"src/sub/gen"`.

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

## Assertions

By default, custos does not treat `assert()` as a guard, because PHP can be
configured to skip assertions. Set `flow.assertions` to `true` only when your
runtime always runs assertions and a failed one stops execution.

## In the language server

`custos lsp` reads the same `custos.json` from the workspace root. Editors can
override it with LSP `initializationOptions`, which take the same JSON shape.
See [Editors](./editors).
