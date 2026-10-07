<?php

class Counter
{
    public static $total = 0;

    public function bump()
    {
        $add = function ($n) {
            SELF::$total += $n;
            Self::$total = $n;
            Counter::$total = $n;
        };
        $reset = fn() => STATIC::$total = 0;
        return [$add, $reset];
    }
}
