<?php
$s = new SplObjectStorage(); $o = new stdClass(); $s[$o] = ["count" => 1]; <warning descr="Assign the modified storage value back.">$s[$o]["count"]++</warning>;
