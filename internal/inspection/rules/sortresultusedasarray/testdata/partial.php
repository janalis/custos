<?php
// The closure sorts later; it is not the success flag of a sort.
function sorter(array $items)
{
    $sorted = sort($items, ...);
    foreach ($sorted as $item) {
    }
}
