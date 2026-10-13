<?php
$top = <weak_warning descr="Closure does not use $this; declare it static.">function</weak_warning> () { return 1; };

class Holder
{
    private $hook;
    private static $shared;

    public function dynamicSelf($name)
    {
        return array_map(function () use ($name) { return self::$name(); }, []);
    }

    public function stores($c)
    {
        $this->hook = function () { return 1; };
        self::$shared = function () { return 2; };
        $list['k'] = function () { return 3; };
        $c ??= function () { return 4; };
        $picked = $c ? <weak_warning descr="Closure does not use $this; declare it static.">function</weak_warning> () { return 5; } : null;
    }

    public function binds($obj, $unknown)
    {
        $a = <weak_warning descr="Closure does not use $this; declare it static.">function</weak_warning> () { return 6; };
        $a::bindTo(null);
        $b = function () { return 7; };
        $obj->bind($b, null);
        $unknown->bind(function () { return 8; }, null);
        $keys = [<weak_warning descr="Closure does not use $this; declare it static.">function</weak_warning> () { return 9; } => 1];
    }

    public function bind($cb, $scope) {}
}
