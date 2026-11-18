<?php
function longOne() {
    $value = someVeryLongFunctionName($firstArgument, $secondArgument, $thirdArgument, 42);
    return $value;
}

function thrower() {
    $e = new Oops();
    throw $e;
}

function destructure() {
    [$a, $b] = someVeryLongFunctionName($firstArgument, $secondArgument, $thirdArgument, 42);
    return $a . $b;
}
