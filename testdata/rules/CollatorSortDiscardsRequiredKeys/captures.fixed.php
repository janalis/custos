<?php
// By-value captures and captures of other variables leave the proof intact.
function byValue(Collator $c): array {
    $labels = ["sku-q" => "Quartz", "sku-a" => "Amber"];
    $peek = function () use ($labels, &$count) { $count = count($labels); };
    $peek();
    $c->asort($labels);
    return $labels;
}
// A capture in another function's scope names a different variable.
function otherScope(Collator $c): array {
    $labels = ["sku-q" => "Quartz", "sku-a" => "Amber"];
    $make = function () { $labels = []; return function () use (&$labels) { $labels[] = 1; }; };
    $lazy = fn () => function () use (&$labels) { $labels = []; };
    $copy = $labels;
    $c->asort($copy);
    return $copy;
}
