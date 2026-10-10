<?php
$p=proc_open("worker",[1=>["pipe","w"],2=>["pipe","w"]],$pipes);stream_set_blocking($pipes[1],false);$stdout=stream_get_contents($pipes[1]);$stderr=stream_get_contents($pipes[2]);stream_get_contents($unknown);stream_get_contents($pipes[1]);strlen("pipes");
