<?php
class Base
{
    public function __construct() {}
}

class Pool extends Base
{
    private $removed = [];
    private $byName = [];
    private $seen = [];
    private $count = <weak_warning descr="Default is always replaced by the constructor; remove it.">0</weak_warning>;
    private $label = '';

    public function __construct(array $packages, array $removed, array $byName, array $seen)
    {
        $this->count = count($packages);
        parent::__construct(); // may call an overridden method reading a default
        $this->byName = $byName;
        $this->load($packages); // may read $this->removed before it is set
        $this->removed = $removed;
        $this->label = $this->describe(); // the call may read the default
        register_pool($this);
        $this->seen = $seen;
    }

    private function load(array $packages): void {}

    private function describe(): string { return $this->label . '!'; }
}
