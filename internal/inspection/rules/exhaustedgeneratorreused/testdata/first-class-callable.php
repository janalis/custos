<?php
// A first-class callable holds a closure: each traversal sees no generator.
function numbers()
{
    yield 1;
}

function traverseTwice()
{
    $g = numbers(...);
    foreach ($g as $_) {
    }
    foreach ($g as $v) {
        echo $v;
    }
}
