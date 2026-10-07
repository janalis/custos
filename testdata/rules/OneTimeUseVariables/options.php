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
    <warning descr="Variable $value is used only once; inline its value.">$value</warning> = someVeryLongFunctionName($firstArgument, $secondArgument, $thirdArgument, 42);
    [$a, $b] = $value;
    return $a . $b;
}
