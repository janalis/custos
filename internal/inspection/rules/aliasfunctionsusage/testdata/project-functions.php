<?php
function close($handle) { return $handle->release(); }
function recode($text) { return strrev($text); }
if (!function_exists('sizeof')) {
    function sizeof($items) { return 0; }
}

close($connection);
$r = \recode($label);
$n = <warning descr="Use 'count(...)' instead of the alias 'sizeof(...)'.">sizeof</warning>($rows);
$d = <warning descr="Use 'is_float(...)' instead of the alias 'is_double(...)'.">is_double</warning>($ratio);
