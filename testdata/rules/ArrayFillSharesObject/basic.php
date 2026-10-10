<?php
$a=<warning descr="Construct a separate object for each array entry.">array_fill(0,3,new stdClass())</warning>; $a[0]->id=1; echo $a[1]->id;
