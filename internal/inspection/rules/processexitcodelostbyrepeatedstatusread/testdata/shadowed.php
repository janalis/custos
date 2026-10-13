<?php namespace Independent;
function proc_get_status(){}

$p=proc_open("cat",[],$pipes);$first=proc_get_status($p);if(!$first["running"]){$exit=proc_get_status($p)["exitcode"];}
