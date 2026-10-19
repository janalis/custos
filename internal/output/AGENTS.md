# Output adapters

`report` renders positioned diagnostic items (text, JSON, Checkstyle,
GitHub, SARIF); `diff` renders source changes. Baselines and core
analysis do not import output. Preserve all format bytes, rune positions,
escaping and deterministic order. SARIF may consume independent inspection
metadata for rule descriptions. Test output/error paths in the adapter.
