<?php
// @noinspection ConnectTimeoutAssumedToBoundReads
$s=fsockopen("example.test",443,$errno,$errstr,2); if ($s!==false) { fgets($s); }
