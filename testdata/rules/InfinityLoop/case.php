<?php
class Loader
{
    public static function boot()
    {
        <error descr="Method calls itself unconditionally; this recursion never ends.">Static::boot()</error>;
    }

    protected function flush()
    {
        return <error descr="Method calls itself unconditionally; this recursion never ends.">SELF::flush()</error>;
    }

    public function getItems()
    {
        return <error descr="Method calls itself unconditionally; this recursion never ends.">$this->getitems()</error>;
    }

    public function load()
    {
        return Parent::load();
    }
}
