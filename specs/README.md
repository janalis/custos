# Rule specs

One clean-room behavioral spec per rule (`<RuleID>.md`, from `_TEMPLATE.md`),
written in our own words by the `spec-rule` skill. Implementers work from these
specs only. Specs must never contain EA source code, EA messages or EA fixture
snippets.

Specs describe observable behavior only. Implementation details (non-obvious
algorithms, budgets and limits, shared helpers, known gaps) go in the rule's
optional `internal/inspection/rules/<lowercase-ID>/NOTES.md`. See
[Adding a rule](../docs/contributing/adding-a-rule.md) and the
[clean-room process](../docs/contributing/clean-room.md).
