<?php
namespace Shop\Cart;

echo \array_reduce([1, 2], initial: 'strlen', callback: fn($c, $i) => $c . $i);
echo \array_reduce([1, 2], initial: 0, callback: '\max');
\array_map(array: [' a'], callback: "\\trim");
\array_map('\trim', ...$lists);
\array_filter(...$args, mode: 0);
\array_reduce(array: [], initial: 'strlen');
