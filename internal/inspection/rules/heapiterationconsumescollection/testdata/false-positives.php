<?php
function dump(SplMinHeap $heap): int {
    $copy = clone $heap;
    $copy->insert(0);
    echo count($copy), ": ";
    foreach ($copy as $v) { echo $v, " "; }
    return count($heap);
}
$h = new SplMinHeap(); $h->insert(4); foreach (clone $h as $item) {}
