<?php
// PHP declares getInnerIterator() nullable, but its iterator wrappers
// return the iterator their constructor took.
class OnlyFiles extends FilterIterator
{
    public function accept(): bool
    {
        $file = $this->getInnerIterator()->current();
        $inner = $this->getInnerIterator();
        return $inner->valid() && $file !== null;
    }
}

// AppendIterator has no inner iterator until one is appended.
function appended(AppendIterator $a)
{
    $a->getInnerIterator()<warning descr="Possible null dereference.">-></warning>rewind();
    $inner = $a->getInnerIterator();
    <warning descr="Possible null dereference.">$inner</warning>->rewind();
}
