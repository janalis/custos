<?php
// @custos-ignore HeapIterationConsumesCollection
$h = new SplMinHeap(); $h->insert(4); foreach ($h as $item) {}
