<?php
namespace Billing;

use function gmmktime;

function epoch_values($flag)
{
    $now   = time();
    $utc   = time();
    $root  = time();
    $dst   = mktime(12, 0, 0, 6, 1, 2010, $flag);
    $dst2  = gmmktime(1, 2, 3, 4, 5, 2001, -1);
}
