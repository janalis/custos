<?php
class Node
{
    public $value;
    public function next(): ?Node { return null; }
    public function take(Node $n) {}
    public function walk(?Node $head, Node $tail = null, ?int $count = null, Node $plain, $loose = null)
    {
        <warning descr="Possible null dereference.">$head</warning>->value = 1;
        <warning descr="Possible null dereference.">$head</warning>();
        if ($head === null) {
            return;
        }
        $head->value = 2;

        $x = clone <warning descr="Possible null dereference.">$tail</warning>;
        $this->take(<warning descr="Possible null dereference.">$tail</warning>);
        echo $tail->value ?? 'none';
        if (isset($tail->value['a'])) {}
        if ($tail) {
            $tail->value = 3;
        }
        $plain->value = 4;
        $loose->value = 5;
        $count++;
    }

    public function chain()
    {
        $this->next()<warning descr="Possible null dereference.">-></warning>next();
        if ($this->next() != null) {
            $this->next()->next();
        }
        $n = $this->next();
        <warning descr="Possible null dereference.">$n</warning>->next();
        <warning descr="Possible null dereference.">$n</warning>['k'];
        $n = new Node();
        $n->value = 0;

        /** @var Node $m */
        $m = $this->next();
        $m->value = 1;

        $cursor = $this->next();
        while ($cursor !== null) {
            $cursor = $cursor->next();
        }

        $probe = $this->next();
        <warning descr="Possible null dereference.">$probe</warning>->value = 1;
        $probe = <warning descr="Possible null dereference.">$probe</warning>->next();

        $f = function () {
            return $this->next()<warning descr="Possible null dereference.">-></warning>next();
        };
    }

    public function refill(Node $left = null, ?Node $right, Node $up = null)
    {
        $left = $left ?: null;
        <warning descr="Possible null dereference.">$left</warning>->value = 1;

        $right = $right ?? null;
        <warning descr="Possible null dereference.">$right</warning>->next();

        $up = <warning descr="Possible null dereference.">$up</warning>->next();
    }

    /** @param Node[] $list */
    public function rewind(array $list)
    {
        foreach ($list as $item) {
            $item = $item->next();
            <warning descr="Possible null dereference.">$item</warning>->value = 1;
        }
    }

    public function casing(?Node $a, ?Node $b)
    {
        self::AssertNotNull($a);
        Verify::that($b)->NotNull();
        return [$a->value, $b->value];
    }
}

function visit(?Node $item) {
    \Assert\Assertion::notNull($item);
    return $item->value;
}

function deref(Node|null $item) {
    return <warning descr="Possible null dereference.">$item</warning>->value;
}

class Shelf
{
    public function maybe(): ?Shelf { return null; }
    public function items(): array { return []; }
    public function find(int $id): ?Shelf { return null; }
    public function first(int $id)
    {
        return $this->find($id)?->maybe()<warning descr="Possible null dereference.">-></warning>items();
    }
}
