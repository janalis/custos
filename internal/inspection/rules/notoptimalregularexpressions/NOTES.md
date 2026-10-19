# NotOptimalRegularExpressions implementation notes

## D22 replacement site

`noreSite` picks the node a D22 boolean test replaces, since preg_match()
yields 1/0 while the rewrite yields true/false:

- logical operand: the call itself, or its direct `!` (inverted);
- direct operand of a comparison with a number literal: the whole
  comparison, inverted when it holds for 0 and not for 1; skipped when it
  holds for both or neither (`== 2`, `< 5`); a number on the left is read
  mirrored (`1 > C` as `C < 1`);
- any other operator operand (arithmetic, `@`, casts, a comparison with a
  non-number or through parentheses): skipped;
- a value position (assignment, return, argument, `echo`): skipped, as
  true/false would leak where 1/0 was expected (`: int` returns under
  `strict_types`, arithmetic).

## Known gaps

- Heredoc and nowdoc patterns are not analysed. A nowdoc, or a heredoc
  without interpolation, could be handled like a quoted literal whose
  contents are its body.
