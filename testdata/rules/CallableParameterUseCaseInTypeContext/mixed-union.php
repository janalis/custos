<?php

/** @return mixed|false */
function lookup(string $key)
{
    return $key === '' ? false : $key;
}

/** @return int|false */
function position(string $key)
{
    return strpos($key, ':');
}

class Helper
{
    /**
     * @param string|null $iface
     */
    public static function run(callable $cb, $iface = null)
    {
        if (func_num_args() > 2) {
            $iface = func_get_arg(2);
        }
        $iface = lookup('x');
        return $cb($iface);
    }

    public static function pos(?string $label = null)
    {
        $label = <warning descr="Assigning a value of type bool does not match the parameter's declared type.">position('a:b')</warning>;
        return $label;
    }
}
