# Test support

`conformance` runs the rules against their own fixtures (`make fixtures`) and,
from a local upstream checkout, against upstream cases (`make conformance`):
findings are matched on severity and range, fixes are applied in one pass and
compared ignoring whitespace. Never store or print upstream messages or
expected fixed source (see `docs/contributing/clean-room.md`). Fixture format:
`docs/contributing/fixtures.md`.

`testbudget.Of(d)` scales timing limits with machine load; hard performance
budgets belong in `make bench`.
