<?php
$s=fsockopen("example.test",443,$errno,$errstr,2); if ($s!==false && stream_set_timeout($s,2)) { fgets($s); }
