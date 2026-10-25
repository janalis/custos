<?php
function both(array $a, $b, $c)
{
    return isset($a['x']) // first
        && <weak_warning descr="Merge this check into the preceding isset() call.">isset($b)</weak_warning>;
}
