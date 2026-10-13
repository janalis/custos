<?php
$s=fsockopen("example.test",443,$errno,$errstr,2); if ($s!==false) { <warning descr="Set a separate timeout for stream reads.">fgets($s)</warning>; }
