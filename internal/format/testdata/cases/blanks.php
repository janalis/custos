<?php // presets: per symfony; set: blank_line_before=break|continue|declare|do|for|foreach|if|return|switch|throw|try|while|yield, max_blank_lines=2
function f()
{
    $a = 1;
    if ($a) {}
    $b = 2;
    for (;;) {}
    $c = 3;
    foreach ($x as $y) {}
    $d = 4;
    while (1) {}
    $e = 5;
    do {} while (0);
    $f = 6;
    switch (1) {}
    $g = 7;
    try {} finally {}
    $h = 8;
    declare(ticks=1);
    $i = 9;
    throw $e;



    yield 1;
    yield 2;
    $j = yield 3;
    while (1) { $k = 1; break; }
    while (1) { $k = 1; continue; }
}
