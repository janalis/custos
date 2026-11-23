<?php
$g = (function() { yield 7; yield 11; })(); $g->rewind();

// A first-class callable of next() advances nothing.
function closureOfNext(): void
{
    $gen = (static function () { yield 1; yield 2; })();
    $advance = $gen->next(...);
    $gen->rewind();
}
