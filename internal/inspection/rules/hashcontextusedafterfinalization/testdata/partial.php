<?php
// The variable holds a closure that creates a context on each call: no
// context was finalized.
function later()
{
    $h = hash_init('sha256', ...);
    hash_final($h);
    hash_update($h, 'x');
}
