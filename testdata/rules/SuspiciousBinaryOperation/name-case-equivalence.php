<?php
function sameOperands($a) {
    if (<error descr="Both operands are the same.">strlen($a) === STRLEN($a)</error>) {}
    if (<error descr="Both operands are the same.">$a->Size() > $a->size()</error>) {}
    if (<error descr="Both operands are the same.">Box::Make() == box::make()</error>) {}
    if ($a->size === $a->Size) {}
}
