<?php
if (!function_exists('is_real')) {
    function is_real($value) { return is_float($value); }
}

$f = is_float($ratio);
