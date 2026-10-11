<?php

$width = -1;
<error descr="Supply a nonnegative array fill count.">array_fill(0, $width, 'c')</error>;

function unrelated()
{
    global $other;
    $other = 5;
}
