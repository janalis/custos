<?php
namespace Shop;

// A write into an offset of null makes it an array; an object
// implementing ArrayAccess stays an object.
class Cache
{
    private ?array $items = null;
    private ?\ArrayObject $store = null;

    public function put(string $key): void
    {
        $this->items[$key] = 1;
        if (<warning descr="This check is always false: the value is array.">$this->items === null</warning>) {
            echo 'empty';
        }
        $this->store[$key] = 1;
        if (<warning descr="This check is always false: the value is ArrayObject|array.">is_null($this->store)</warning>) {
            echo 'empty';
        }
        if ($this->store instanceof \ArrayObject) {
            echo 'object';
        }
    }
}

function fill(array|false $rows, ?array $seen, string $key): void
{
    // A write into false (deprecated) is not followed: false stays.
    $rows[] = 1;
    if (is_array($rows)) {
        echo 'rows';
    }
    $seen[$key]++;
    if (<warning descr="This check is always false: the value is array.">is_null($seen)</warning>) {
        echo 'unseen';
    }
}

function found(): array|false
{
    return false;
}

// A write into false makes it an array too (deprecated from PHP 8.1): the
// check is not decided either way.
function vivify(): void
{
    $list = false;
    $list[] = 1;
    if (is_array($list)) {
        echo 'list';
    }
    $rows = found();
    $rows['k'] = 1;
    if (is_bool($rows)) {
        echo 'failed';
    }
}

// A loop writing into an offset makes null an array once it runs: after a
// loop over a non-empty literal the variable is an array, and after one
// that may not run it may be either.
function loops(array $rows): void
{
    $seen = null;
    foreach ($rows as $seen['last']) {
    }
    if ($seen === null) {
        echo 'none';
    }
    $sizes = null;
    foreach ([1, 2] as $sizes[]) {
    }
    if (<warning descr="This check is always false: the value is array.">is_null($sizes)</warning>) {
        echo 'none';
    }
}

// So does a write into an offset of a property holding false.
class LazyList
{
    private static array|false $items = false;

    public static function fill(): void
    {
        if (self::$items === false) {
            self::$items[] = 1;
            if (is_bool(self::$items)) {
                echo 'failed';
            }
        }
    }
}
