<?php

function foreachRewinds(): void
{
    $it = new ArrayIterator(['a' => 1, 'b' => 2]);
    $total = iterator_count($it);
    foreach ($it as $value) {
        echo $it->key(), '=', $value, ' of ', $total, "\n";
    }
}

function forHeaderRewinds(): void
{
    $it = new ArrayIterator([10, 20]);
    $total = iterator_count($it);
    for ($it->rewind(); $it->valid(); $it->next()) {
        echo $it->key(), ':', $it->current(), ' of ', $total, "\n";
    }
}

function forHeaderSeeks(): void
{
    $it = new ArrayIterator([10, 20]);
    iterator_count($it);
    for ($it->seek(1); $it->valid(); $it->next()) {
        echo $it->current();
    }
}

function wrappedIterator(): void
{
    $it = new ArrayIterator([1, 2]);
    iterator_count($it);
    $limited = new LimitIterator($it, 0, 1);
    foreach ($limited as $item) {
        echo $it->current();
    }
}

function conditionalRewind(bool $again): void
{
    $it = new ArrayIterator([1, 2]);
    iterator_count($it);
    if ($again) {
        $it->rewind();
    }
    echo $it->current();
}

function aliasRewinds(): void
{
    $it = new ArrayIterator([1, 2]);
    $alias = $it;
    iterator_count($it);
    $alias->rewind();
    echo $it->current();
}

function closureRewinds(): void
{
    $it = new ArrayIterator([1, 2]);
    iterator_count($it);
    $restore = function () use ($it) {
        $it->rewind();
    };
    $restore();
    echo $it->current();
}

function emptyIterator(): void
{
    $it = new ArrayIterator([]);
    iterator_count($it);
    var_dump($it->key());
}
