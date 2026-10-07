<?php
namespace Clock;

function mktime() { return 0; }

function stamps()
{
    $own  = mktime();
    $real = <warning descr="Call time() instead; mktime()/gmmktime() without arguments is deprecated.">\MkTime()</warning>;
    $utc  = <warning descr="Call time() instead; mktime()/gmmktime() without arguments is deprecated.">GMMKTIME()</warning>;
}
