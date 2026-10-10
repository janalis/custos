<?php
// @custos-ignore CollatorSortDiscardsRequiredKeys
$c=new Collator("en_US"); $a=["part-q"=>"Quartz","part-a"=>"Amber"]; $c->sort($a);
