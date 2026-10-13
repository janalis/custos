<?php
$p=<warning descr="Preserve numeric keys before reading the original key.">array_slice([100=>"a",200=>"b"],0,1)</warning>; echo $p[100];
