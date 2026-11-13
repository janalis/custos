<?php
declare(strict_types=1);

// A polyfill of a function PHP declares in a later version may not be the
// declaration that runs.
if (!function_exists('array_first')) {
    function array_first(array $array) {}
}
array_first('x');
