# Editors (LSP)

`custos lsp` is a language server that speaks LSP over stdio. Run it
**alongside** your main PHP language server (Intelephense, Phpactor…): custos
only adds diagnostics and code actions.

What it offers:

- diagnostics as you type, for the open files;
- quick-fixes (`quickfix` code actions), resolved lazily when the client
  supports it;
- *Suppress … for this statement*: inserts `// @custos-ignore <Rule>`
  above the statement. It is only offered when it silences exactly that
  finding;
- `source.fixAll.custos`: apply every quick-fix in the file, for example on
  save;
- the commands `custos.fixFile` (argument: the document URI) and
  `custos.fixRule` (arguments: the document URI and a rule ID), which apply
  the quick-fixes of an open document.

Diagnostics follow your typing with a short delay (150 ms). Once you pause
for a second, the server updates the cross-file flow summaries of the edited
files in the background and, when they changed, re-checks the other open
files. Saving a file updates the project index and re-checks the other open
files. When the client supports dynamic file watchers, the server also
watches `**/*.php`, so changes made outside the editor (a `git checkout`,
generated code) reach the index, and the [baseline](./suppressing#baseline)
file, so regenerating it takes effect at once. The server starts from the
declarations the command line [keeps between runs](./cli#paths) and parses
only the files that changed since. With `flow.order` set, it also summarizes
the cross-file flow in the order the last run saved, and saves its own.

## Settings

The server uses the first workspace folder and finds the
[project root](./configuration#project-root) from it, like the CLI does from
its start directory. It reads `custos.json` there and applies the project
baseline, so the editor shows the same findings as `custos analyse`.

Editors may override settings with `initializationOptions`, or later with
the `settings` of `workspace/didChangeConfiguration`, which replace them.
Both take the shape of [`custos.json`](./configuration) with these limits:

- `php`, `comparisonStyle`, `shortOpenTag` and `flow` replace the project
  values;
- a `rules` entry overrides the project's entry for that rule field by
  field: `{"severity": "error"}` keeps the project's `enabled` and options;
- `paths`, `exclude` and `baseline` are ignored: what to analyze stays the
  project's choice.

When a document's [analysis status](./cli#analysis-status) changes (complete,
or incomplete with other reasons), the server sends an informational
`window/logMessage` naming the reasons. It adds no diagnostic.

## Neovim (0.11+)

```lua
vim.lsp.config('custos', {
  cmd = { 'custos', 'lsp' },
  filetypes = { 'php' },
  root_markers = { 'custos.json', 'composer.json', '.git' },
})
vim.lsp.enable('custos')
```

To apply all fixes on save:

```lua
vim.api.nvim_create_autocmd('BufWritePre', {
  pattern = '*.php',
  callback = function()
    vim.lsp.buf.code_action({ context = { only = { 'source.fixAll.custos' } }, apply = true })
  end,
})
```

## Helix

In `languages.toml`:

```toml
[language-server.custos]
command = "custos"
args = ["lsp"]

[[language]]
name = "php"
language-servers = ["intelephense", "custos"]   # keep your main server first
```

## VS Code

Use a generic LSP client extension and point it at `custos lsp` for the
`php` language. Or wrap it in a minimal extension built on
`vscode-languageclient`:

```ts
const client = new LanguageClient('custos', 'custos', {
  command: 'custos', args: ['lsp'], transport: TransportKind.stdio,
}, { documentSelector: [{ scheme: 'file', language: 'php' }] });
await client.start();
```

With `TransportKind.stdio`, the client appends `--stdio` to the arguments;
`custos lsp` accepts it.

## PhpStorm and other JetBrains IDEs

Install the [LSP4IJ](https://plugins.jetbrains.com/plugin/23257-lsp4ij)
plugin, then add a language server with the command `custos lsp`, mapped to
PHP files.

## Sublime Text

With the [LSP](https://packagecontrol.io/packages/LSP) package, in
*Preferences › Package Settings › LSP › Settings*:

```json
{
  "clients": {
    "custos": {
      "enabled": true,
      "command": ["custos", "lsp"],
      "selector": "source.php"
    }
  }
}
```

## Emacs (lsp-mode)

`lsp-mode` can run custos as an add-on next to your main PHP server:

```elisp
(with-eval-after-load 'lsp-mode
  (lsp-register-client
   (make-lsp-client :new-connection (lsp-stdio-connection '("custos" "lsp"))
                    :activation-fn (lsp-activate-on "php")
                    :add-on? t
                    :server-id 'custos)))
```

## Composer installs

If custos is installed with Composer, use `vendor/bin/custos` (or its
absolute path) as the command instead of `custos`.
