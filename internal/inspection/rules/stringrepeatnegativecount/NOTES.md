# StringRepeatNegativeCount implementation notes

## Approach

Two proofs, tried in order:

1. `semanticquery.NativeIntRange` gives the interval of the `times`
   argument; the call is reported when its upper bound is negative. Loop
   heads widen changing bounds after their second visit.
2. Otherwise `semanticquery.NativeIntEntry` looks for a concrete witness: it
   walks the acyclic control-flow prefix from the scope entry and reports
   when the first value read is negative, even if the variable is rewritten
   later.
