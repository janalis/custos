<?php
for ($i = 0; $i < 3; $i++) {}
foreach ($items as $k => $v) {
    foreach ($v as $x) {}
}
$fn = function ($p) use ($q) {
    foreach ($p as $q) {}
};
foreach ($a as $i) {
    while (true) { $i++; break; }
}
for (;;) { break; }
foreach ($a as $$name) {}
foreach ($a as [$m, $m]) {}
for ($x = 0; $c = $x * 2, ($z = 0), $x++ < 75;) {}
for ($j = 0; $j++, next($a), $it->next(), Cursor::step(), $j < 9;) {}
