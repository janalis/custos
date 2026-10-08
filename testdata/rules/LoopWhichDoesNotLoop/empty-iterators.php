<?php
class Lazy implements IteratorAggregate
{
    public function getIterator(): Iterator { return new ArrayIterator([]); }
}

/** @param object $o @param iterable $it @param mixed $m @param int[] $ids */
function init_collections(Lazy $lazy, $o, $it, $m, array $ids)
{
    // Iterating initialises the collection.
    foreach ($lazy as $item) {
    }
    foreach ($o as $item) {
    }
    foreach ($it as $item) {
    }
    foreach ($m as $item) {
    }
    <warning descr="Loop body exits on the first iteration; the loop never repeats.">foreach</warning> ($ids as $id) {
    }
}
