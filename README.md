<p align="center">
  <img src="docs/public/logo.png" alt="custos" width="160">
</p>

<h1 align="center">custos</h1>

<p align="center">
  <strong>Fast PHP inspector and fixer.</strong><br>
  Rules with quick-fixes · one static binary · CLI, CI, LSP and MCP
</p>

<p align="center">
  <a href="https://github.com/janalis/custos/actions/workflows/ci.yml"><img src="https://github.com/janalis/custos/actions/workflows/ci.yml/badge.svg" alt="CI"></a>
  <a href="https://janalis.github.io/custos/"><img src="https://github.com/janalis/custos/actions/workflows/docs.yml/badge.svg" alt="Docs"></a>
  <a href="https://github.com/janalis/custos/releases"><img src="https://img.shields.io/github/v/release/janalis/custos?sort=semver" alt="Latest release"></a>
  <a href="https://packagist.org/packages/janalis/custos"><img src="https://img.shields.io/packagist/v/janalis/custos" alt="Packagist"></a>
  <a href="LICENSE"><img src="https://img.shields.io/github/license/janalis/custos" alt="MIT license"></a>
</p>

<p align="center">
  <a href="https://janalis.github.io/custos/"><strong>Documentation</strong></a> ·
  <a href="https://janalis.github.io/custos/guide/getting-started">Getting started</a> ·
  <a href="https://janalis.github.io/custos/rules/">Rules</a> ·
  <a href="https://janalis.github.io/custos/guide/editors">Editors</a> ·
  <a href="https://janalis.github.io/custos/guide/mcp">AI agents</a>
</p>

---

custos finds probable bugs, performance and security issues, needless
complexity and outdated constructs in PHP code (PHP 5.3 to 8.5), and fixes
many of them for you. It is written in Go with its own PHP parser and type
inference, so it needs no PHP runtime. It runs from the command line, in CI,
in editors as a language server, and with AI agents through MCP.

The catalogue mixes rules modeled on *Php Inspections (EA Extended)*, whose
existing `@noinspection XxxInspection` comments keep working, and native rules
designed for custos. custos is an independent clean-room implementation (see
[NOTICE](NOTICE)).

## Install

```sh
brew install janalis/tap/custos           # macOS / Linux
composer require --dev janalis/custos     # per project → vendor/bin/custos
```

Archives and bare binaries for Linux, macOS and Windows (amd64/arm64) are
on the [releases page](https://github.com/janalis/custos/releases). See
[Installation](https://janalis.github.io/custos/guide/installation) for
details and building from source.

## Quick start

```sh
custos analyse                            # report problems in the project
custos fix --dry-run --diff src           # preview quick-fixes
custos fix src                            # apply them
custos explain OneTimeUseVariables        # what a rule does, its options
custos mcp --root /path/to/project        # MCP server for AI agents
```

```text
src/Invoice.php:3:5: warning: Variable $total is used only once; inline its value. [OneTimeUseVariables] (fixable)
src/Invoice.php:7:6: error: Restrict the classes unserialize() may create via its second argument. [UnserializeExploits]
```

More in the guide: [command line](https://janalis.github.io/custos/guide/cli),
[baseline for legacy code](https://janalis.github.io/custos/guide/suppressing#baseline),
[CI recipes](https://janalis.github.io/custos/guide/ci),
[editor setup](https://janalis.github.io/custos/guide/editors),
[MCP setup](https://janalis.github.io/custos/guide/mcp) and the
[rule reference](https://janalis.github.io/custos/rules/).

## Configuration

An optional `custos.json` at the project root pins the PHP version, chooses
paths and turns rules on or off; see
[Getting started](https://janalis.github.io/custos/guide/getting-started#_3-configure)
for an example and
[Configuration](https://janalis.github.io/custos/guide/configuration) for
every key.

## Contributing

Contributions are welcome: bug reports with a PHP snippet, false positives,
fixes. Read [CONTRIBUTING.md](CONTRIBUTING.md) and the
[contributor guide](https://janalis.github.io/custos/contributing/). Please
note the **clean-room rule**: no code or text from the upstream plugin may
enter this repository.

```sh
make build          # → bin/custos
make verify         # architecture + lint + tests + fixtures + 100 % coverage + clean-room scan
```

## License

[MIT](LICENSE). Builtin symbol data from
[JetBrains phpstorm-stubs](https://github.com/JetBrains/phpstorm-stubs)
(Apache-2.0). See [NOTICE](NOTICE).
