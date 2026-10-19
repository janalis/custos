# DecompressionFailureConsumed implementation notes

## Joined producer retention

Producers reaching a use come from the shared flow analysis
(`semanticquery.NativeLocalBuiltinResults` over `Flow().Reaching`), which
keeps one origin per joined assignment:

- Ordinary origin sets retain up to `flow.MaxOrigins` (64) expressions.
  Past that, finite resolved native contracts keep real representatives for
  each false/null guard state, so an eligible unguarded producer still counts
  when other paths use another API or are guarded.
- Argument-sensitive contracts keep their individual producers, so an
  irrelevant variant cannot hide a later eligible call.
- Unknown non-call paths keep no exact identity.
- Origin storage is charged to the scope's retained-state budget. Exhausting
  it leaves the scope unknown (`ProducerBudgetExceeded`) instead of
  fabricating a producer or a guard, and the rule reports nothing.
