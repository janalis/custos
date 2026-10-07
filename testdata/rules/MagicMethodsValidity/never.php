<?php
// A declared `never` always throws: PHP accepts it for every magic method
// with a return contract.
final class Sealed
{
    public function __serialize(): never
    {
        throw new \LogicException('sealed objects cannot be serialized');
    }

    public function __toString(): never
    {
        throw new \LogicException('no text form');
    }

    public function __debugInfo(): never
    {
        throw new \LogicException('hidden');
    }

    public function <error descr="__sleep must return array; got 'int'.">__sleep</error>(): int
    {
        throw new \LogicException('mixed contract');
    }
}
