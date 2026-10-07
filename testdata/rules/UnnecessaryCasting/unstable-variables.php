<?php
function grow(array $parts)
{
    $label = 'n';
    foreach ($parts as $p) {
        $label .= $p;
    }
    $count = 0;
    $count++;
    return [(string) $label, (int) $count];
}
