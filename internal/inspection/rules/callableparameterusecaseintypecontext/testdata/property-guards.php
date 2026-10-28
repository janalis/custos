<?php
final class Slugger
{
    private bool|string $emoji = false;

    public function guarded(?string $locale): ?string
    {
        if (\is_string($this->emoji)) {
            $locale = $this->emoji;
        }
        return $locale;
    }

    public function written(?string $locale): ?string
    {
        if (\is_string($this->emoji)) {
            $this->emoji = true;
            $locale = <warning descr="Assigning a value of type bool does not match the parameter's declared type.">$this->emoji</warning>;
        }
        return $locale;
    }

    public function unguarded(?string $locale): ?string
    {
        $locale = <warning descr="Assigning a value of type bool does not match the parameter's declared type.">$this->emoji</warning>;
        return $locale;
    }
}

class Node {}
final class Leaf extends Node {}

final class Tree
{
    private ?Node $root = null;

    public function leaf(?Node $node): ?Node
    {
        if ($this->root instanceof Leaf) {
            $node = $this->root;
        }
        return $node;
    }
}
