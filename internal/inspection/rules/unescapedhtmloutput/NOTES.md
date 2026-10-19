# UnescapedHtmlOutput implementation notes

## Request superglobal rewrites

`requestRewritten` (taint.go) implements E6. It reads the writes of the
output's scope in source order:

- A write to another literal key than the ones read, or one that stores a
  value still holding request text, replaces nothing the output depends on.
- A reference bound to the superglobal replaces nothing by itself; writes to
  the bound name are then read as writes to the superglobal.
- A write that stores request text into the whole table, into an element a
  replacement emptied, or into an element no literal names (computed or
  nested key, reference to the table, `$GLOBALS` entry) fills it again.
- So does code the scope does not show: a call or construction not known to
  spare the request tables (`flow.Env.Spares`) may store request text in any
  of them.
- In a loop around the output, what fills a table after the output does so
  before the next turn.

## Budgets and limits

- Constant-condition pruning only affects flow successors; lexical
  reachability queries keep their existing contract.
- HTML flag resolution (`flow.Env.htmlFlagBits`) follows bitwise-OR operands
  up to depth 16, then records a `flow-budget` gap and treats the flags as
  unknown.
- `flow.Env.ReferenceResult` gives up on calls with more than 256 arguments
  or callees with more than 64 parameters and records an `argument-budget`
  gap.
- Project-function global propagation follows longer chains within the shared
  analysis budget.
