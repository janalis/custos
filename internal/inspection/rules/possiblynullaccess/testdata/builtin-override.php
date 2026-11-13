<?php
namespace Shop;

// A class of the project may override the method and return null.
class Lazy extends \CachingIterator
{
    public function getInnerIterator(): ?\Iterator { return null; }
}

function cached(\CachingIterator $c): void
{
    <warning descr="Check this value for null before the access: it is Iterator|null.">$c->getInnerIterator()</warning>->rewind();
}

// Subclasses that keep PHP's method.
final class Pages extends \LimitIterator
{
}

class Chunks extends \LimitIterator
{
}

class Rows extends Chunks
{
}

function limited(\LimitIterator $l, Pages $p, Chunks $c): void
{
    $l->getInnerIterator()->rewind();
    $p->getInnerIterator()->rewind();
    $c->getInnerIterator()->rewind();
}
