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
