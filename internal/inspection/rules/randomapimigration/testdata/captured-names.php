<?php
namespace Dice {
    function mt_rand($min, $max) { return $min; }

    $a = <warning descr="Prefer mt_rand() over rand().">rand(1, 6)</warning>;
    $b = <warning descr="Prefer mt_rand() over rand().">\rand(1, 6)</warning>;
}

namespace Seeded {
    use function Dice\mt_srand;

    <warning descr="Prefer mt_srand() over srand().">srand()</warning>;
    $c = <warning descr="Prefer mt_rand() over rand().">rand(1, 6)</warning>;
}
