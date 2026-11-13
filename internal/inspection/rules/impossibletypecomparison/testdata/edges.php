<?php
namespace Shop;

use Acme\Input;
use Acme\Output;

final class Order {}

// Classes declared by a dependency: a later release may change how they
// relate, but not what a final class of the project extends. sort()
// returns bool before PHP 8.2 and true after.
function edges(Input $in, Output $out, Order $order, array $list): array
{
    $sorted = sort($list);
    return [
        $in === $out,
        <error descr="This comparison is always false: Shop\Order and Acme\Input share no value.">$order === $in</error>,
        $sorted === 'x',
        sort($list) === 1,
        $order == $in, // loosely, a dependency class is never judged
    ];
}

// A class Line extends, but no class here: the comparison stays open.
final class Line extends Missing {}

function unresolved(Line $line, Order $order): bool
{
    return $line === $order;
}
