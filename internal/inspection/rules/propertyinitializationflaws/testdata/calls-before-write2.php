<?php
class Registry
{
    private $items = [];
    private $names = [];
    private $tag = <weak_warning descr="Default is always replaced by the constructor; remove it.">'x'</weak_warning>;

    public function __construct(array $items, array $names)
    {
        $this->tag = 'y';
        $fn = function () { return $this->items; }; // not called here
        register_registry($this); // may read the defaults
        $this->items = $items;
        $this->names = $names;
    }
}
