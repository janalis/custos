<?php
function walk($r, $f, $m)
{
    while (<weak_warning descr="Put the constant operand on the right side of the comparison.">false !== $r = $r->next()</weak_warning>) {}
    if (<weak_warning descr="Put the constant operand on the right side of the comparison.">null === $x = $f()</weak_warning>) {}
    $y = <weak_warning descr="Put the constant operand on the right side of the comparison.">0 == ($m ?: 1)</weak_warning>;
    $z = <weak_warning descr="Put the constant operand on the right side of the comparison.">'a' === $m . 'b'</weak_warning>;
    return <weak_warning descr="Put the constant operand on the right side of the comparison.">true === $m</weak_warning> ?? $f;
}
