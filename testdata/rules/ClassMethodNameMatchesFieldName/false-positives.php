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
