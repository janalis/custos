<?php
final class Evaluator
{
    private string $name = 'x';

    public static function create(): callable
    {
        return static function () {
            return $this->name;
        };
    }

    public static function arrow(): callable
    {
        return function () {
            return fn () => static fn () => $this->name;
        };
    }
}

function outside(): callable
{
    return fn () => $this;
}
