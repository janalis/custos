<?php

function readInUnrelatedLoop(array $rows): void
{
    $it = new ArrayIterator([7, 11]);
    $total = iterator_count($it);
    foreach ($rows as $row) {
        echo $row, <warning descr="Restore the iterator position after counting.">$it->key()</warning>, $total;
    }
}

function advanceOnly(): void
{
    $it = new ArrayIterator([7, 11]);
    iterator_count($it);
    $it->next();
    if (!$it->valid()) {
        echo <warning descr="Restore the iterator position after counting.">$it->current()</warning>;
    }
}

function countedAgain(): void
{
    $it = new ArrayIterator([7, 11]);
    iterator_count($it);
    $it->rewind();
    iterator_count($it);
    echo <warning descr="Restore the iterator position after counting.">$it->current()</warning>;
}
