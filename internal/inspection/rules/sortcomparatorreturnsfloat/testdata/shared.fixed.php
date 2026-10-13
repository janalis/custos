<?php
function priceDiff(float $a, float $b): float { return $a - $b; }
usort($prices, 'priceDiff');
echo priceDiff(9.0, 1.25);
$delta = fn(float $a, float $b) => $a - $b;
uasort($prices, $delta);
echo $delta(9.0, 1.25);
uksort($prices, (function (float $a, float $b) { return $a <=> $b; }));
usort(callback: static fn(float $a, float $b) => ($a <=> $b), array: $prices);
