<?php
$p=proc_open("cat",[],$pipes);$first=proc_get_status($p);if(!$first["running"]){$exit=$first["exitcode"];}proc_get_status($p)["running"];function dead(){return;proc_get_status($p)["exitcode"];}

strlen("unrelated");

$p=proc_open("cat",[],$pipes);$a=proc_get_status($p);if($a["running"]){proc_get_status($p)["exitcode"];}if(!$a["unknown"]){proc_get_status($p)["exitcode"];}if(!$unrelated){proc_get_status($p)["exitcode"];}if(!is_resource($p)){proc_get_status($p)["exitcode"];}echo $unknown["exitcode"];

$p=proc_open('cat',[],$pipes);$first=proc_get_status($p);if(!$first['running']){$alias=&$p;unknown($alias);proc_get_status($p)['exitcode'];}
