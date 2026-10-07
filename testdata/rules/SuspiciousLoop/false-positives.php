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
