<?php
class Token {
    /** @var string */
    public $raw = '';
}
/**
 * @param string $buf
 * @param int $at
 * @param Token $t
 */
function peek($buf, $at, $t)
{
    $a = <warning descr="Use '$buf[$at]' (string offset access) instead.">substr($buf, $at, 1)</warning>;
    $b = <warning descr="Use '$buf[strlen($buf) - 1]' (string offset access) instead.">\substr($buf, -1, 1)</warning>;
    $c = <warning descr="Use '$t->raw[0]' (string offset access) instead.">substr($t->raw, 0, 1)</warning>;
    return [$a, $b, $c];
}
