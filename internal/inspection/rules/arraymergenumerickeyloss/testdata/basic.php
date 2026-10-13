<?php
<warning descr="Preserve numeric identifier keys when combining arrays.">array_merge([42=>"a"],[87=>"b"])</warning>;
<warning descr="Preserve numeric identifier keys when combining arrays.">array_merge(['label' => 'x', 1 => 'y'], [])</warning>;
<warning descr="Preserve numeric identifier keys when combining arrays.">array_merge(['x' => 1, 'y' => 2, 2 => 'z'])</warning>;
<warning descr="Preserve numeric identifier keys when combining arrays.">array_merge([1 => 'a', 0 => 'b'])</warning>;
<warning descr="Preserve numeric identifier keys when combining arrays.">array_merge([-1 => 'a'])</warning>;
<warning descr="Preserve numeric identifier keys when combining arrays.">array_merge(['a', 'b', 3 => 'c'])</warning>;
<warning descr="Preserve numeric identifier keys when combining arrays.">array_merge([0 => 'a', 0 => 'b', 2 => 'c'])</warning>;
<warning descr="Preserve numeric identifier keys when combining arrays.">array_merge($unknown, [$key => 'a', 'b', 5 => 'c'])</warning>;
<warning descr="Preserve numeric identifier keys when combining arrays.">array_merge([$key => 'a', 1 => 'b', 3 => 'c'])</warning>;
