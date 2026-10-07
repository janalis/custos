<?php
function rewind_clock(DateTime $d, $usec)
{
    $d->setTime(8, 30, 0, $usec);
    date_time_set($d, 8, 30, 0, 0);
}
