<?php
function paths()
{
    $a = __DIR__;
    $b = __DIR__;
    $c = __DIR__;
    $d = DIRNAME(__file__, 2);
    $e = DirName(__line__);
    return [$a, $b, $c, $d, $e];
}
