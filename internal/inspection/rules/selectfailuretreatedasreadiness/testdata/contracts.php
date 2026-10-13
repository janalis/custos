<?php
$r=stream_select($read,$write,$except,1); if (<warning descr="Handle selection failure before processing ready streams.">$r!==0</warning>) { echo "ready"; }
