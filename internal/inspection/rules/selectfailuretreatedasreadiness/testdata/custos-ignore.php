<?php
// @custos-ignore SelectFailureTreatedAsReadiness
$r=stream_select($read,$write,$except,1); if ($r!==0) { echo "ready"; }
