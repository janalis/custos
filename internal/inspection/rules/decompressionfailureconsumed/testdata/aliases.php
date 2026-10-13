<?php
use function gzdecode as builtinCall5;
$s=builtinCall5($bytes); <warning descr="Check decompression before consuming data.">strlen($s)</warning>;
