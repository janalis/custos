<?php
[$host, $port] = explode(':', $addr);
['w' => $w, 'h' => $h] = $size;
[$a, [$b, $c]] = $nested;

foreach ($pairs as [$left, $right]) {
    swap($left, $right);
}
foreach ($grid as $row => [, $cell]) {
    show($row, $cell);
}
