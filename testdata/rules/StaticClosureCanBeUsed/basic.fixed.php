<?php
abstract class Base {
    public function work() {}
    public static function tool() {}
}

final class Runner extends Base {
    private $limit = 3;

    public function cases() {
        $squares = array_map(static function ($n) { return $n * $n; }, [1, 2]);
        $capped  = array_map(function ($n) { return min($n, $this->limit); }, [4]);
        $helper  = array_map(static function ($n) { return parent::tool(); }, [5]);
        $inst    = array_map(function ($n) { return parent::work(); }, [6]);
        $noop    = array_map(static function () { ; }, []);
        $empty   = array_map(function () {}, []);
        $twice   = array_map(static fn ($n) => $n * 2, [7]);
        $self    = array_map(fn ($n) => $n + $this->limit, [8]);
        $nested  = array_map(function () { return fn() => "{$this->limit}"; }, []);
        $attr    = array_map(#[Pure] static fn ($n) => $n, [9]);
    }

    public function rebinding() {
        $detached = static function () { return 42; };
        $copy = Closure::bind($detached, null, self::class);
        $other = $detached->bindTo(null);

        $attached = function () { return 7; };
        $attached->bindTo($this);
    }

    public function dispatching($queue) {
        $job = static function () { return 'ok'; };
        Registry::push($job);

        $task = function () { return 'later'; };
        $queue->push($task);

        new Worker(function () { return 1; });
    }

    public function scopedMap() {
        return ['a' => function () { return 1; }];
    }
}

$routes = [
    'home' => function () { return 'index'; },
    'about' => static function () { return 'about'; },
];
$plain = [function () { return 0; }];

final class Registry
{
    public static function push(callable $job): void {}
}

/** @method static void listen(callable $cb) */
final class Events
{
    public static function __callStatic($name, $args) {}
}
Events::listen(function () { return 1; });

final class Bus
{
    public static function __callStatic($name, $args) {}
}
Bus::dispatch(function () { return 2; });
Unknown::run(static function () { return 3; });
