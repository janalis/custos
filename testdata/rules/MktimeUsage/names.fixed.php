<?php
namespace Clock;

function mktime() { return 0; }

function stamps()
{
    $own  = mktime();
    $real = time();
    $utc  = time();
}
