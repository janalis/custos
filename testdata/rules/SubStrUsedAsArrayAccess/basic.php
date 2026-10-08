<?php
function peek(string $buf, ?string $maybe, string|int $mixed, array $rows, int $at)
{
    $a = <warning descr="Use '($buf[$at] ?? '')' (string offset access) instead.">substr($buf, $at, 1)</warning>;
    $b = <warning descr="Use '($buf[strlen($buf) - 1] ?? '')' (string offset access) instead.">substr($buf, -1, 1)</warning>;
    $c = <warning descr="Use '($maybe[0] ?? '')' (string offset access) instead.">substr($maybe, 0, 1)</warning>;
    $d = <warning descr="Use '$mixed[2]' (string offset access) instead.">substr($mixed, 2, 1)</warning>;
    return [$a, $b, $c, $d];
}
