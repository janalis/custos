# PHP language ownership

`syntax` (lexer, parser, arena AST, traversal, positions), `version`
(PHP 5.3–8.5) and `phpdoc` (doc tags and type expressions). Production
imports stay within this root. Resolution, inspection decisions and
filesystem/project coordination belong elsewhere.

Preserve arena ownership, recovery behavior, byte spans and PHP 5.3–8.5
binding rules. Regenerate node kinds with `go generate ./internal/php/syntax`
after editing `syntax/ast.go`. Parser/lexer hot-path changes need benchmarks
with allocation measurements.
