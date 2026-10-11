<?php
function fixed($i, $p) {
    $level = 10;
    var_dump($level);
    $g = function () use ($level) {};
    <error descr="Supply PNG compression between zero and nine.">imagepng($i, $p, $level)</error>;
}

$top = 10;
<error descr="Supply PNG compression between zero and nine.">imagepng($i, $p, $top)</error>;
