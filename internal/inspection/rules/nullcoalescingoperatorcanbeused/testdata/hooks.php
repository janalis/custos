<?php
class HookLookup
{
    private array $items = [];

    public mixed $reference {
        &get {
            if (isset($this->items['current'])) {
                return $this->items['current'];
            }
            return $this->items;
        }
    }

    public mixed $value {
        get {
            <weak_warning descr="Simplify to 'return $this->items['current'] ?? null' using the null coalescing operator.">if</weak_warning> (isset($this->items['current'])) {
                return $this->items['current'];
            }
            return null;
        }
    }
}
