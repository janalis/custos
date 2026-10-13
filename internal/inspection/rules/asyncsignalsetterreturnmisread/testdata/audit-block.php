<?php
function f(){if (<warning descr="Query the signal state after changing it.">!pcntl_async_signals(true)</warning>) {echo "disabled";}}
