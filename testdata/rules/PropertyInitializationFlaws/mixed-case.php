<?php
class Box
{
    private $items = [];

    public function __Construct()
    {
        <weak_warning descr="Assignment writes the property's default value; remove it.">$this->items = [];</weak_warning>
    }
}
