<?php
class ConstructorValue {}
class ConstructorConsumer
{
    public function __construct(ConstructorValue $value) {}
}

// Constructor parameter resolution does not expand this rule's U11 scope.
function constructorArgument(?ConstructorValue $value)
{
    new ConstructorConsumer($value);
}
