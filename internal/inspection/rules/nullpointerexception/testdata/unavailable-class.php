<?php
function reflectedName(string $type): string
{
    return (new ReflectionEnum($type))->getBackingType()->getName();
}
