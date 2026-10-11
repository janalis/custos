<?php
namespace Shop\Cart;

echo \array_reduce([1, 2], initial: 'strlen', callback: fn($c, $i) => $c . $i);
echo \array_reduce([1, 2], initial: 0, callback: <weak_warning descr="Write '\max' to allow compile-time binding.">'max'</weak_warning>);
\array_map(array: [' a'], callback: <weak_warning descr="Write '\trim' to allow compile-time binding.">"trim"</weak_warning>);
\array_map(<weak_warning descr="Write '\trim' to allow compile-time binding.">'trim'</weak_warning>, ...$lists);
\array_filter(...$args, mode: 0);
\array_reduce(array: [], initial: 'strlen');
