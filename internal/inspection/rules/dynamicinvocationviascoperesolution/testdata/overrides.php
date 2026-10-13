<?php
class Printer
{
    public function run() { return <warning descr="Call 'operation' with '$this->' instead of '::'.">self::operation()</warning>; }
    public function operation() { return 'base'; }
    public function tags() { return <warning descr="Call 'initTags' with '$this->' instead of '::'.">self::initTags()</warning> . <warning descr="Call 'secret' with '$this->' instead of '::'.">self::secret()</warning> . <warning descr="Call 'locked' with '$this->' instead of '::'.">self::locked()</warning> . <warning descr="Call 'operation' with '$this->' instead of '::'.">static::operation()</warning>; }
    public function initTags() { return 'base'; }
    private function secret() { return 's'; }
    final public function locked() { return 'l'; }
}

class NetPrinter extends Printer
{
    public function operation() { return 'net'; }
    public function initTags() { return <warning descr="Call 'operation' with '$this->' instead of '::'.">Printer::operation()</warning>; }
}

final class Sealed
{
    public function a() { return <warning descr="Call 'b' with '$this->' instead of '::'.">self::b()</warning>; }
    public function b() { return 1; }
}

enum Mode
{
    case On;
    public function label() { return <warning descr="Call 'name2' with '$this->' instead of '::'.">self::name2()</warning>; }
    public function name2() { return 'x'; }
}

trait Greets
{
    public function hello() { return <warning descr="Call 'name' with '$this->' instead of '::'.">self::name()</warning>; }
    public function name() { return 't'; }
}
