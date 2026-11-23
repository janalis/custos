# custos for Claude Code

This plugin connects Claude Code to [custos](https://github.com/janalis/custos),
a PHP inspector, fixer and formatter. It needs the `custos` binary on `PATH`
(see the [installation guide](https://janalis.github.io/custos/guide/installation)),
and `jq` for its hooks.

## What it runs

- **MCP server**: `custos mcp` over stdio, started from the directory Claude
  Code starts in. Its tools analyse PHP code, list findings, explain rules,
  look up symbols and types, and preview or apply quick-fixes and formatting.
  They read the project's files and change them only when Claude applies
  fixes or formatting. Like the command line, custos keeps a cache and the
  MCP checkpoints under `custos` in your user cache directory.
- **Language server**: `custos lsp`, for diagnostics on `.php` files. Claude
  Code runs one language server per file extension, so it does not start while
  another plugin, such as `php-lsp`, handles `.php`.
- **PostToolUse hook** (`hooks/analyse-edited-file`): after each `Edit`,
  `MultiEdit` or `Write` of a `.php` file, it runs `custos analyse` on that
  file and passes the findings to Claude: all of them on the first edit of the
  file in a session, then only the ones the edit added, at most 40 at a time.
  It never blocks an edit.
- **SessionEnd hook** (`hooks/forget-session`): removes what the edit hook
  remembered for the session. That state is a list of finding messages per
  file, kept in a directory of your own under `TMPDIR`; state of sessions that
  did not end cleanly is removed after 7 days.

All of them go through `scripts/custos`, which runs `custos` from `PATH`.

## Composer installations

With `CUSTOS_PROJECT_LAUNCHER=1` in Claude Code's environment, `scripts/custos`
runs the project's `vendor/bin/custos` instead. That runs code from the
repository as soon as Claude Code starts in it, so turn it on only for
repositories you trust. On its first run, that Composer launcher downloads the
custos release binary from GitHub (`github.com` and its release asset hosts)
and checks it against the release's SHA-256 sums. Nothing else in the plugin
uses the network, and no data is sent anywhere.

See the [MCP guide](https://janalis.github.io/custos/guide/mcp#claude-code) for
details.
