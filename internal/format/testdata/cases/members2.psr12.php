<?php
// presets: per laravel psr12
class A
{
    use T1;
    use T2;
    use T3, T4 { foo as bar; }

    public const X = 1;
    public $v;
    public static $s;
    public function f()
    {
    }
    abstract public function g();
    final public static function h()
    {
    }
    public function __construct(
        #[\SensitiveParameter] public $secret,
        #[A] #[B] int $n,
    ) {}
}
$o = new class (1) {};
$long = $this->someProperty[$key]
    = compute();
$x =
    compute();
