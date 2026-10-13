<?php
class Counter
{
    public static function Counter() { return new self(); }
}

class Wallet
{
    public function __construct($amount) {}
    public function Wallet($amount) { $this->__construct($amount); }
}

class Purse
{
    public function Purse() {}
    public function __CONSTRUCT() {}
}

trait Logger
{
    public function Logger() {}
}

interface Printer
{
    public function Printer();
}

enum Suit
{
    case Hearts;
    public function Suit() {}
}

$x = new class {
    public function Anon() {}
};
