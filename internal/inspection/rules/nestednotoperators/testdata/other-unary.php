<?php
// Only `!` and parentheses extend a chain (spec Divergences).
$a = !(int)!$v;
$b = !-!$v;
$c = !@!$v;
