<?php
function priceDiff(float $a, float $b): float { return $a - $b; }
<warning descr="Compare float values without subtracting them.">usort($prices, 'priceDiff')</warning>;
echo priceDiff(9.0, 1.25);
$delta = fn(float $a, float $b) => $a - $b;
<warning descr="Compare float values without subtracting them.">uasort($prices, $delta)</warning>;
echo $delta(9.0, 1.25);
<warning descr="Compare float values without subtracting them.">uksort($prices, (function (float $a, float $b) { return $a - $b; }))</warning>;
<warning descr="Compare float values without subtracting them.">usort(callback: static fn(float $a, float $b) => ($a - $b), array: $prices)</warning>;
