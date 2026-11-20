<?php
// presets: per laravel psr12
class A
{
    use T1, T2;
    use T3, T4 { foo as bar; }

    const X = 1;

    public $v;

    public static $s;

    public function f() {}

    abstract public function g();

    final public static function h() {}

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
function g(null|int|string $a, (A&B)|null $b, int|null $c, ?int $d): null|static {}
switch ($x) {
    case 1:  return 1;
    case 22: return 2;
    default: { return 3; }
}
$y = $foo
    ->bar();
class B
{
    public function class() {}
}
$this->swap(Vite::class, new class extends Vite
{
    public function f() {}
});
