<?php
function rebound(SplMinHeap $heap, bool $fresh): void {
    $copy = clone $heap;
    if (!$fresh) { $copy = $heap; }
    foreach (<warning descr="Iterate a clone to preserve the heap.">$copy</warning> as $v) { echo $v; }
}
function branch(SplMinHeap $heap, bool $fresh): void {
    $copy = $heap;
    if ($fresh) { $copy = clone $heap; }
    foreach (<warning descr="Iterate a clone to preserve the heap.">$copy</warning> as $v) { echo $v; }
}
