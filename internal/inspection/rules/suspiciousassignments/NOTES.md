# SuspiciousAssignments implementation notes

## Approach

- Check C (parameter overwritten before use) uses the shared flow proof
  that no reachable read retains the parameter entry value. `saFlowVars` scans the body in source order, visiting an
  assignment's right-hand side before its target and skipping nested
  closures (their text still counts for D10). One walk collects the first
  two accesses of every parameter; a walk per parameter was quadratic.
- D16 skips conditional blocks that always terminate using
  `syntax.Terminates`.
