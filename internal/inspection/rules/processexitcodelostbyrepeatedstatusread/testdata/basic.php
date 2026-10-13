<?php
$p=proc_open("cat",[],$pipes);$first=proc_get_status($p);if(!$first["running"]){$exit=<warning descr="Keep the exit code from the first completed status.">proc_get_status($p)["exitcode"]</warning>;}
