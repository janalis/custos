<?php

interface Registry
{
    <weak_warning descr="Reorder modifiers as: public static.">static public</weak_warning> function instance();
    public static function reset();
}

abstract class Repository
{
    public static function boot() {}
    abstract protected function table();
    final static function hash() {}
    protected
        static function cache() {}
    PRIVATE STATIC function lower() {}
    public function plain() {}
    static function single() {}

    <weak_warning descr="Reorder modifiers as: abstract protected.">protected abstract</weak_warning> function columns();
    <weak_warning descr="Reorder modifiers as: final protected static.">static
        protected final</weak_warning> function guard() {}
    <weak_warning descr="Reorder modifiers as: final private static.">Static Private Final</weak_warning> function seal() {}
    <weak_warning descr="Reorder modifiers as: abstract public static.">public static abstract</weak_warning> function make();
    public static $prop;
}
