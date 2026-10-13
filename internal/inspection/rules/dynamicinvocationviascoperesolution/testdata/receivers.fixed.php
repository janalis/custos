<?php

class Element
{
    public function id() { return 'id'; }

    public function viaAncestor(Node $other)
    {
        return $other::name();
    }
}

class Node extends Element
{
    public Node $next;

    public string $summary {
        get => $this->next::name();
    }

    public function name() { return 'node'; }

    public function same(Node $other)
    {
        return $other::name()
            . $other::id()
            . $this->name()
            . ($this)->name();
    }

    public function nested(Node $other)
    {
        $closure = function () use ($other) {
            return $other::name();
        };
        $arrow = fn () => $other::name();
        $staticClosure = static function () use ($other) {
            return $other->name();
        };
        $staticArrow = static fn () => $other->name();

        return [$closure, $arrow, $staticClosure, $staticArrow];
    }

    public function declares(Node $other)
    {
        function helper(Node $inner)
        {
            return $inner->name();
        }

        return new class {
            public function read(Node $inner)
            {
                return $inner->name();
            }
        };
    }

    public static function fromStatic(Node $other)
    {
        return $other->name();
    }
}

trait Shared
{
    public function shared() { return 'shared'; }

    public function reads(Node $other)
    {
        return $other::name();
    }
}

class Leaf
{
    use Shared;

    public function reads2(Leaf $other)
    {
        return $other::shared();
    }
}

class Stranger
{
    public function reads(Node $other, Leaf $leaf)
    {
        return $other->name()
            . $leaf::shared();
    }
}

enum Level
{
    case Low;

    public function label() { return 'low'; }

    public function describe()
    {
        return self::Low::label();
    }
}
