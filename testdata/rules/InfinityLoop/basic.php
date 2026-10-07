<?php
class Settings
{
    private $theme;

    public function theme()
    {
        return <error descr="Method calls itself unconditionally; this recursion never ends.">$this->theme()</error>;
    }

    public static function boot($env)
    {
        <error descr="Method calls itself unconditionally; this recursion never ends.">static::boot($env)</error>;
    }

    protected function flush()
    {
        // delegate
        <error descr="Method calls itself unconditionally; this recursion never ends.">self::flush()</error>;
    }

    public function next()
    {
        /** @var int */
        return <error descr="Method calls itself unconditionally; this recursion never ends.">$this?->next(1, 2)</error>;
    }
}

trait Proxy
{
    public function call() { return <error descr="Method calls itself unconditionally; this recursion never ends.">$this->call()</error>; }
}
