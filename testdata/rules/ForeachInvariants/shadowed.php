<?php
namespace Paging;

function count($a) { return 3; }
function each(array &$a) { return false; }

function show(array $items) {
    for ($i = 0; $i < count($items); $i++) {
        echo $items[$i];
    }
    while (list($k, $v) = each($items)) {
        echo $v;
    }
}
