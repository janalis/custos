# What is custos?

custos is a static analyser for PHP. It reads your code, reports problems,
and fixes many of them for you. It looks for:

- **probable bugs**: duplicate array keys, `printf()` arguments that do not
  match the format, suspicious assignments, forgotten debug output, a
  by-reference `foreach` variable altered later…
- **performance** issues: slow array operations in loops, inefficient
  regular expressions, string functions used where a cheaper one would do…
- **security** issues: weak randomness, disabled TLS certificate checks,
  `unserialize()` on untrusted data…
- **simplifications and code style**: redundant casts, nested `!`,
  needlessly complex conditions…
- **language-level migration**: constructs that newer PHP versions replace,
  deprecate or remove, checked against the PHP version your project targets.

There are 378 rules in all, including 200 native custos inspections; see the [rule reference](/rules/).

## Why custos

- **One binary.** It is written in Go with its own PHP parser and type
  inference, so it needs no PHP runtime, no IDE and no indexing step. Install
  it with Homebrew or Composer, or download a release archive.
- **Fast.** Files are analysed in parallel, each in a single pass over its
  syntax tree, with every rule dispatched by node kind.
- **Fixes, not just reports.** Over a hundred rules have a quick-fix. Use
  `custos fix` on the command line, or apply them one at a time from your
  editor.
- **Same diagnostics everywhere.** The CLI, CI and the language server share
  one engine and one configuration file (`custos.json`).
- **Version-aware.** Rules are gated on your target PHP version (5.3 to 8.5),
  read from `custos.json` or `composer.json`.

## Relation to Php Inspections (EA Extended)

178 rules are modelled on the PhpStorm plugin
[Php Inspections (EA Extended)](https://github.com/kalessil/phpinspectionsea).
The catalogue also includes 200 independently designed [native inspections](./native-rules).
custos is an **independent, clean-room implementation** released under the
MIT license. None of that project's code, messages, descriptions or test
fixtures are included. Every rule was independently specified before it
was implemented (see the [clean-room process](/contributing/clean-room)).

The modelled rules retain their identifiers: `@noinspection FooInspection` comments written for
PhpStorm also suppress custos findings. When the upstream behaviour has false
positives, or its fixes change semantics or produce invalid PHP, custos
diverges on purpose.

## custos and other tools

custos works alongside type checkers (PHPStan, Psalm), formatters
(PHP-CS-Fixer, PHP_CodeSniffer) and your main PHP language server
(Intelephense, Phpactor, PhpStorm). It covers the inspection layer: patterns
that are legal PHP but probably wrong, slow, insecure or outdated.
