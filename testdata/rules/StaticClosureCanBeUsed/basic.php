<?php
abstract class Base {
    public function work() {}
    public static function tool() {}
}

final class Runner extends Base {
    private $limit = 3;

    public function cases() {
        $squares = array_map(<weak_warning descr="Closure does not use $this; declare it static.">function</weak_warning> ($n) { return $n * $n; }, [1, 2]);
        $capped  = array_map(function ($n) { return min($n, $this->limit); }, [4]);
        $helper  = array_map(<weak_warning descr="Closure does not use $this; declare it static.">function</weak_warning> ($n) { return parent::tool(); }, [5]);
        $inst    = array_map(function ($n) { return parent::work(); }, [6]);
        $noop    = array_map(<weak_warning descr="Closure does not use $this; declare it static.">function</weak_warning> () { ; }, []);
        $empty   = array_map(function () {}, []);
        $twice   = array_map(<weak_warning descr="Closure does not use $this; declare it static.">fn</weak_warning> ($n) => $n * 2, [7]);
        $self    = array_map(fn ($n) => $n + $this->limit, [8]);
        $nested  = array_map(function () { return fn() => "{$this->limit}"; }, []);
        $attr    = array_map(#[Pure] <weak_warning descr="Closure does not use $this; declare it static.">fn</weak_warning> ($n) => $n, [9]);
    }

    public function rebinding() {
        $detached = <weak_warning descr="Closure does not use $this; declare it static.">function</weak_warning> () { return 42; };
        $copy = Closure::bind($detached, null, self::class);
        $other = $detached->bindTo(null);

        $attached = function () { return 7; };
        $attached->bindTo($this);
    }

    public function dispatching($queue) {
        $job = <weak_warning descr="Closure does not use $this; declare it static.">function</weak_warning> () { return 'ok'; };
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
    'home' => <weak_warning descr="Closure does not use $this; declare it static.">function</weak_warning> () { return 'index'; },
    'about' => static function () { return 'about'; },
];
$plain = [function () { return 0; }];
