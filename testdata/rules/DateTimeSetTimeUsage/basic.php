<?php
class Clock extends DateTime {}
class Tuned extends DateTime { public function setTime($h, $m, $s = 0, $us = 0) { return $this; } }

function rewind_clock(DateTime $d, Clock $c, DateTimeImmutable $i, Tuned $t, $usec, $unknown)
{
    $d->setTime(8, 30, 0, <error descr="Microseconds argument requires PHP 7.1+; on this version the call returns false.">$usec</error>);
    $c->setTime(8, 30, 0, <error descr="Microseconds argument requires PHP 7.1+; on this version the call returns false.">250</error>);
    $d->setTime(8, 30, 0);
    $d->settime(8, 30, 0, <error descr="Microseconds argument requires PHP 7.1+; on this version the call returns false.">1</error>);
    $i->setTime(8, 30, 0, 250);
    $t->setTime(8, 30, 0, 250);
    $unknown->setTime(8, 30, 0, 250);
    date_time_set($d, 8, 30, 0, <error descr="Microseconds argument requires PHP 7.1+; on this version the call returns false.">0</error>);
    \date_time_set($d, 8, 30, 0, <error descr="Microseconds argument requires PHP 7.1+; on this version the call returns false.">null</error>);
    date_time_set($d, 8, 30, 0);
}
