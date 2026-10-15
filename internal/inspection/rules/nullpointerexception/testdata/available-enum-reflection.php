<?php
function reflectedName(string $type): string
{
    return (new ReflectionEnum($type))->getBackingType()<warning descr="Possible null dereference.">-></warning>getName();
}
