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

    public function <weak_warning descr="Declare ': void' as the return type.">reset</weak_warning>($flag)
    {
        if ($flag) {
            return;
        }
    }

    public function <weak_warning descr="Declare ': ?\Stub' as the return type.">find</weak_warning>($flag)
    {
        if ($flag) {
            return null;
        }
        return new Stub();
    }

    // Generators accept a bare return; the call always returns a Generator.
    public function <weak_warning descr="Declare ': \Generator' as the return type.">items</weak_warning>($flag)
    {
        if ($flag) {
            return;
        }
        yield 1;
    }
}
