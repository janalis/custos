<?php
// @custos-ignore ProcessExitCodeLostByRepeatedStatusRead

$p=proc_open("cat",[],$pipes);$first=proc_get_status($p);if(!$first["running"]){$exit=proc_get_status($p)["exitcode"];}
