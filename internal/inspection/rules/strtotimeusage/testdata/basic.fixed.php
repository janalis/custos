<?php
namespace Billing;

function due($spec) {
    $now   = time();
    $other = \time();
    $next  = strtotime($spec);
    $last  = strtotime('last monday');
    $d = strtotime('now');
    return [$now, $other, $next, $last, $d];
}
