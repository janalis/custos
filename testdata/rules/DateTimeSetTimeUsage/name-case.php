<?php
class Alarm extends DateTime {}

function ring(Alarm $a, DateTime $d)
{
    $a->SETTIME(6, 0, 0, <error descr="Microseconds argument requires PHP 7.1+; on this version the call returns false.">10</error>);
    $d->SetTime(6, 0, 0);
    Date_Time_Set($d, 6, 0, 0, <error descr="Microseconds argument requires PHP 7.1+; on this version the call returns false.">10</error>);
    \DATE_TIME_SET($d, 6, 0, 0, <error descr="Microseconds argument requires PHP 7.1+; on this version the call returns false.">10</error>);
}
