<?php
$queue = [];
<warning descr="Use '$queue[] = $job' instead; it avoids a function call.">array_push($queue, $job)</warning>;

class Log {
    /** @var string[] */
    private $lines = [];
    private array $rows = [];

    public function add(array $items, string $line) {
        <warning descr="Use '$this->lines[] = $line' instead; it avoids a function call.">\array_push($this->lines, $line)</warning>;
        <warning descr="Use '$this->rows[] = [1, 2]' instead; it avoids a function call.">array_push($this->rows, [1, 2])</warning>;
        <warning descr="Use '$items[] = $line' instead; it avoids a function call.">array_push($items, $line)</warning>;
    }
}

$queue[count($queue)] = $job;
$this->items[count($this->items)] = $job;
$queue[\count($queue)] = &$job;
