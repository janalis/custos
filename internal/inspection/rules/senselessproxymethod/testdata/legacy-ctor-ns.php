<?php
namespace Streams;

class Reader
{
    public function __construct($path) {}
}

class Buffered extends Reader
{
    public function <weak_warning descr="Method '__construct' only forwards to its parent; remove it.">__construct</weak_warning>($path)
    {
        parent::__construct($path);
    }

    public function Buffered($path) {}
}
