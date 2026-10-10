<?php
$s = new SplObjectStorage(); $o = new stdClass(); $s[$o] = ["count" => 1]; $v = $s[$o]; $v["count"]++; $s[$o] = $v;
