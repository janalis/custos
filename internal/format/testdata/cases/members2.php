<?php // presets: per laravel psr12
class A {
    use T1, T2;
    use T3, T4 { foo as bar; }
    const X = 1;
    var $v;
    static $s;
    function f() {}
    abstract function g();
    final static function h() {}
    public function __construct(
        #[\SensitiveParameter] public $secret,
        #[A] #[B] int $n,
    ) {}
}
$o = new class(1) {};
$long = $this->someProperty[$key]
    = compute();
$x =
    compute();
