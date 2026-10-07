<?php

class Account
{
    public $id;
    protected static $count = 0;
    private readonly int $balance;
    public(set) string $owner;
    protected(set) string $bank;

    public function __construct(private int $x) {}
    protected function load() {}
    private static function cache() {}
    final public function seal() {}
    public const A = 1;
    private const B = 2;
}
