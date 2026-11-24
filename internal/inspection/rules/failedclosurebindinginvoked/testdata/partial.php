<?php
// The closure binds later: calling it is not calling a failed binding.
function bindLater(Closure $closure, object $object)
{
    $bind = Closure::bind($closure, $object, ...);
    $bind();
}
