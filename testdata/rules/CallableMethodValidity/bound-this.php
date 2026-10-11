<?php
namespace Bound;

class Engine
{
    public function run() {}
    public static function make() {}
}

class Turbo extends Engine
{
    public function run()
    {
        $plain = function () { return is_callable('Bound\Engine::run'); };
        $arrow = fn() => is_callable([parent::class, 'run']);
        $fixed = static function () { return is_callable(<warning descr="Method 'run' is not static but is referenced without an object.">'Bound\Engine::run'</warning>); };
        $short = static fn() => is_callable(<warning descr="Method 'run' is not static but is referenced without an object.">[Engine::class, 'run']</warning>);
        $inner = static function () {
            return function () { return is_callable(<warning descr="Method 'run' is not static but is referenced without an object.">[self::class, 'run']</warning>); };
        };
        return [
            is_callable([parent::class, 'run']),
            is_callable([self::class, 'run']),
            is_callable([static::class, 'run']),
            is_callable([Engine::class, 'run']),
            is_callable(['Bound\Engine', 'run']),
            is_callable('Bound\Engine::run'),
            is_callable(<warning descr="Method 'run' is not static but is referenced without an object.">[Nitro::class, 'run']</warning>),
            is_callable(<warning descr="Method 'run' is not static but is referenced without an object.">[Bench::class, 'run']</warning>),
            $plain, $arrow, $fixed, $short, $inner,
        ];
    }

    public static function build()
    {
        return is_callable(<warning descr="Method 'run' is not static but is referenced without an object.">[parent::class, 'run']</warning>);
    }

    public function nested()
    {
        function helper() { return is_callable(<warning descr="Method 'run' is not static but is referenced without an object.">'Bound\Engine::run'</warning>); }
        $anon = new class extends Engine {
            public function t() { return is_callable([Engine::class, 'run']); }
        };
        $loose = new class {
            public function t() { return is_callable(<warning descr="Method 'run' is not static but is referenced without an object.">[Engine::class, 'run']</warning>); }
        };
        return [$anon, $loose];
    }
}

class Nitro extends Turbo {}

class Bench
{
    public function run() {}

    public function probe()
    {
        return [
            is_callable([self::class, 'run']),
            is_callable(<warning descr="Method 'run' is not static but is referenced without an object.">[Engine::class, 'run']</warning>),
        ];
    }
}

trait Tuning
{
    public function tune() { return is_callable([Engine::class, 'run']); }
    public static function preset() { return is_callable(<warning descr="Method 'run' is not static but is referenced without an object.">[Engine::class, 'run']</warning>); }
}

function outside()
{
    return is_callable(<warning descr="Method 'run' is not static but is referenced without an object.">[Engine::class, 'run']</warning>);
}

is_callable(<warning descr="Method 'run' is not static but is referenced without an object.">'Bound\Engine::run'</warning>);
