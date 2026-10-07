<?php

namespace Shop;

use Shop\Cache as Store;

class Cache
{
    public static $hits = 0;
    public static $ttl;

    public function touch()
    {
        self::$hits += 1;
        static::$ttl = 30;
        Store::$ttl = 60;
        SELF::$ttl = &$x;

        array_map(function ($v) {
            <weak_warning descr="Modify this static property only from the class that declares it.">Cache::$hits = $v</weak_warning>;
            self::$hits = $v;
        }, []);
        $f = fn() => <weak_warning descr="Modify this static property only from the class that declares it.">static::$ttl ??= 1</weak_warning>;
    }
}

class DiskCache extends Cache
{
    public static $ttl = 5;

    public function reset()
    {
        DiskCache::$ttl = 0;
        <weak_warning descr="Modify this static property only from the class that declares it.">parent::$hits = 0</weak_warning>;
        <weak_warning descr="Modify this static property only from the class that declares it.">DiskCache::$hits = 0</weak_warning>;
        Unknown::$whatever = 1;
        Cache::$hits++;
        Cache::$list[] = 1;
        $obj::$hits = 2;
        Cache::$$name = 3;
        echo Cache::$hits;
    }
}

function warmUp()
{
    <weak_warning descr="Modify this static property only from the class that declares it.">Cache::$ttl = 10</weak_warning>;
}

<weak_warning descr="Modify this static property only from the class that declares it.">Nowhere::$thing .= 'x'</weak_warning>;
echo Cache::$hits;
