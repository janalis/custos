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
    public function read() {}
}
