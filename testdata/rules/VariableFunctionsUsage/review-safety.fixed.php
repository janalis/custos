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
        self::make();
        static::make();
        parent::make();
        strtoupper('x');
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
        $store->peek(...$rest);
        $store->take();
        count($rows);
        Store::peek($rows);

        // A callable held in a variable is resolved through its assignments.
        $pop = 'array_pop';
        call_user_func($pop, $rows);
        $bump = function (&$n) { $n++; };
        call_user_func($bump, $rows);
        $fill = fn($first, &...$slots) => count($slots);
        call_user_func($fill, 1, $rows);
        $size = 'count';
        $size($rows);
        $pair = [$store, 'peek'];
        $pair($rows);

        // Names that are not identifiers.
        $store->{'rows.list'}();
        Store::{"find-all"}();
        $store->{''}();
        $store->né();
        call_user_func('Store::find-all');
        call_user_func(['Store', 'Store::find-all']);
        call_user_func(['Store', 'a:b']);
        call_user_func(['Sto re', 'make']);
        call_user_func('Sto re::make');
        call_user_func('not a name', 1);
        call_user_func('');
    }
}
