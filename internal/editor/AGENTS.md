# Editor adapter

`lsp` owns transport, settings, workspace indexing, document lifecycle,
baseline filtering, diagnostics and code actions. `mcp` owns MCP tools,
transports, authorization, checkpoints, prepared fix plans and confined
writes; it receives the inspection list from the CLI and must not import the
catalogue or report formats (`make architecture`). Reuse project buffer/index operations and
diagnostic byte edits. Preserve UTF-16 positions, document version checks,
queued workspace updates and lazy actions.

Publish `Engine.WithIndex`/`WithFlow` replacement copies while holding the
server mutex (see `docs/contributing/architecture.md`). Keep analysis
snapshots independent of later engine publication. Validate
document/settings/action changes with LSP tests and the race detector, and
MCP changes with its hostile-input and authorization tests.
