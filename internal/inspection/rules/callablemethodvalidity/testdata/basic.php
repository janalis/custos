<?php
class Shop
{
    public function open() {}
    protected function audit() {}
    public static function hours() {}
    private static function vault() {}
}

class Outlet extends Shop {}

function probe(Shop $shop, $flag)
{
    $handler = ['Shop', 'open'];
    $either = $flag ? 'Shop::open' : 'Shop::hours';
    return [
        is_callable('ucfirst'),
        is_callable(<warning descr="Method 'open' is not static but is referenced without an object.">[Shop::class, 'open']</warning>),
        is_callable(<warning descr="Method 'open' is not static but is referenced without an object.">'Shop::open'</warning>),
        is_callable(<warning descr="Method 'open' is not static but is referenced without an object.">'\Shop::OPEN'</warning>),
        is_callable(<warning descr="Method 'open' is not static but is referenced without an object.">$handler</warning>),
        is_callable(<warning descr="Method 'open' is not static but is referenced without an object.">array('Outlet', 'open')</warning>),
        is_callable(['Shop', 'hours']),
        is_callable('Shop::hours'),
        is_callable([$shop, 'open']),
        is_callable([new Shop(), 'open']),
        is_callable([$shop, 'hours']),
        is_callable(<warning descr="Method 'audit' is not public, so the callback cannot be invoked from outside.">[$shop, 'audit']</warning>),
        is_callable(<warning descr="Method 'vault' is not public, so the callback cannot be invoked from outside.">[Shop::class, 'vault']</warning>),
        is_callable(['Shop::open']),
        is_callable(['Shop', 'open', 1]),
        is_callable(['c' => 'Shop', 'm' => 'open']),
        is_callable([$shop, 'missing']),
        is_callable(['Missing', 'open']),
        is_callable($either),
        is_callable('Shop::open', true),
        is_callable(),
    ];
}

$top = 'Shop::open';
is_callable($top);

class Locker { private function seal() {} }
function check() {
    return is_callable(<warning descr="Method 'seal' is not public, so the callback cannot be invoked from outside."><warning descr="Method 'seal' is not static but is referenced without an object.">['Locker', 'seal']</warning></warning>);
}
