<?php
class Printer
{
    public function run() { return self::operation(); }
    public function operation() { return 'base'; }
    public function tags() { return self::initTags() . $this->secret() . $this->locked() . $this->operation(); }
    public function initTags() { return 'base'; }
    private function secret() { return 's'; }
    final public function locked() { return 'l'; }
}

class NetPrinter extends Printer
{
    public function operation() { return 'net'; }
    public function initTags() { return Printer::operation(); }
}

final class Sealed
{
    public function a() { return $this->b(); }
    public function b() { return 1; }
}

enum Mode
{
    case On;
    public function label() { return $this->name2(); }
    public function name2() { return 'x'; }
}

trait Greets
{
    public function hello() { return self::name(); }
    public function name() { return 't'; }
}
