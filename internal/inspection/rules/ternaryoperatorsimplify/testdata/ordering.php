<?php
function ordering(int $count, string $name, float $ratio, $any, ?int $max) {
    $r = <weak_warning descr="Replace the ternary with '$count >= 10'.">$count < 10 ? false : true</weak_warning>;
    $r = <weak_warning descr="Replace the ternary with '$name > 'm''.">$name <= 'm' ? false : true</weak_warning>;
    $r = <weak_warning descr="Replace the ternary with '$max < 3'.">$max >= 3 ? false : true</weak_warning>;
    $r = <weak_warning descr="Replace the ternary with '!($ratio < 0.5)'.">$ratio < 0.5 ? false : true</weak_warning>;
    $r = <weak_warning descr="Replace the ternary with '!($any > $count)'.">$any > $count ? false : true</weak_warning>;
    $r = <weak_warning descr="Replace the ternary with '$ratio != 0.5'.">$ratio == 0.5 ? false : true</weak_warning>;
    $r = <weak_warning descr="Replace the ternary with '$ratio < 0.5'.">$ratio < 0.5 ? true : false</weak_warning>;
    return $r;
}
