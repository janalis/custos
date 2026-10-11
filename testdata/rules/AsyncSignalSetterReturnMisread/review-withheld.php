<?php
function f($x) {
    foreach ([1] as $y) if (!<warning descr="Query the signal state after changing it.">pcntl_async_signals(true)</warning>) { echo "B"; }
    if ($x) { echo 1; } else if (!<warning descr="Query the signal state after changing it.">pcntl_async_signals(true)</warning>) { echo "A"; }
    lbl: if (pcntl_async_signals(true) === false) { echo "H"; }
    if (false === pcntl_async_signals(true)) { echo "H"; }
    if (pcntl_async_signals(true) == true) { echo "H"; }
}
