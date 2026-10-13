<?php
// Mixing list() and [] patterns does not compile; still a pattern.
list([2 => $c, 1 => $b, 0 => $a]) = $mixed;
