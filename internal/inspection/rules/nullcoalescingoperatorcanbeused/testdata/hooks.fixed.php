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
            return $this->items['current'] ?? null;
        }
    }
}
