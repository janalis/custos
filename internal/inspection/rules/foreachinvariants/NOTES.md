# ForeachInvariants implementation notes

## Element writes

`foreachInvElementsChanged` decides whether the loop writes container
elements in a way the `foreach` snapshot and the fix's value variable would
not see:

- an element other than `$c[$i]` (`$c[$i + 1] = …`, `$c[$k]['x'] = …`): a
  later iteration would read the new value;
- a path under `$c[$i]` followed in the body by a mention of an overlapping
  path (`$c[$i] = trim($c[$i]); echo $c[$i];`): the fix would read the stale
  `$iValue`.

Taking a reference (`$r = &$c[$i]`) is not a write. With `escapes` set, an
element captured by reference (`$r = &$c[$i]`, `[&$c[$i]]`) counts as
written, since it may be written later without being named. Such a loop is
still reported; only its fix is withheld (D9), as it would read a stale copy.
