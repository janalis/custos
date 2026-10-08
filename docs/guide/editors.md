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
- the commands `custos.fixFile` and `custos.fixRule`.

Saving a file updates the project index and re-checks the other open files.
Settings come from `custos.json` at the workspace root. Editors may override
them with `initializationOptions`, which take the same JSON shape as
[`custos.json`](./configuration).

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
`vscode-languageclient`, with `command: "custos", args: ["lsp"]`.

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
