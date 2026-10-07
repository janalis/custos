<?php
class Report
{
    public function render()
    {
        $fmt = function () {
            $total = 0;
            return $total;
        };
        return <error descr="Variable '$total' is not defined in this scope.">$total</error> ?? $fmt();
    }

    public function nested()
    {
        $inner = function () {
            return isset(<error descr="Variable '$state' is not defined in this scope.">$state</error>);
        };
        $state = 1;
        return [$inner, $state];
    }

    public function declared()
    {
        function helper() { $flag = true; return $flag; }
        return empty(<error descr="Variable '$flag' is not defined in this scope.">$flag</error>);
    }

    public function shadowed()
    {
        $map = fn($key) => $key;
        return isset(<error descr="Variable '$key' is not defined in this scope.">$key</error>) ? $map : null;
    }

    public function imported()
    {
        $add = function () use (&$sum) { $sum = 1; };
        $add();
        return $sum ?? 0;
    }

    public function captured()
    {
        $get = fn() => $limit;
        return $limit ?? $get;
    }
}
