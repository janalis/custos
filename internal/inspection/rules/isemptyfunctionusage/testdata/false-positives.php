<?php

function f(?\DateTime $when, array $rows, mixed $m, object $o, callable $c, $u)
{
    return [
        empty($rows),
        empty($m),
        empty($o),
        empty($c),
        empty($u),
        empty($rows[0]),
    ];
}
