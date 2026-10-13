<?php
function paths()
{
    $a = <warning descr="Replace dirname(__FILE__) with __DIR__.">Dirname(__FILE__)</warning>;
    $b = <warning descr="Replace dirname(__FILE__) with __DIR__.">dirname(__file__)</warning>;
    $c = <warning descr="Replace dirname(__FILE__) with __DIR__.">\DIRNAME(__File__)</warning>;
    $d = DIRNAME(__file__, 2);
    $e = DirName(__line__);
    return [$a, $b, $c, $d, $e];
}
