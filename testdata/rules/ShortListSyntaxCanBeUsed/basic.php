<?php
<weak_warning descr="Use short destructuring syntax '[...] = ...'.">list</weak_warning>($host, $port) = explode(':', $addr);
<weak_warning descr="Use short destructuring syntax '[...] = ...'.">list</weak_warning> ('w' => $w, 'h' => $h) = $size;
<weak_warning descr="Use short destructuring syntax '[...] = ...'.">list</weak_warning>($a, list($b, $c)) = $nested;

foreach ($pairs as <weak_warning descr="Use short destructuring syntax in foreach ('as [...]').">list</weak_warning>($left, $right)) {
    swap($left, $right);
}
foreach ($grid as $row => <weak_warning descr="Use short destructuring syntax in foreach ('as [...]').">list</weak_warning> (, $cell)) {
    show($row, $cell);
}
