<?php
if (!function_exists('is_real')) {
    function is_real($value) { return is_float($value); }
}

$f = <warning descr="Use 'is_float(...)' instead of the alias 'is_real(...)'.">is_real</warning>($ratio);
