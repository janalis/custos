<?php
// A documented `object|resource` is a handle in its PHP 8 object form or
// its older resource form.
final class Handle {}
final class Other {}

final class Response
{
    /** @var object|resource|null */
    private $handle;
    /** @var Handle|resource */
    private $typed;
    /** @var Other|resource */
    private $other;
    /** @var resource */
    private $stream;

    public function open(Handle|string $ch, Handle|string $h, Handle|string $o, Handle|string $s): void
    {
        if (!$ch instanceof Handle) {
            $ch = $this->handle;
        }
        $h = $this->typed;
        $o = <warning descr="Assigning a value of type \Other does not match the parameter's declared type.">$this->other</warning>;
        $s = <warning descr="Assigning a value of type resource does not match the parameter's declared type.">$this->stream</warning>;
    }
}
