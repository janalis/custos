<p align="center">
  <img src="docs/public/logo.png" alt="custos" width="160">
</p>

<h1 align="center">custos</h1>

<p align="center">
  <strong>Fast PHP inspector and fixer.</strong><br>
  178 inspections with quick-fixes · one static binary · CLI, CI and LSP
</p>

<p align="center">
  <a href="https://github.com/janalis/custos/actions/workflows/ci.yml"><img src="https://github.com/janalis/custos/actions/workflows/ci.yml/badge.svg" alt="CI"></a>
  <a href="https://janalis.github.io/custos/"><img src="https://github.com/janalis/custos/actions/workflows/docs.yml/badge.svg" alt="Docs"></a>
  <a href="https://github.com/janalis/custos/releases"><img src="https://img.shields.io/github/v/release/janalis/custos" alt="Latest release"></a>
  <a href="https://packagist.org/packages/janalis/custos"><img src="https://img.shields.io/packagist/v/janalis/custos" alt="Packagist"></a>
  <a href="LICENSE"><img src="https://img.shields.io/github/license/janalis/custos" alt="MIT license"></a>
</p>

<p align="center">
  <a href="https://janalis.github.io/custos/"><strong>Documentation</strong></a> ·
  <a href="https://janalis.github.io/custos/guide/getting-started">Getting started</a> ·
  <a href="https://janalis.github.io/custos/rules/">Rules</a> ·
  <a href="https://janalis.github.io/custos/guide/editors">Editors</a>
</p>

---

custos finds probable bugs, performance and security issues, needless
complexity and outdated constructs in PHP code (PHP 5.3 to 8.5), and fixes
many of them for you. It is written in Go with its own PHP parser and type
inference, so it needs no PHP runtime. It runs from the command line, in CI,
and in any editor as a language server.

Its rule catalogue is modelled on *Php Inspections (EA Extended)*, and rule IDs
are compatible: existing `@noinspection XxxInspection` comments keep working.
custos is an independent clean-room implementation (see [NOTICE](NOTICE)).

## Install

```sh
brew install janalis/tap/custos           # macOS / Linux
composer require --dev janalis/custos     # per project → vendor/bin/custos
```

Release archives for Linux, macOS and Windows (amd64/arm64) are on the
[releases page](https://github.com/janalis/custos/releases). See
[Installation](https://janalis.github.io/custos/guide/installation) for
details and building from source.

## Quick start

```sh
custos analyse                            # report problems in the project
custos fix --dry-run --diff src           # preview quick-fixes
custos fix src                            # apply them
custos analyse --generate-baseline custos-baseline.json   # adopt on legacy code
custos explain OneTimeUseVariables        # what a rule does, its options
custos lsp                                # language server over stdio
```

```text
src/Invoice.php:3:5: warning: Variable $total is used only once; inline its value. [OneTimeUseVariables] (fixable)
src/Invoice.php:7:6: error: Restrict the classes unserialize() may create via its second argument. [UnserializeExploits]
```

## Highlights

- **178 rules** in 12 groups: probable bugs, performance, security, control
  flow, code style, unused code, PHPUnit, language-level migration…
  ([reference](https://janalis.github.io/custos/rules/))
- **Quick-fixes** for over a hundred rules, applied by `custos fix` or as
  editor code actions.
- **Version-aware**: rules follow the target PHP version from `custos.json`
  or `composer.json`.
- **CI-ready**: text, JSON, Checkstyle, GitHub annotations and SARIF output;
  baselines for legacy code. ([CI recipes](https://janalis.github.io/custos/guide/ci))
- **Editor integration** through LSP: Neovim, Helix, VS Code, PhpStorm,
  Sublime Text, Emacs. ([setup](https://janalis.github.io/custos/guide/editors))

## Configuration

An optional `custos.json` at the project root:

```json
{
  "php": "8.3",
  "paths": ["src", "tests"],
  "baseline": "custos-baseline.json",
  "rules": {
    "MultipleReturnStatements": { "enabled": false },
    "OneTimeUseVariables": { "options": { "ALLOW_LONG_STATEMENTS": false } }
  }
}
```

All keys are described in
[Configuration](https://janalis.github.io/custos/guide/configuration).

## Contributing

Contributions are welcome: bug reports with a PHP snippet, false positives,
fixes. Read [CONTRIBUTING.md](CONTRIBUTING.md) and the
[contributor guide](https://janalis.github.io/custos/contributing/). Please
note the **clean-room rule**: no code or text from the upstream plugin may
enter this repository.

```sh
make build          # → bin/custos
make verify         # lint + tests + fixtures + 100 % coverage + clean-room scan
```

## License

[MIT](LICENSE). Builtin symbol data from
[JetBrains phpstorm-stubs](https://github.com/JetBrains/phpstorm-stubs)
(Apache-2.0). See [NOTICE](NOTICE).
