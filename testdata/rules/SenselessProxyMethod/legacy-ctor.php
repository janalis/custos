<?php
class LegacyReader
{
    public function __construct($path) {}
}

class CachedLegacyReader extends LegacyReader
{
    public function __construct($path)
    {
        parent::__construct($path);
    }

    public function CachedLegacyReader($path)
    {
        self::__construct($path);
    }
}

class PlainLegacyReader extends LegacyReader
{
    public function <weak_warning descr="Method '__construct' only forwards to its parent; remove it.">__construct</weak_warning>($path)
    {
        parent::__construct($path);
    }

    public function read() {}
}
