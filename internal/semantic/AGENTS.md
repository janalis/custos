# Semantic ownership

Names, types, symbols, inference and embedded stubs live here. Production
code may import this root and `internal/php`; never import inspections or
adapters. Keep inference environment caches and symbol-index synchronization
within their owning packages.

Inference is split by environment, expressions, operators, members, calls,
variables, conditions, guards and invalidation. Extend the appropriate file
with focused coverage and hot-path benchmarks. Move embedded assets together
with their package and update generator defaults when paths change.
