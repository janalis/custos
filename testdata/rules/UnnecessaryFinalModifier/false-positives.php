<?php
class Gauge
{
    public int $level {
        final get => 1;
    }
    final public function read() {}
    final protected function write() {}
    final private function __destruct() { flush(); }
    private function plain() {}
    final function implicit() {}
}

enum Mode
{
    final public function label() {}
}

$anon = new class {
    final public function run() {}
};

final class Sealed
{
    public function open() {}
    private function hidden() {}
}
