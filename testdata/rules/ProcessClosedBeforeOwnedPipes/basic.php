<?php
$p=proc_open("worker",[1=>["pipe","w"]],$pipes);<warning descr="Close owned process pipes before waiting for process exit.">proc_close($p)</warning>;
