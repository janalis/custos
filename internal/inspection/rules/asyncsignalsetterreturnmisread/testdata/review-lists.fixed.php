<?php
function f($x) {
    switch ($x) { case 1: pcntl_async_signals(true);
if (!pcntl_async_signals()) { echo "E"; } }
    if ($x): pcntl_async_signals(false);
if (!pcntl_async_signals()) { echo "F"; } endif;
    lbl: pcntl_async_signals(true);
if (!pcntl_async_signals()) { echo "D"; }
}
