<?php
$p=proc_open("worker",[1=>["file","/tmp/log","w"]],$pipes);proc_close($p);$p=proc_open("worker",[1=>["pipe","w"]],$pipes);fclose($pipes[1]);proc_close($p);proc_close($unknown);strlen("process");
