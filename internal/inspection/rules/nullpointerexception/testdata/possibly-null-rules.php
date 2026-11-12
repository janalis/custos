<?php
// With PossiblyNullAccess and PossiblyNullArgument enabled, the dereferences
// they report are not reported again here; the others still are.
class Node
{
    public $value;
    public function next(): ?Node { return null; }
    public function take(Node $n) {}

    public function walk(?Node $head, ?Node $tail, ?Node $copy)
    {
        <warning descr="Check this value for null before the access: it is Node|null.">$head</warning>->value = 1;
        $this->take(<warning descr="Check this value for null before passing it to $n of Node::take(): it is Node|null.">$tail</warning>);
        $x = clone <warning descr="Possible null dereference.">$copy</warning>;
        <warning descr="Check this value for null before the access: it is Node|null.">$this->next()</warning>->next();
    }

    /** @return Node|null */
    public function documented() { return null; }

    public function chain()
    {
        $this->documented()<warning descr="Possible null dereference.">-></warning>next();
        $d = $this->documented();
        $this->take(<warning descr="Possible null dereference.">$d</warning>);
    }
}
