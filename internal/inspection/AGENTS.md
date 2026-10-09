# Inspection ownership

Implement from clean-room specs only. Each `rules/<lowercase-ID>` package
owns its private implementation, companions and tests, and exposes `New()`.
Never import another rule package. Add constructors explicitly in catalogue
execution order; metadata is independent of implementations.

Keep helpers local until shared. Distinguish written AST names (`astquery`),
local flow (`flowquery`) and resolved symbols/types (`semanticquery`). PHPUnit
call and version support belongs in `phpunit`. Diagnostic contracts live in
`internal/diagnostic`; lazy edits must retain exact byte spans and fix safety.

Run `make architecture`, single-ID fixtures/conformance, and coverage. Run
`make rules-doc` when specs, fixtures or status change.
