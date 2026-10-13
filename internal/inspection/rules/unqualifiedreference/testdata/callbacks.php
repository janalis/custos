<?php
namespace Shop\Cart;

function money($v) { return $v; }

\array_map('no_such_function', []);
\array_map('money', []);
\array_filter([], 'runtime_defined_later');
\array_map(<weak_warning descr="Write '\trim' to allow compile-time binding.">'trim'</weak_warning>, []);
\call_user_func(<weak_warning descr="Write '\Strtoupper' to allow compile-time binding.">'Strtoupper'</weak_warning>, 'x');
