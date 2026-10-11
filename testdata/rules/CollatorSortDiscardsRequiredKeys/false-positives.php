<?php
// A closure capturing the array by reference can replace it before the sort.
function viaCapture(Collator $c): array {
    $labels = ["sku-q" => "Quartz", "sku-a" => "Amber"];
    $reload = function () use (&$labels) { $labels = ["Zinc", "Agate"]; };
    $reload();
    $c->sort($labels);
    return $labels;
}
// The sorted array is a copy of a local that such a closure may have replaced.
function viaCopy(Collator $c): array {
    $labels = ["sku-q" => "Quartz", "sku-a" => "Amber"];
    $reload = function () use (&$labels, $c) { $labels = ["Zinc", "Agate"]; };
    $reload();
    $copy = $labels;
    $copy = $copy;
    $c->sort(($copy));
    return $copy;
}
// A catch clause rebinds the variable to the caught exception.
function viaCatch(Collator $c) {
    $labels = ["sku-q" => "Quartz", "sku-a" => "Amber"];
    try { risky(); } catch (RuntimeException $labels) { } catch (LogicException) { }
    $c->sort($labels);
}
// Inside a closure, a by-reference capture is shared with the defining scope.
$inner = function (Collator $c) use (&$shared) {
    $shared = ["sku-q" => "Quartz", "sku-a" => "Amber"];
    notify();
    $c->sort($shared);
};
// Top-level code follows the same rule.
$c = new Collator("en_US");
$stock = ["sku-q" => "Quartz", "sku-a" => "Amber"];
$refill = function () use (&$stock) { $stock = ["Zinc", "Agate"]; };
$refill();
$c->sort($stock);
