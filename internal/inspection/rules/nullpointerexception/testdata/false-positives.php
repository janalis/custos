<?php
class Node
{
    public $value;
    public function next(): ?Node { return null; }
    public function sure(): Node { return $this; }

    public function guarded(?Node $a, ?Node $b, ?Node $c, ?Node $d, ?Node $e)
    {
        if ($a instanceof Node) {}
        $a->value = 1;
        if (!empty($b)) {}
        $b->value = 1;
        $x = $c ? 1 : 0;
        $c->value = 1;
        $this->assertNotNull($d);
        $d->value = 1;
        $e?->next();
        $e::class;
        echo $e->value ?? 2;
    }

    public function locals()
    {
        $p = $this->value;
        $p->value = 1;
        $q = null;
        $q->value = 1;
        $r = $this->sure();
        $r->value = 1;
        $r->sure()->sure();
        if (is_null($s = $this->next())) {}
        /** @var Node|null $t */
        $t = $this->next();
        if ($t !== null) { $t->value = 1; }
        if ($this->next() && $this->next()->next()) {}
    }

    public function asserted(?Node $a, ?Node $b, ?Node $c, ?Node $d, ?Node $e, ?Node $f)
    {
        static::assertNotNull($a);
        Verify::that($b)->isObject()->notNull();
        $this->that($c)->notNull();
        $this?->assertInstanceOf(Node::class, $d);
        Assert::that($e)::notNull();
        is_null($f);

        return [$a->value, $b->value, $c->value, $d->value, $e->value, $f->value];
    }

    public function loop()
    {
        $cursor = $this->next();
        while ($cursor !== null) {
            $cursor = $cursor->next();
        }
    }

    /** @param Node[] $list */
    public function rewound(array $list)
    {
        foreach ($list as $item) {
            $item = $item->next();
        }
    }

    abstract function nobody(?Node $n);
}

class NodeTest
{
    public function testIt(?Node $n) { $n->value = 1; }
}

function visitStatic(?Node $item) {
    \Assert\Assertion::notNull($item);
    return $item->value;
}

class Region
{
    public function panel(): Region { return $this; }
    public function maybe(): ?Region { return null; }
    public function items(): array { return []; }
    public function find(int $id): ?Region { return null; }
    public function first(int $id)
    {
        return $this->find($id)?->panel()->items();
    }
}
