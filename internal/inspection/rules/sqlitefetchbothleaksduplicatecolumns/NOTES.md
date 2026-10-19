# SqliteFetchBothLeaksDuplicateColumns implementation notes

## Budgets and limits

The consumer scan over the producer's scope visits at most 256 nodes. Past
that budget the JSON-only consumer proof is dropped, so no finding is emitted.
