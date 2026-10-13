<?php

class Element
{
    public function id() { return 'id'; }

    public function viaAncestor(Node $other)
    {
        return <warning descr="Call 'name' on an instance with '->' instead of '::'.">$other::name()</warning>;
    }
}

class Node extends Element
{
    public Node $next;

    public string $summary {
        get => <warning descr="Call 'name' on an instance with '->' instead of '::'.">$this->next::name()</warning>;
    }

    public function name() { return 'node'; }

    public function same(Node $other)
    {
        return <warning descr="Call 'name' on an instance with '->' instead of '::'.">$other::name()</warning>
            . <warning descr="Call 'id' on an instance with '->' instead of '::'.">$other::id()</warning>
            . <warning descr="Call 'name' on an instance with '->' instead of '::'.">$this::name()</warning>
            . <warning descr="Call 'name' on an instance with '->' instead of '::'.">($this)::name()</warning>;
    }

    public function nested(Node $other)
    {
        $closure = function () use ($other) {
            return <warning descr="Call 'name' on an instance with '->' instead of '::'.">$other::name()</warning>;
        };
        $arrow = fn () => <warning descr="Call 'name' on an instance with '->' instead of '::'.">$other::name()</warning>;
        $staticClosure = static function () use ($other) {
            return <warning descr="Call 'name' on an instance with '->' instead of '::'.">$other::name()</warning>;
        };
        $staticArrow = static fn () => <warning descr="Call 'name' on an instance with '->' instead of '::'.">$other::name()</warning>;

        return [$closure, $arrow, $staticClosure, $staticArrow];
    }

    public function declares(Node $other)
    {
        function helper(Node $inner)
        {
            return <warning descr="Call 'name' on an instance with '->' instead of '::'.">$inner::name()</warning>;
        }

        return new class {
            public function read(Node $inner)
            {
                return <warning descr="Call 'name' on an instance with '->' instead of '::'.">$inner::name()</warning>;
            }
        };
    }

    public static function fromStatic(Node $other)
    {
        return <warning descr="Call 'name' on an instance with '->' instead of '::'.">$other::name()</warning>;
    }
}

trait Shared
{
    public function shared() { return 'shared'; }

    public function reads(Node $other)
    {
        return <warning descr="Call 'name' on an instance with '->' instead of '::'.">$other::name()</warning>;
    }
}

class Leaf
{
    use Shared;

    public function reads2(Leaf $other)
    {
        return <warning descr="Call 'shared' on an instance with '->' instead of '::'.">$other::shared()</warning>;
    }
}

class Stranger
{
    public function reads(Node $other, Leaf $leaf)
    {
        return <warning descr="Call 'name' on an instance with '->' instead of '::'.">$other::name()</warning>
            . <warning descr="Call 'shared' on an instance with '->' instead of '::'.">$leaf::shared()</warning>;
    }
}

enum Level
{
    case Low;

    public function label() { return 'low'; }

    public function describe()
    {
        return <warning descr="Call 'label' on an instance with '->' instead of '::'.">self::Low::label()</warning>;
    }
}
