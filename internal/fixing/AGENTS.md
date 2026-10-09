# Fix application

Depend on diagnostic edits and a small parsed-file Analyzer contract.
Preserve stable edit ordering, insertion conflicts, multi-edit selection,
ten-iteration limits and single-pass behavior. Rule implementations own fix
safety; project operations own preparation; adapters own filesystem writes.
