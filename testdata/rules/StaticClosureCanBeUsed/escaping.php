<?php
abstract class Base
{
    public static function make() { return 1; }
}

final class Factory extends Base
{
    private $hook;

    public static function helper() { return 1; }
    public function instance() { return 2; }

    public function returned() {
        return function () { return 1; };
    }

    public function returnedArrow() {
        return (fn () => 2);
    }

    public function returnedVar() {
        $cb = function () { return 3; };
        return $cb;
    }

    public function stored() {
        $cb = function () { return 4; };
        $this->hook = $cb;
    }

    public function captured() {
        $cb = function () { return 5; };
        return array_map(static function () use ($cb) { return 0; }, []);
    }

    public function invoked() {
        $cb = <weak_warning descr="Closure does not use $this; declare it static.">function</weak_warning> () { return 6; };
        return $cb() + $cb();
    }

    public function selfInstance() {
        return array_map(function () { return self::instance(); }, []);
    }

    public function staticInstance() {
        return array_map(function () { return static::instance(); }, []);
    }

    public function selfUnknown() {
        return array_map(function () { return self::missing(); }, []);
    }

    public function selfStatic() {
        return array_map(<weak_warning descr="Closure does not use $this; declare it static.">function</weak_warning> () { return self::helper() + static::make(); }, []);
    }

    public function nestedClass() {
        return array_map(<weak_warning descr="Closure does not use $this; declare it static.">function</weak_warning> () {
            return new class { public function a() { return self::b(); } public function b() {} };
        }, []);
    }
}
