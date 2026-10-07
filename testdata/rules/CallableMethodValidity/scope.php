<?php
namespace App;

trait Hooks { private function hook() {} }

class Base
{
    use Hooks;

    private function secret() {}
    protected static function guarded() {}

    public function self()
    {
        $fn = function () { return is_callable([$this, 'secret']); };
        return [
            is_callable([$this, 'secret']),
            is_callable([self::class, 'guarded']),
            is_callable([$this, 'hook']),
            is_callable('App\Base::guarded'),
            $fn(),
        ];
    }
}

class Child extends Base
{
    public function probe()
    {
        return [
            is_callable([$this, 'guarded']),
            is_callable([Base::class, 'guarded']),
            is_callable(<warning descr="Method 'secret' is not public, so the callback cannot be invoked from outside.">[$this, 'secret']</warning>),
        ];
    }
}

class Stranger
{
    public function probe(Base $b)
    {
        $anon = new class extends Base {
            public function t() { return is_callable([$this, 'guarded']); }
        };
        return [
            is_callable(<warning descr="Method 'guarded' is not public, so the callback cannot be invoked from outside.">[$b, 'guarded']</warning>),
            is_callable(<warning descr="Method 'secret' is not public, so the callback cannot be invoked from outside.">[$b, 'secret']</warning>),
        ];
    }
}

function outside(Base $b)
{
    return is_callable(<warning descr="Method 'guarded' is not public, so the callback cannot be invoked from outside.">[$b, 'guarded']</warning>);
}
