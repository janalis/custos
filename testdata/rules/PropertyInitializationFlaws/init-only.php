<?php
class Cart
{
    protected $notes = null;
    private $total = 0;
    private $lines = [];
    public function __construct()
    {
        $this->total = 100;
        <weak_warning descr="Assignment writes the property's default value; remove it.">$this->lines = [];</weak_warning>
    }
}
