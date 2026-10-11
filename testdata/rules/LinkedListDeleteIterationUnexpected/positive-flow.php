<?php
function conditionalThenDelete(bool $keep): void
{
    $l = new SplQueue();
    if ($keep) {
        $l->setIteratorMode(SplDoublyLinkedList::IT_MODE_KEEP);
    }
    $l->setIteratorMode(SplDoublyLinkedList::IT_MODE_DELETE);
    foreach (<warning descr="Preserve linked-list elements during iteration.">$l</warning> as $item) {
    }
}

function keptOnlyAfterwards(): void
{
    $l = new SplQueue();
    $l->setIteratorMode(SplDoublyLinkedList::IT_MODE_DELETE);
    $l->push(1);
    foreach (<warning descr="Preserve linked-list elements during iteration.">$l</warning> as $item) {
    }
    $l->setIteratorMode(SplDoublyLinkedList::IT_MODE_KEEP);
    $later = function () use ($l) {
        $l->setIteratorMode(SplDoublyLinkedList::IT_MODE_KEEP);
    };
}

function unrelatedClosureAndArray(array $rows): void
{
    $l = new SplQueue();
    $l->setIteratorMode(SplDoublyLinkedList::IT_MODE_DELETE);
    $sizes = array_map(function ($row) {
        return count($row);
    }, $rows);
    $pair = [$sizes, $rows];
    foreach (<warning descr="Preserve linked-list elements during iteration.">$l</warning> as $item) {
    }
}
