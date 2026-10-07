<?php
function capture(callable $fn) {
    ob_start();
    $fn();
    $out = <warning descr="Use ob_get_clean() instead of ob_get_contents() + ob_end_clean().">OB_GET_CONTENTS()</warning>;
    Ob_End_Clean();
    return $out;
}
