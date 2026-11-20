# Formatter

`format` lays out PHP source for `custos format`, LSP formatting, MCP
`custos_format` and the Formatting rule. `settings.go` is the settings
registry (the single source of IDs, kinds, values, descriptions and risky
flags; `tools/formatdoc` renders it), `presets.go` the presets and
resolution layers. One pass over the token stream decides the gap between
every two significant tokens, using the syntax tree for context, and yields
the output plus the changes producing it, so a check reports exactly what a
fix writes. Every output passes the safety check (same tokens apart from
rewrites, no new syntax error) or the file is left as written. Never touch
string, heredoc or inline HTML contents. New behavior needs a golden case
in `testdata/cases` reviewed by hand. `editorconfig` parses `.editorconfig`
files and resolves per-file properties through a caller-supplied reader.
`format` imports only `internal/php`, `editorconfig` only the standard
library, and neither reaches the system (`make architecture`).
