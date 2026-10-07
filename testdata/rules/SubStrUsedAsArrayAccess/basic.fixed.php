<?php
function peek(string $buf, ?string $maybe, string|int $mixed, array $rows, int $at)
{
    $a = ($buf[$at] ?? '');
    $b = ($buf[strlen($buf) - 1] ?? '');
    $c = ($maybe[0] ?? '');
    $d = ($mixed[2] ?? '');
    return [$a, $b, $c, $d];
}
