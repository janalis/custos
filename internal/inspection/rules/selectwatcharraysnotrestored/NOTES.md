# SelectWatchArraysNotRestored implementation notes

## Budgets and limits

The search for the dominating watch-array assignment visits at most 256 nodes
of the selection's variable scope (the file at top level), skipping nested
scopes and stopping at the loop start. Exhausting the budget marks the slot
invalid, so nothing is reported for it.
