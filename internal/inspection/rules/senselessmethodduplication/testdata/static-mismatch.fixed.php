<?php
class Registry
{
    public function create() { return new \ArrayObject(); }
    public static function shared() { return 42; }
    public static function build() { return 7; }
}

class LocalRegistry extends Registry
{
    public static function create() { return new \ArrayObject(); }
    public function shared() { return 42; }
}
