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
    public function <weak_warning descr="Declare ': int' as the return type.">documented</weak_warning>()
    {
        return static::SIDES;
    }

    public function <weak_warning descr="Declare ': int' as the return type.">viaSelf</weak_warning>()
    {
        return self::SIDES;
    }

    public function <weak_warning descr="Declare ': string' as the return type.">name</weak_warning>()
    {
        return static::NAME;
    }

    public function <weak_warning descr="Declare ': int' as the return type.">locked</weak_warning>()
    {
        return static::LOCKED;
    }

    public function <weak_warning descr="Declare ': int' as the return type.">typed</weak_warning>()
    {
        return static::TYPED;
    }

    public function <weak_warning descr="Declare ': string' as the return type.">className</weak_warning>()
    {
        return static::class;
    }

    public function <weak_warning descr="Declare ': int' as the return type.">other</weak_warning>(Shape $shape)
    {
        return $shape::LOCKED;
    }

    public function dynamic(string $name)
    {
        return static::{$name};
    }

    public function <weak_warning descr="Declare ': int' as the return type.">viaCall</weak_warning>()
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
