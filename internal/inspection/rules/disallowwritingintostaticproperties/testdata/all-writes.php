<?php

class Counter
{
    public static $value = 0;

    public function bump()
    {
        <weak_warning descr="Avoid modifying static properties.">self::$value = 1</weak_warning>;
        <weak_warning descr="Avoid modifying static properties.">static::$value *= 2</weak_warning>;
        <weak_warning descr="Avoid modifying static properties.">$this::$value = 3</weak_warning>;
        self::$value++;
        return self::$value;
    }
}

<weak_warning descr="Avoid modifying static properties.">Counter::$value = 7</weak_warning>;
