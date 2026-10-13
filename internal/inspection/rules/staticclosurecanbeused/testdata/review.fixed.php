<?php
abstract class Origin
{
    public function work($n) { return $n; }
    public static function tool() { return 1; }
}

function review_consume($label, $body) {}

final class Reviewed extends Origin
{
    private $handler;

    public function helper($n) { return $n; }
    public static function shared() { return 1; }
    public static function macro($name, $cb) {}

    public function classNames($n)
    {
        $a = array_map(function ($n) { return Reviewed::helper($n); }, []);
        $b = array_map(function ($n) { return Origin::work($n); }, []);
        $c = array_map(static function ($n) { return Reviewed::shared() + Origin::tool(); }, []);
        $d = array_map(static function ($n) { return $n::make(); }, []);
    }

    public function loads($file)
    {
        call_user_func(function () use ($file) { include $file; });
        call_user_func(function () use ($file) { return require_once $file; });
        call_user_func(function () use ($file) { return eval($file); });
    }

    public function forwarded($h, $k)
    {
        $this->handler = $h ?? function () { return 1; };
        $this->handler = function () { return 2; } ?? $h;
        $this->handler = $k ? function () { return 3; } : (function () { return 4; });
        $this->handler = match ($k) { 1 => function () { return 5; }, default => null };
        $m = match ($k) { 1 => static function () { return 6; }, default => null };
        $n = match (true) { function () { return 7; } => 1, default => 2 };
        $o = function () { return 8; } ? 1 : 2;
        $p = function () { return 9; } == $h;
        $q = array_map($h ?? static function () { return 10; }, []);
        $r = static fn () => function () { return 11; };
        return $h ?: function () { return 12; };
    }

    public function shortForm($h)
    {
        return function () { return 13; } ?: $h;
    }

    public function receivers($obj)
    {
        (function () { return 14; })->call($obj);
        (function () { return 15; })->bindTo($obj);
        (static function () { return 16; })->bindTo(null);
        (static function () { return 17; })::bindTo(null);
        (static function () { return 18; })();
        static function () { return 19; };
        yield function () { return 20; };
    }

    public function consumers($name)
    {
        review_consume('adds', function () { return 21; });
        Reviewed::macro('upper', function () { return 22; });
        self::$name(function () { return 23; });
        Closure::fromCallable(static function () { return 24; });
        $cb = function () { return 25; };
        static::macro('lower', $cb);
        parent::work(function () { return 26; });
    }

    public function chained($h)
    {
        $this->handler = $kept = function () { return 27; };
        if ($probe = function () { return 28; }) { }
        return $made = function () { return 29; };
    }

    public function loadsNested($file)
    {
        $run = function () use ($file) {
            call_user_func(function () use ($file) { include $file; });
        };
        $run();
        call_user_func(fn () => call_user_func(function () use ($file) { include $file; }));
        $detached = static function () use ($file) {
            call_user_func(static function () use ($file) { include $file; });
        };
        $short = static fn () => call_user_func(static function () use ($file) { include $file; });
    }

    public static function loadsStatically($file)
    {
        call_user_func(static function () use ($file) { return require $file; });
    }
}

function review_load($file)
{
    spl_autoload_register(static function ($class) { require __DIR__ . '/' . $class . '.php'; });
    call_user_func(static function () use ($file) { return eval($file); });
}

spl_autoload_register(static function ($class) { require __DIR__ . '/' . $class . '.php'; });
