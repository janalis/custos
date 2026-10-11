<?php
namespace Dice {
    function mt_rand($min, $max) { return $min; }

    $a = \mt_rand(1, 6);
    $b = \mt_rand(1, 6);
}

namespace Seeded {
    use function Dice\mt_srand;

    \mt_srand();
    $c = mt_rand(1, 6);
}
