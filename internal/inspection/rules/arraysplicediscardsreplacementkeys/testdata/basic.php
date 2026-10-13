<?php
$a=[]; <warning descr="Preserve replacement keys before reading the inserted key.">array_splice($a,0,0,["name"=>"Ada"])</warning>; echo $a["name"];
