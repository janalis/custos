<?php
namespace Shop;

// IteratorIterator::getInnerIterator() is declared nullable, but returns
// the iterator its constructor took: null never comes back.
class OnlyFiles extends \FilterIterator
{
    public function accept(): bool
    {
        $file = $this->getInnerIterator()->current();
        $inner = $this->getInnerIterator();
        return $inner->valid() && $file !== null;
    }
}

class Wrapped extends \IteratorIterator
{
    public function inner(): ?\Iterator { return null; }

    public function first(): mixed
    {
        // A method of the project declared nullable is still reported.
        return <warning descr="Check this value for null before the access: it is Iterator|null.">$this->inner()</warning>->current();
    }
}

function rewind(\IteratorIterator $it): void
{
    $it->getInnerIterator()->rewind();
}

function wrappers(\LimitIterator $l, \CachingIterator $c, \NoRewindIterator $n, \RegexIterator $r, \CallbackFilterIterator $cb, \InfiniteIterator $i): void
{
    $l->getInnerIterator()->rewind();
    $c->getInnerIterator()->rewind();
    $n->getInnerIterator()->rewind();
    $r->getInnerIterator()->rewind();
    $cb->getInnerIterator()->rewind();
    $i->getInnerIterator()->rewind();
}

// AppendIterator has no inner iterator until one is appended.
function appended(\AppendIterator $a): void
{
    <warning descr="Check this value for null before the access: it is Iterator|null.">$a->getInnerIterator()</warning>->rewind();
}

// A receiver that may be something else than an object.
function mixedReceiver(Wrapped|false $w): void
{
    <warning descr="Check this value for null before the access: it is Iterator|null.">$w->inner()</warning>->current();
}
