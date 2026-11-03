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

function skipped(object $constraint, array $types, $any): array
{
    // The class name of an object is never numeric either.
    $class = get_class($constraint);
    return [
        \in_array($class, $types),
        \in_array(get_class($constraint), $types),
        \in_array($constraint::class, $types),
        <weak_warning descr="Pass a third argument to say whether this search must be type-strict.">\in_array(get_class(), $types)</weak_warning>,
        <weak_warning descr="Pass a third argument to say whether this search must be type-strict.">\in_array(gettype($any), $types)</weak_warning>,
        <weak_warning descr="Pass a third argument to say whether this search must be type-strict.">\in_array($constraint::{$any}, $types)</weak_warning>,
        <weak_warning descr="Pass a third argument to say whether this search must be type-strict.">\in_array(Header::NAME, $types)</weak_warning>,
    ];
}
