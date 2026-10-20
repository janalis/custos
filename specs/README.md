# Rule specs

One clean-room behavioral spec per rule (`<RuleID>.md`, from `_TEMPLATE.md`),
written in our own words: by the `spec-rule` skill for EA-modeled rules,
directly from the PHP behavior for native rules. Implementers work from these
specs only. Specs must never contain EA source code, EA messages or EA fixture
snippets.

Specs describe observable behavior only. Implementation details (non-obvious
algorithms, budgets and limits, shared helpers, known gaps) go in the rule's
optional `internal/inspection/rules/<lowercase-ID>/NOTES.md`. Summary,
Options and, for native rules, Detection, Exceptions, Fix and PHP versions
are published on the rule's page and by `custos explain`, so describe
options in words rather than by D/E label. See
[Adding a rule](../docs/contributing/adding-a-rule.md) and the
[clean-room process](../docs/contributing/clean-room.md).
