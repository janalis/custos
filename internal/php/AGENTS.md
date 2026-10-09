# PHP language ownership

This root owns source parsing, AST/traversal, source positions, PHP versions
and PHPDoc syntax. Production imports stay within this root. Resolution,
inspection decisions and filesystem/project coordination belong elsewhere.

Preserve arena ownership, recovery behavior, byte spans and PHP 5.3–8.5
binding rules. Regenerate node kinds with `go generate ./internal/php/syntax`.
Parser/lexer hot-path changes need benchmarks with allocation measurements.
