<?php
namespace Dice {
    // user helper whose arguments are (max, min)
    function mt_rand($max, $min) { return \mt_rand($min, $max); }

    function roll()
    {
        return [
            mt_rand(6, 1),
            Legacy\rand(6, 1),
            <error descr="Minimum is greater than maximum in this random range.">\MT_RAND(6, 1)</error>,
            <error descr="Minimum is greater than maximum in this random range.">Random_Int(6, 1)</error>,
        ];
    }
}

namespace Cards {
    use function Dice\mt_rand;

    function draw()
    {
        return [mt_rand(52, 1), <error descr="Minimum is greater than maximum in this random range.">RAND(52, 1)</error>];
    }
}
