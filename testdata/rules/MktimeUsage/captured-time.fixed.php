<?php
namespace Clock {
    function time() { return 0; }

    $a = \time();
}

namespace Imported {
    use function Clock\time;

    $b = \time();
}

namespace Plain {
    $c = time();
    $d = time();
}
