<?php
function typed(string $s = '', int|float $n = 0, array $a = [], mixed $m = 1, $z = \null, $y = Null) {}

class Base {
    protected function render($mode) {}
}
class Mid extends Base {}
class Leaf extends Mid implements Countable {
    protected function render($mode = 'text') {}
    public function count(): int { return 0; }
}

class Clocked
{
    public function __construct(
        private $clock = new SystemClock(),
    ) {
    }

    public function at($zone = (new DefaultZone)) {}
}
