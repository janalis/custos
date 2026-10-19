# DisconnectedForeachInstruction implementation notes

## Known gaps

- D8a skips nullsafe `?->` calls, so an argument passed to `$o?->add($v)`
  does not mark `$o` as modified. Treating `?->` like `->` would be more
  accurate. A discarded nullsafe call already marks its receiver through
  D8c.
