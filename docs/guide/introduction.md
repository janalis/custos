# What is custos?

custos is a static analyzer for PHP. It reads your code, reports problems,
and fixes many of them for you. Its rules are organized in these groups:

- **Probable bugs**: duplicate array keys, `printf()` arguments that do not
  match the format, forgotten debug output, misused extension APIs…
- **Security**: weak randomness, disabled TLS certificate checks,
  `unserialize()` on untrusted data…
- **Performance**: slow array operations in loops, inefficient regular
  expressions, string functions where a cheaper one would do…
- **Control flow**: one-time variables, redundant `else` branches, loops that
  do not loop, conditions that can be simplified…
- **Code style**: redundant casts, nested `!`, needlessly complex conditions…
- **Confusing constructs**: nested ternaries, senseless ternaries, useless
  `return`…
- **Architecture**: exception handling, coupling, inheritance chains and
  class design.
- **Unused**: write-only parameters, unused constructor dependencies, proxy
  methods, unused `goto` labels…
- **Language level migration**: constructs that newer PHP versions replace,
  deprecate or remove, checked against your target PHP version.
- **Compatibility**: deprecated `ini` options, binary-unsafe `fopen()` modes,
  API migrations.
- **Type compatibility**: loose comparisons and type-unsafe array searches.
- **PHPUnit**: assertion and test-case improvements.

See the [rule reference](/rules/) for detection conditions, examples and
available quick-fixes.

## Why custos

- **One binary.** It is written in Go with its own PHP parser and type
  inference, so it needs no PHP runtime and no IDE. Install it with Homebrew
  or Composer, or download it from a release.
- **Fast.** custos first indexes the symbols of your project and its
  `vendor/` directory, then analyzes files in parallel, each in a single pass
  over its syntax tree.
- **Fixes, not just reports.** Rules offer quick-fixes when a safe correction
  can be established. Apply them with `custos fix`, or one at a time from
  your editor.
- **Same findings everywhere.** The CLI, CI, the language server and the
  [MCP server](./mcp) for AI agents share one engine, one configuration file
  (`custos.json`) and one baseline.
- **Version-aware.** Rules are gated on your target PHP version (5.3 to 8.5),
  from `--php`, `custos.json` or `composer.json` (see
  [Target PHP version](./configuration#target-php-version)).

## Relation to Php Inspections (EA Extended)

The catalogue mixes two kinds of rules: rules modeled on the PhpStorm plugin
[Php Inspections (EA Extended)](https://github.com/kalessil/phpinspectionsea)
(*EA rules*), and [native rules](./native-rules) designed for custos. The
[rule reference](/rules/) shows the origin of each rule.

custos is an **independent, clean-room implementation** released under the
MIT license. None of that project's code, messages, descriptions or test
fixtures are included. Every rule was independently specified before it
was implemented (see the [clean-room process](/contributing/clean-room)).

EA rules keep their PhpStorm names, so `@noinspection FooInspection` comments
written for PhpStorm also suppress custos findings. When the upstream
behavior has false positives, or its fixes change semantics or produce
invalid PHP, custos diverges on purpose.

## Glossary

| Term | Meaning |
|---|---|
| Rule | One check, such as `OneTimeUseVariables`. Rules are on or off by default and can be configured. |
| Finding | One problem reported by a rule at a source range. |
| Quick-fix | An automatic edit a rule offers for a finding, applied by `custos fix` or an editor code action. |
| Severity | `error`, `warning` or `info`. Each rule has a default severity, which `custos.json` can change. |
| Rule ID | The custos name of a rule (`OneTimeUseVariables`), used everywhere. |
| PhpStorm name | For EA rules only, the ID followed by `Inspection` (`OneTimeUseVariablesInspection`). It is accepted wherever a rule ID is. |

## custos and other tools

custos works alongside type checkers (PHPStan, Psalm), formatters
(PHP-CS-Fixer, PHP_CodeSniffer) and your main PHP language server
(Intelephense, Phpactor, PhpStorm). It covers patterns that are legal PHP
but probably wrong, slow, insecure or outdated. Through its
[MCP server](./mcp), AI coding agents can run the same checks and apply the
same quick-fixes.
