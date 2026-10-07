<?php
function clauses(array $rows, $a, $b)
{
    if ($a) foreach ($rows as &<warning descr="Unset '$row' right after the loop: it is still a reference to the last element.">$row</warning>) { $row++; } elseif ($b) { echo 1; } else { echo 2; }
    if ($a) foreach ($rows as $item) { echo $item; }
}

foreach ([1, 2] as $top) {
    echo $top;
}
