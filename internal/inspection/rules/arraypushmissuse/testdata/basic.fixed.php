<?php
$queue = [];
$queue[] = $job;

class Log {
    /** @var string[] */
    private $lines = [];
    private array $rows = [];

    public function add(array $items, string $line) {
        $this->lines[] = $line;
        $this->rows[] = [1, 2];
        $items[] = $line;
    }
}

$queue[count($queue)] = $job;
$this->items[count($this->items)] = $job;
$queue[\count($queue)] = &$job;
