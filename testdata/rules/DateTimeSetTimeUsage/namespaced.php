<?php
namespace Shop;

function date_time_set($a, $b, $c, $d, $e) {}

function f(\DateTime $d)
{
    date_time_set($d, 8, 30, 0, 0);
    \date_time_set($d, 8, 30, 0, <error descr="Microseconds argument requires PHP 7.1+; on this version the call returns false.">0</error>);
}
