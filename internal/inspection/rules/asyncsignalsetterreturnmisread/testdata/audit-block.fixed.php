<?php
function f(){pcntl_async_signals(true);
if (!pcntl_async_signals()) {echo "disabled";}}
