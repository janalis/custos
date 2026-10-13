<?php
trait CollectionTrait
{
    public function isEmpty(): bool
    {
        foreach ($this as $el) {
            return false;
        }
        return true;
    }

    public function first(array $items)
    {
        <warning descr="Loop body exits on the first iteration; the loop never repeats.">foreach</warning> ($items as $item) {
            return $item;
        }
        return null;
    }
}
