<?php
class Stub
{
    // `return null;` is a compile error in a void function.
    public function getLabel()
    {
        return null;
    }

    /** @param null $v */
    public function echoNull($v)
    {
        return $v;
    }

    // A bare `return;` is a compile error under ?Stub.
    public function maybeSelf($flag)
    {
        if ($flag) {
            return;
        }
        return new Stub();
    }

    public function reset($flag): void
    {
        if ($flag) {
            return;
        }
    }

    public function find($flag): ?\Stub
    {
        if ($flag) {
            return null;
        }
        return new Stub();
    }

    // Generators accept a bare return; the call always returns a Generator.
    public function items($flag): \Generator
    {
        if ($flag) {
            return;
        }
        yield 1;
    }
}
