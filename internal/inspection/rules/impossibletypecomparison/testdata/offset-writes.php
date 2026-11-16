<?php
namespace Shop;

// A write into false makes it an array (deprecated from PHP 8.1): the
// comparison may hold.
function vivify(): bool
{
    $list = false;
    $list[] = 1;
    return $list === [1];
}

// So does a write into an offset of a property holding false.
class LazyList
{
    private static array|false $items = false;

    public static function first(): bool
    {
        if (self::$items === false) {
            self::$items[] = 1;
            return self::$items === [1];
        }
        return false;
    }
}
