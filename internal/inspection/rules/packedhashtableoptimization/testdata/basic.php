<?php
$grid = <weak_warning descr="Sort the integer keys ascending so the array can be stored packed.">[</weak_warning>
    5 => 'e', 3 => 'c', 9 => 'i'
];
$cells = <weak_warning descr="Sort the integer keys ascending so the array can be stored packed.">array</weak_warning>(
    -1 => 'a', -4 => 'b', 0 => 'c', 2 => 'd'
);
$names = <weak_warning descr="Write the keys as integers so the array can be stored packed.">[</weak_warning>
    '3' => 'x', 4 => 'y', '8' => 'z'
];
$mixed = <weak_warning descr="Sort the integer keys ascending so the array can be stored packed.">array</weak_warning>('7' => 1, '2' => 2, 9 => 3);
$neg = <weak_warning descr="Write the keys as integers so the array can be stored packed.">[</weak_warning>'-2' => 1, - 1 => 2, "0" => 3];
