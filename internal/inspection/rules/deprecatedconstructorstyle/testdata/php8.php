<?php
class Query
{
    private $sql = 'select 1';

    public function query(): string { return "run {$this->sql}"; }
}

class Filter
{
    public function filter($v) { return $v * 2; }
}

class Legacy
{
    public function Legacy() {}
}
