<?php

use function Vendor\Polyfill\is_iterable;

function imported($v) {
    // The bare name is the imported function here, so the built-in is written qualified.
    return \is_iterable($v);
}
