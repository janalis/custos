<?php
function operands($x, $a, $b)
{
    return [
        <warning descr="Compare directly: '$x == ($a ?? $b)'.">in_array($x, [$a ?? $b])</warning>,
        <warning descr="Compare directly: '$x == $a + $b'.">in_array($x, [$a + $b])</warning>,
        <warning descr="Compare directly: '$x == ($a ? 1 : 2)'.">in_array($x, [$a ? 1 : 2])</warning>,
        !!<warning descr="Compare directly: '$x != $a'.">!in_array($x, [$a])</warning>,
        (int) <warning descr="Compare directly: '$x == $a'.">in_array($x, [$a])</warning>,
    ];
}
