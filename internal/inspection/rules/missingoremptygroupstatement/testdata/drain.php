<?php
function drain(Iterator $chunks, iterable $items, array $rows, \ArrayObject $bag): void
{
    foreach ($chunks as $chunk) {
    }
    foreach ($items as $item) {}
    foreach ($bag as $b) {}
    foreach (gen() as $v) {
    }
    <weak_warning descr="This construct has an empty body block.">foreach</weak_warning> ($rows as $row) {}
    <weak_warning descr="This construct has an empty body block.">foreach</weak_warning> ($unknown as $u) {}
    <weak_warning descr="This construct has an empty body block.">foreach</weak_warning> ([new ArrayIterator([])] as $it) {}
}

function gen(): Generator { yield 1; }
