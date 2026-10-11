<?php
function f($x) {
    switch ($x) { case 1: if (<warning descr="Query the signal state after changing it.">!pcntl_async_signals(true)</warning>) { echo "E"; } }
    if ($x): if (<warning descr="Query the signal state after changing it.">!pcntl_async_signals(false)</warning>) { echo "F"; } endif;
    lbl: if (<warning descr="Query the signal state after changing it.">!pcntl_async_signals(true)</warning>) { echo "D"; }
}
