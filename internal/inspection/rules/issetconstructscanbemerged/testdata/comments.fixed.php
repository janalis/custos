<?php
function both(array $a, $b, $c)
{
    return isset($a['x']) // first
        && isset($b);
}
