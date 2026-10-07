<?php
namespace Shop;

class <weak_warning descr="Depends on 7 distinct classes; consider splitting it up.">Registry</weak_warning> {
    const Unit UNIT = Unit::One;
    private Store $store;

    public function load()
    {
        Loader::boot();
        function helper(): Config { return new Config(); }
        $f = function (): Clock { return Clock::now(); };
        $g = fn(): Cache => Cache::make();
        return [$f, $g, #[Pure] fn() => 1];
    }
}
