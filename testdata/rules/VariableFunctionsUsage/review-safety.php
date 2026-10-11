<?php
// Rewrites that keep the call's behaviour, and the shapes that would not.
class Store {
    public static function make() { return static::class; }
    public static function grab(array &$rows) { return array_pop($rows); }
    public function take(&$row, $flag = 0) { return $row; }
    public function peek($row) { return $row; }
    public static function __callStatic($name, $args) { return $name; }
    public function __call($name, $args) { return $name; }
}

class Shop extends Store {
    public function run(Store $store, $hook, array $rows, array $rest) {
        // forward_static_call(): only self/static/parent keep forwarding.
        <weak_warning descr="Call it directly: 'self::make()'.">forward_static_call('self::make')</weak_warning>;
        <weak_warning descr="Call it directly: 'static::make()'.">forward_static_call(['static', 'make'])</weak_warning>;
        <weak_warning descr="Call it directly: 'parent::make()'.">forward_static_call(['parent', 'make'])</weak_warning>;
        <weak_warning descr="Call it directly: 'strtoupper('x')'.">forward_static_call('strtoupper', 'x')</weak_warning>;
        forward_static_call('Store::make');
        forward_static_call(['Store', 'make']);
        forward_static_call([$store, 'make']);
        forward_static_call($hook);
        forward_static_call(['Store', 'Store::make']);
        call_user_func(['Store', 'self::make']);

        // By-reference parameters: call_user_func() passes copies.
        call_user_func('sort', $rows);
        call_user_func('\array_multisort', $rows);
        call_user_func('Store::grab', $rows);
        call_user_func(['Store', 'grab'], $rows);
        call_user_func(['self', 'grab'], $rows);
        call_user_func(['parent', 'grab'], $rows);
        call_user_func(['Base', 'Store::grab'], $rows);
        call_user_func([$store, 'take'], $rows);
        call_user_func([$store, 'take'], flag: 1, row: $rows);
        call_user_func([$store, 'take'], ...$rest);
        <weak_warning descr="Call it directly: '$store->peek(...$rest)'.">call_user_func([$store, 'peek'], ...$rest)</weak_warning>;
        <weak_warning descr="Call it directly: '$store->take()'.">call_user_func([$store, 'take'])</weak_warning>;
        <weak_warning descr="Call it directly: 'count($rows)'.">call_user_func('count', $rows)</weak_warning>;
        <weak_warning descr="Call it directly: 'Store::peek($rows)'.">call_user_func('Store::peek', $rows)</weak_warning>;

        // A callable held in a variable is resolved through its assignments.
        $pop = 'array_pop';
        call_user_func($pop, $rows);
        $bump = function (&$n) { $n++; };
        call_user_func($bump, $rows);
        $fill = fn($first, &...$slots) => count($slots);
        call_user_func($fill, 1, $rows);
        $size = 'count';
        <weak_warning descr="Call it directly: '$size($rows)'.">call_user_func($size, $rows)</weak_warning>;
        $pair = [$store, 'peek'];
        <weak_warning descr="Call it directly: '$pair($rows)'.">call_user_func($pair, $rows)</weak_warning>;

        // Names that are not identifiers.
        <weak_warning descr="Call it directly: '$store->{'rows.list'}()'.">call_user_func([$store, 'rows.list'])</weak_warning>;
        <weak_warning descr="Call it directly: 'Store::{&quot;find-all&quot;}()'.">call_user_func(['Store', "find-all"])</weak_warning>;
        <weak_warning descr="Call it directly: '$store->{''}()'.">call_user_func([$store, ''])</weak_warning>;
        <weak_warning descr="Call it directly: '$store->né()'.">call_user_func([$store, 'né'])</weak_warning>;
        call_user_func('Store::find-all');
        call_user_func(['Store', 'Store::find-all']);
        call_user_func(['Store', 'a:b']);
        call_user_func(['Sto re', 'make']);
        call_user_func('Sto re::make');
        call_user_func('not a name', 1);
        call_user_func('');
    }
}
