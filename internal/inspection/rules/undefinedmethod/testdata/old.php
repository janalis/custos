<?php
namespace Shop;

function reflect(\ReflectionProperty $property, \ReflectionClass $class): void
{
    // PHP 8.0 adds these methods: code for several versions may call them.
    $property->hasDefaultValue();
    $class->getAttributes();
    $class-><error descr="Fix the method name: ReflectionClass declares no getAttribute().">getAttribute</error>();
}
