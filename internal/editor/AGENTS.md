# Editor adapter

LSP owns transport, settings, document lifecycle and code actions. Reuse
project buffer/index operations and diagnostic byte edits. Preserve UTF-16
positions, document version checks, queued workspace updates and lazy actions.

Publish `Engine.WithIndex` replacement copies while holding the existing
server mutex. Keep analysis snapshots independent of later engine publication.
Validate document/settings/action changes with LSP tests and the race detector.
