<?php
class Collector
{
    private array $data = [];
    private int $count = 0;
    private array $roots;
    private array $copies;

    public function __construct(array $seed)
    {
        // Holding references makes clones share $data and $count.
        $this->roots = [&$this->data, &$this->count];
        <weak_warning descr="Private property is only used in the constructor; likely dead code.">$this->copies</weak_warning> = [$seed, 'key' => $this->count];
    }

    public function add($item): void
    {
        $this->data[] = $item;
        ++$this->count;
    }
}
