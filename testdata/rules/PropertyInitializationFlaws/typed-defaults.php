<?php
final class Typed
{
    private array $items = [];
    private int $count = 0;

    public function __construct(array $items)
    {
        // Without its default a typed property is uninitialised for objects
        // built without the constructor (unserialize, reflection).
        $this->items = $items;
        $this->count = \count($items);
    }
}
