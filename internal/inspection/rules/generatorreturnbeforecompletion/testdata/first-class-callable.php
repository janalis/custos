<?php
// The closure wraps the generator function: no generator was started.
function wrapped()
{
    $g = (function () {
        yield 1;
        return 2;
    })(...);
    $r = $g->getReturn();
    return $r;
}
