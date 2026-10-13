<?php
$c=new Collator("en_US"); $a=["part-q"=>"Quartz","part-a"=>"Amber"]; <warning descr="Preserve keys during collation sorting.">$c->sort($a)</warning>;
