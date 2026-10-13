<?php
$ok   = <weak_warning descr="Simplify the stacked negations to '(bool)$token'.">!!$token</weak_warning>;
$off  = <weak_warning descr="Simplify the stacked negations to '!$token'.">!!!$token</weak_warning>;
$six  = <weak_warning descr="Simplify the stacked negations to '(bool)$token'.">!!!!!!$token</weak_warning>;
$both = <weak_warning descr="Simplify the stacked negations to '(bool)($m > $n)'.">!!($m > $n)</weak_warning>;
$neg  = <weak_warning descr="Simplify the stacked negations to '!($p instanceof Item)'.">!!! ($p instanceof Item)</weak_warning>;
$wrap = <weak_warning descr="Simplify the stacked negations to '(bool)count($list)'.">!( !(count($list)) )</weak_warning>;
$deep = <weak_warning descr="Simplify the stacked negations to '!$flag'.">!(!(!(($flag))))</weak_warning>;
$and  = <weak_warning descr="Simplify the stacked negations to '(bool)($a && $b)'.">!! ($a && $b)</weak_warning>;
$tern = <weak_warning descr="Simplify the stacked negations to '(bool)($a ? $b : $c)'.">!!($a ? $b : $c)</weak_warning>;
$asg  = <weak_warning descr="Simplify the stacked negations to '(bool)($x = load())'.">!!($x = load())</weak_warning>;
$par  = (<weak_warning descr="Simplify the stacked negations to '(bool)$token'.">!!$token</weak_warning>);
$neg2 = -<weak_warning descr="Simplify the stacked negations to '(bool)$token'.">!!$token</weak_warning>;
if (<weak_warning descr="Simplify the stacked negations to '(bool)$obj->ready()'.">!!$obj->ready()</weak_warning>) {}
