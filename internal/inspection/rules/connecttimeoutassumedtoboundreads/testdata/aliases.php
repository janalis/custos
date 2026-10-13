<?php
use function fsockopen as builtinCall14;
$s=builtinCall14("example.test",443,$errno,$errstr,2); if ($s!==false) { <warning descr="Set a separate timeout for stream reads.">fgets($s)</warning>; }
