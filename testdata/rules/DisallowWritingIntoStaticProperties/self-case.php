<?php

class Counter
{
    public static $total = 0;

    public function bump()
    {
        $add = function ($n) {
            SELF::$total += $n;
            Self::$total = $n;
            <weak_warning descr="Modify this static property only from the class that declares it.">Counter::$total = $n</weak_warning>;
        };
        $reset = fn() => <weak_warning descr="Modify this static property only from the class that declares it.">STATIC::$total = 0</weak_warning>;
        return [$add, $reset];
    }
}
