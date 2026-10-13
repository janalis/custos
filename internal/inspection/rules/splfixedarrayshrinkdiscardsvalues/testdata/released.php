<?php
$kept = new SplFixedArray(4);
$kept[3] = 'tail';
$kept->offsetUnset(3);
$kept[3] = 'again';
<warning descr="Preserve populated entries before shrinking.">$kept->setSize(2)</warning>;

$filled = new SplFixedArray(4);
$filled->offsetSet(3, 'tail');
echo count($filled), $filled->getSize();
<warning descr="Preserve populated entries before shrinking.">$filled->setSize(2)</warning>;
