<?php

use function Vendor\Polyfill\is_iterable;

function imported($v) {
    // The bare name is the imported function here, so the built-in is written qualified.
    return <weak_warning descr="Use 'is_iterable($v)' instead of the is_array()/instanceof Traversable pair.">is_array($v)</weak_warning> || $v instanceof Traversable;
}
