<?php
// SimpleXMLElement|false lookups; loosely documented index unions.
function stat(string $s): string
{
    $xml = simplexml_load_string($s);
    return (string) $xml['stat'];
}

/** @return string[]|string */
function unsubst($p)
{
    return $p;
}

class Paths
{
    private array $map = [];

    public function get($p)
    {
        return $this->map[unsubst($p)];
    }

    /** @param \stdClass|null $k */
    public function byObject($k)
    {
        return $this->map[<error descr="Index of type \stdClass does not fit the accepted string|int.">$k</error>];
    }

    public function typed(array $k)
    {
        return $this->map[<error descr="Index of type array does not fit the accepted string|int.">$k</error>];
    }
}
