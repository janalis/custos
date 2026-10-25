<?php
class Node
{
    public array $children = [];
    public function kids(): array { return []; }
}

class Finder
{
    public function find(array $tree, string $id): ?Node { return null; }

    public function walk(array $tree, string $id)
    {
        foreach ($tree as $node) {
            $node = $this->find($node->children, $id);
            if ($node) {
                return $node;
            }
        }
        return null;
    }

    public function first(array $pairs, string $id)
    {
        [$node] = $pairs;
        $node = $this->find($node->kids(), $id);
        return <warning descr="Possible null dereference.">$node</warning>->children;
    }
}
