<?php
function literals($s)
{
    $s = str_replace(['a', 'b'], ['1', '2'], $s);
    $s = <warning descr="Fold this str_replace() into the preceding one on the same variable.">str_replace(['c'], ['3'], $s)</warning>;
    return $s;
}
