# UnknownInspection implementation notes

## Known inspection set

- custos rule IDs and legacy EA names (the `legacyId` entries of
  `internal/inspection/meta/rules.json`) are both resolved by `meta.Lookup`.
- The PhpStorm short names live in `phpstorm_inspections.txt`, embedded at
  build time (one name per line, `#` starts a comment). Extend that file to
  grow the list.

## Known gaps

The curated PhpStorm list cannot be extended by users. A configuration
option for extra names would remove false reports for third-party plugin
inspections.
