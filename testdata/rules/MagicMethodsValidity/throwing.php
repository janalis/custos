<?php

class Ghost
{
    /** @var callable */
    private $render;

    public function __clone()
    {
    }

    public function __serialize(): array
    {
        return [];
    }


    public function __toString()
    {
        throw new \LogicException('not printable');
    }

    public function __sleep()
    {
        if (PHP_SAPI === 'cli') {
            throw new \LogicException('cli');
        }
        exit(1);
    }

    public function <error descr="__debugInfo must return array|null; got ''.">__debugInfo</error>()
    {
        if (PHP_SAPI === 'cli') {
            throw new \LogicException('cli');
        }
    }
}

final class Delegating
{
    /** @var callable */
    private $render;

    public function __toString()
    {
        return call_user_func($this->render);
    }
}
