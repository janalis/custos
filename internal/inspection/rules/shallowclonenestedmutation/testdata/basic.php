<?php
$a=new stdClass(); $a->child=new stdClass(); $b=clone $a; <warning descr="Clone the nested object before mutating independent state.">$b->child->label</warning>='changed';

final class ReviewChild { public $v; }
final class ReviewSealed { public $child; }
class ReviewOpen { public $child; }
function reviewSealed(ReviewSealed $a) { $a->child = new ReviewChild; $b = clone $a; <warning descr="Clone the nested object before mutating independent state.">$b->child->v</warning> = 1; }
function reviewConstructed() { $a = new ReviewOpen; $a->child = new ReviewChild; $b = clone $a; <warning descr="Clone the nested object before mutating independent state.">$b->child->v</warning> = 1; }
