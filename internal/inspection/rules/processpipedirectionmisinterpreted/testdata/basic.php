<?php
$p=proc_open(["cat"],[0=>["pipe","w"]],$pipes);if(is_resource($p)){<warning descr="Choose pipe modes from the child perspective.">fwrite($pipes[0],"input")</warning>;}
