<?php
$proto = load();
$proto = clone $proto;
echo $proto->name;

$count = 3;
echo $count;
$count = 4;
throw $count;
