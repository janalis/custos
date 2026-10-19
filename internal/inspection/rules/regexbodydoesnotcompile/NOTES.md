# RegexBodyDoesNotCompile implementation notes

## Approach

The body is scanned by hand (`malformed`) for a small PCRE subset: ordinary
character classes, groups and closing parentheses. Do not validate with Go's
`regexp` package: its RE2 syntax rejects valid PCRE features (look-around,
backreferences, possessive quantifiers), which would turn into false
positives.
