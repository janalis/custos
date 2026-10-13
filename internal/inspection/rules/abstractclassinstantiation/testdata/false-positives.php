<?php
abstract class Shape
{
    // The runtime class decides what `static` creates.
    public static function make()
    {
        return new static();
    }

    // A closure can be rebound to another class scope.
    public function factory()
    {
        return function () {
            return new self();
        };
    }

    public function shortFactory()
    {
        return fn () => new self();
    }

    // Only evaluated when the argument is omitted.
    public function with($other = new self())
    {
        return $other;
    }

    // A nested function has no class scope: left to the runtime.
    public function nested()
    {
        function helper()
        {
            return new self();
        }
    }
}

abstract class Base
{
}

class Circle extends Base
{
    public static function make()
    {
        return new self();
    }

    public function base()
    {
        return new parent();
    }
}

// Inside a trait, `self` is the class using it.
trait Builds
{
    public function build()
    {
        return new self();
    }
}

abstract class Holder
{
    public function anonymous()
    {
        return new class {
            public function again()
            {
                return new self();
            }
        };
    }
}

// Outside any class.
$outside = static function () {
    return new self();
};
new self();
