<?php
// presets: per; set: single_import_per_statement=false, import_order=alpha, yoda=yoda, line_length=50
use B\Two, A\One;
use C\Three;
use D\Same;
use d\Same as S2;

echo 0;
use E\Four /* why */;

foreach ($rows as [$a, [$b, $c]]) {}
foreach ($rows as list($a, $b)) {}
#[
    First(1, 2, 3),
    Second(4, 5, 6),
    Third(7, 8, 9),
    Fourth
]
function f() {}
if (
    1 == ($a ? $b : $c)
    && $a . $b == 'x'
    && $e instanceof F == true
) {}
$v = new Foo();
call(fn($x) => $x, function () {
    return 1;
}, $last);
