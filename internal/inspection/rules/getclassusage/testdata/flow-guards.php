<?php
class Driver
{
    public function guarded($object = null)
    {
        if (is_object($object)) {
            return get_class($object);
        }
        return '';
    }

    public function guardedApart($object = null, $parent = '')
    {
        $known = is_subclass_of($object, $parent);
        log_it($known);
        return get_class($object);
    }

    public function replaced(?Driver $target = null): string
    {
        $target = $target ?: $this;
        return get_class($target);
    }

    public function unguarded($object = null): string
    {
        return <warning descr="get_class() rejects null on PHP 7.2+; guard the argument.">get_class($object)</warning>;
    }

    public function typedNullable(?Driver $target = null): string
    {
        return <warning descr="get_class() rejects null on PHP 7.2+; guard the argument.">get_class($target)</warning>;
    }
}
