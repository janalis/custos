<?php
$r=stream_select($read,$write,$except,1); if ($r===false) { throw new RuntimeException(); } if ($r>0) { echo "ready"; }
