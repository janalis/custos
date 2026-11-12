<?php
namespace Shop;

use Acme\Input;
use Acme\RawInput;

// Classes declared by a dependency, builtins whose return type changes in
// a later PHP version (sort() returns bool before PHP 8.2, true after).
function edges(Input $in, array $list, \ArrayObject $bag): array
{
    $sorted = sort($list);
    return [$in instanceof RawInput, $sorted === null, is_bool(sort($list)), is_null($bag->asort())];
}
