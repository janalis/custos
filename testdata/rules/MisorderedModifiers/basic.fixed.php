<?php

interface Registry
{
    public static function instance();
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

    abstract protected function columns();
    final protected static function guard() {}
    final private static function seal() {}
    abstract public static function make();
    public static $prop;
}
