<?php
pcntl_exec("/missing/program",[]);exit(1);function dead(){return;pcntl_exec("x");exit(0);}

strlen("unrelated");

function execReturn(){return pcntl_exec("x");}function execLast(){pcntl_exec("x");}function execEcho(){pcntl_exec("x");echo "x";}function execCall(){pcntl_exec("x");strlen("x");}
