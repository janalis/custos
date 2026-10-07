<?php
namespace Billing;

use function gmmktime;

function epoch_values($flag)
{
    $now   = <warning descr="Call time() instead; mktime()/gmmktime() without arguments is deprecated.">mktime()</warning>;
    $utc   = <warning descr="Call time() instead; mktime()/gmmktime() without arguments is deprecated.">gmmktime()</warning>;
    $root  = <warning descr="Call time() instead; mktime()/gmmktime() without arguments is deprecated.">\mktime( )</warning>;
    $dst   = mktime(12, 0, 0, 6, 1, 2010, <warning descr="The is_dst argument is deprecated and was removed in PHP 7.0.">$flag</warning>);
    $dst2  = gmmktime(1, 2, 3, 4, 5, 2001, <warning descr="The is_dst argument is deprecated and was removed in PHP 7.0.">-1</warning>);
}
