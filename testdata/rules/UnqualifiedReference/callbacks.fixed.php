<?php
namespace Shop\Cart;

function money($v) { return $v; }

\array_map('no_such_function', []);
\array_map('money', []);
\array_filter([], 'runtime_defined_later');
\array_map('\trim', []);
\call_user_func('\Strtoupper', 'x');
