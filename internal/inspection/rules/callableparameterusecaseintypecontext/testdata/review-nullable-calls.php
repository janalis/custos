<?php
// A nullable class result of a call next to a class-typed parameter.
namespace Review\Nullable;

enum Suit: string
{
    case Hearts = 'h';
}

class Node
{
    public ?Node $parent = null;

    public function up(): ?Node
    {
        return $this->parent;
    }

    public static function find(int $id): ?Node
    {
        return null;
    }

    public function depth(): ?int
    {
        return null;
    }
}

function findNode(int $id): ?Node
{
    return null;
}

function root(Node $node): Node
{
    $last = $node;
    while ($node = $node->up()) {
        $last = $node;
    }
    return $last;
}

function lookups(Node $a, Node $b, Suit $s, \Closure $c, object $o, Node $d, int $n)
{
    $a = Node::find(1);
    $b = findNode(1);
    $s = Suit::tryFrom('h');
    $c = $c->bindTo($o);
    $c = \Closure::bind($c, $o);
    // The scalar part of a nullable result is still checked.
    $d = <warning descr="Assigning a value of type int does not match the parameter's declared type.">$d->depth()</warning>;
    // A null next to a scalar parameter is still reported.
    $n = <warning descr="Assigning a value of type null does not match the parameter's declared type.">$d->depth()</warning>;
    return [$a, $b, $s, $c, $d, $n];
}
