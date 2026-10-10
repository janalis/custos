<?php
use function stream_select as builtinCall13;
$r=builtinCall13($read,$write,$except,1); if (<warning descr="Handle selection failure before processing ready streams.">$r!==0</warning>) { echo "ready"; }
