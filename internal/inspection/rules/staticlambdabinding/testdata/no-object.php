<?php
final class Evaluator
{
    private string $name = 'x';

    public static function create(): callable
    {
        return static function () {
            return <error descr="Static closures have no $this; remove 'static' or avoid $this.">$this</error>->name;
        };
    }

    public static function arrow(): callable
    {
        return function () {
            return fn () => static fn () => <error descr="Static closures have no $this; remove 'static' or avoid $this.">$this</error>->name;
        };
    }
}

function outside(): callable
{
    return static fn () => <error descr="Static closures have no $this; remove 'static' or avoid $this.">$this</error>;
}
