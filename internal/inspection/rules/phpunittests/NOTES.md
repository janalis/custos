# PhpUnitTests implementation notes

## Known gaps

- `@depends clone x` / `@depends shallowClone x`: the modifier word is
  resolved as the reference, so these tags are reported. Skipping the
  modifier and resolving the following word would remove the false reports.
- PHPUnit 10+ attributes (`#[DataProvider]`, `#[Depends]`, `#[CoversClass]`,
  `#[Test]`) are not read; supporting them would need a separate pass over
  attribute arguments.
