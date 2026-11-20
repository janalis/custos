<?php // presets: laravel drupal; set: property_separation=one, constant_separation=one, method_separation=one
class A
{
    use T1;
    use T2 { foo as bar; }
    const X = 1;
    const Y = 2;
    public $a;
    public $b;
    public function f() {}
    public $c;
    const Z = 3;
    case Q;
    public function g()
    {
        $a = 1;
        yield 1;
        yield 2;
        $x = yield 3;
        throw new E();
    }
}
