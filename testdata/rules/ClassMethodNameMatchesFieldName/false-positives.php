<?php
class Counter {
    private $total = 0;
    private ?string $label;
    /** @var int */
    private $hits;
    private $list = [];
    const size = 3;
    public function total() {}
    public function label() {}
    public function hits() {}
    public function list() {}
    public function size() {}
    public function Total() {}
}

interface Named {
    public function name();
}

// E4: properties of built-in classes are not reported as "type unknown".
class ApiFailure extends \RuntimeException {
    public function file(): string { return basename($this->file); }
    public function line(): int { return $this->line; }
    public function code(): int { return (int) $this->code; }
}
