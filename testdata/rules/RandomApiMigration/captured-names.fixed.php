<?php
namespace Dice {
    function random_int($min, $max) { return $min; }

    $a = \random_int(1, 6);
    $b = \random_int(1, 6);
}

namespace Seeded {
    use function Dice\mt_srand;

    \mt_srand();
    $c = random_int(1, 6);
}
