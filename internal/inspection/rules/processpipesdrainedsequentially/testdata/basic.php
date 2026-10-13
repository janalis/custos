<?php
$p=proc_open("worker",[1=>["pipe","w"],2=>["pipe","w"]],$pipes);$stdout=stream_get_contents($pipes[1]);$stderr=<warning descr="Drain process output pipes concurrently.">stream_get_contents($pipes[2])</warning>;
