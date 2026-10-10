<?php
$a=new stdClass(); $a->child=new stdClass(); $b=clone $a; <warning descr="Clone the nested object before mutating independent state.">$b->child->label</warning>='changed';
