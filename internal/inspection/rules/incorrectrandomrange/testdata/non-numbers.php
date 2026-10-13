<?php
function odd($x)
{
    return [
        rand(UNDECLARED_LIMIT, 1),
        rand(true, false),
        rand(-PHP_INT_MAX, 1),
        rand(-$x, 1),
        rand('9', 1),
        rand(PHP_EOL, 1),
        <error descr="Minimum is greater than maximum in this random range.">rand(E_ALL, 1)</error>,
    ];
}
