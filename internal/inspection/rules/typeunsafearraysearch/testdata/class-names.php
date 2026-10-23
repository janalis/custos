<?php
namespace App;

final class Header {}

function accepts($type, object $o): bool
{
    // Class names are never numeric strings.
    $a = \in_array($type, ['string', 'array', Header::class]);
    // `$o::class` reads a value.
    $b = <weak_warning descr="Pass a third argument to say whether this search must be type-strict.">\in_array($type, ['string', $o::class])</weak_warning>;
    return $a && $b;
}
