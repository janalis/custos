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
  "flow": { "assertions": false },
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
| `php` | Target PHP version: 5.3–5.6, 7.0–7.4, or 8.0–8.5. Rules that depend on the language level use it. |
| `comparisonStyle` | `"regular"` (`$x === null`, default) or `"yoda"` (`null === $x`): the operand order that comparison rules expect and that fixes write. |
| `shortOpenTag` | Treat `<?` as an opening tag, as with PHP's `short_open_tag` setting. Default `false`. |
| `flow.assertions` | Assume direct calls to PHP's `assert` are enabled and failed assertions terminate execution. Default `false`. |
| `paths` | What `custos analyse` and `custos fix` check when no path is given. Relative to the root; must stay inside it. |
| `exclude` | Extra directory names or root-relative paths to skip. |
| `baseline` | Baseline file used by `analyse` (see [Suppressing findings](./suppressing#baseline)). |
| `rules` | Per-rule settings, keyed by rule ID (or PhpStorm inspection name). |

By default, `assert()` does not establish a guard. On PHP 7 and later, its
arguments may be skipped at runtime, so assignments and sanitizers inside those
arguments cannot establish definite facts either. Dynamic callable invocations
evaluate their arguments as ordinary calls. Set `flow.assertions` to `true` only
when your runtime guarantees that assertions run and a failed assertion stops
execution. This setting also applies to inference, project summaries and LSP
analysis; the LSP rebuilds dependent summaries when it changes.

Before PHP 8, string assertions can execute code in the current scope. custos
does not interpret that code. A string or an expression whose result is not
syntactically known to be non-string invalidates local proofs, even when
assertion assumptions are enabled. Native type predicates such as
`assert(is_string($value))` and strict comparisons such as
`assert($value !== false)` remain usable. On PHP 8 and later, a plain variable
assertion may also supply a guard.

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
