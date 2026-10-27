<?php

class Bag
{
    public function __construct(
        protected array $items = [],
    ) {
    }
}

class FileItems extends Bag
{
    public function __construct(array $items = [])
    {
        $this->replace($items);
    }

    public function replace(array $items = []): void
    {
        $this->items = [];
        foreach ($items as $k => $v) {
            $this->items[$k] = (string) $v;
        }
    }
}

class DirectItems extends Bag
{
    public function __construct(array $items = [])
    {
        $this->items = array_map('strval', $items);
    }
}

abstract class Indirect extends Bag
{
    public function <error descr="__construct does not call Bag::__construct().">__construct</error>(array $items = [])
    {
        $this->fill($items);
        static::{$items[0]}();
        $this->undeclared();
    }

    public function fill(array $items): void
    {
        $this->load($items);
    }

    public function load(array $items): void
    {
        $this->items = $items;
    }

    abstract public function undeclared(): void;
}
