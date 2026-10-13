<?php
class Shape
{
    const SIDES = 0;
    const NAME = 'shape';
    final const LOCKED = 1;
    const int TYPED = 2;

    public function sides()
    {
        return static::SIDES;
    }

    public function viaThis()
    {
        return $this::SIDES;
    }

    /** @return int */
    public function documented()
    {
        return static::SIDES;
    }

    public function viaSelf(): int
    {
        return self::SIDES;
    }

    public function name(): string
    {
        return static::NAME;
    }

    public function locked(): int
    {
        return static::LOCKED;
    }

    public function typed(): int
    {
        return static::TYPED;
    }

    public function className(): string
    {
        return static::class;
    }

    public function other(Shape $shape): int
    {
        return $shape::LOCKED;
    }

    public function dynamic(string $name)
    {
        return static::{$name};
    }

    public function viaCall(): int
    {
        return self::make()::LOCKED;
    }

    private static function make(): Shape
    {
        return new Shape();
    }
}

class Blob extends Shape
{
    const SIDES = 'none';
}

trait Counts
{
    public function count()
    {
        return static::TOTAL;
    }
}
