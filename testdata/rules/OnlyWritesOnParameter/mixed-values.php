<?php
function appendTo(\ReflectionProperty $prop, object $owner, object $item): void
{
    $list = $prop->getValue($owner);
    $list[] = $item;
}

function mixedLocal(callable $make): void
{
    $bag = $make();
    $bag['k'] = 1;
}
