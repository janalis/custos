<?php
class Typed
{
    private ?int $n = <weak_warning descr="Explicit null default is redundant; remove it.">null</weak_warning>;
    public function __construct() { <weak_warning descr="Assignment writes the property's default value; remove it.">$this->n = null;</weak_warning> }
}
