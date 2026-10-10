<?php
$h = new SplMinHeap(); $h->insert(4); foreach (clone $h as $item) {}
