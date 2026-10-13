<?php
function operands($x, $a, $b)
{
    return [
        $x == ($a ?? $b),
        $x == $a + $b,
        $x == ($a ? 1 : 2),
        !!($x != $a),
        (int) ($x == $a),
    ];
}
