<?php
function glued($x, $b)
{
    if (<weak_warning descr="Put the constant operand on the left side of the comparison.">$x=="a"</weak_warning>and $b) {}
    return<weak_warning descr="Put the constant operand on the left side of the comparison.">$x==null</weak_warning>;
}

function ordered($i, $o)
{
    $a = <weak_warning descr="Put the constant operand on the left side of the comparison.">$i++ == "$i"</weak_warning>;
    $b = <weak_warning descr="Put the constant operand on the left side of the comparison.">($i = $o->id()) . '' === "$i"</weak_warning>;
    return <weak_warning descr="Put the constant operand on the left side of the comparison.">$o->pos == "{$o->next()}"</weak_warning>;
}
