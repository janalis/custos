# SuspiciousAssignments implementation notes

## Approach

- Check C (parameter overwritten before use) does not build a control-flow
  graph. `saFlowVars` scans the body in source order, visiting an
  assignment's right-hand side before its target and skipping nested
  closures (their text still counts for D10). One walk collects the first
  two accesses of every parameter; a walk per parameter was quadratic.
- D16 skips conditional blocks that always terminate using
  `syntax.Terminates`.

## Known gaps

A stub return union that carries other scalars next to `array`
(`array|int|float|false`) is still reported by D20a. Silencing those calls
needs argument-aware return types in the type layer, not a change to this
rule.
