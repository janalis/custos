<?php
namespace Dice {
    function random_int($min, $max) { return $min; }

    $a = <warning descr="Prefer random_int() over rand().">rand(1, 6)</warning>;
    $b = <warning descr="Prefer random_int() over mt_rand().">\mt_rand(1, 6)</warning>;
}

namespace Seeded {
    use function Dice\mt_srand;

    <warning descr="Prefer mt_srand() over srand().">srand()</warning>;
    $c = <warning descr="Prefer random_int() over rand().">rand(1, 6)</warning>;
}
