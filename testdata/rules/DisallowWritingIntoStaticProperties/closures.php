<?php

class Registry
{
    protected static $resolver;
    public static $hits = 0;

    public static function reset(): \Closure
    {
        $outer = function () {
            static::$resolver = null;
            $inner = fn () => Registry::$hits = 0;
            $inner();
        };
        return $outer;
    }
}

class Consumer
{
    public function run(): void
    {
        $cb = function () {
            <weak_warning descr="Modify this static property only from the class that declares it.">Registry::$hits = 1</weak_warning>;
        };
        $cb();
    }
}

$loose = function () {
    <weak_warning descr="Modify this static property only from the class that declares it.">Registry::$hits = 2</weak_warning>;
};
