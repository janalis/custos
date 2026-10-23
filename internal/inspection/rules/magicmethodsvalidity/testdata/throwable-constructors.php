<?php
// The constructors of Exception and Error only store arguments that
// default to what the properties hold.
final class LazyResponseException extends \Exception
{
    public function __construct(private string $response) {}
}

final class EngineFailure extends \Error
{
    public function __construct(public readonly int $step) {}
}

// Another builtin parent may do work of its own.
final class Bag extends \ArrayObject
{
    public function <error descr="__construct does not call ArrayObject::__construct().">__construct</error>(array $items) {}
}
