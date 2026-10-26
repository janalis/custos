<?php
final class Context
{
    private array $listeners = [];
}

final class Tracer
{
    public function wrap(Context $context): void
    {
        \Closure::bind(function () {
            $this->listeners = [];
        }, $context, Context::class)();
        (function () {
            $this->listeners = [];
        })->call($context);
        $binder->bind(function () {
            $this->listeners = [];
        }, $context);
        $fn = function () {
            <warning descr="Declare this property before assigning it.">$this->listeners</warning> = [];
        };
        (function () {
            <warning descr="Declare this property before assigning it.">$this->listeners</warning> = [];
        })->other($context);
        \Closure::fromCallable(function () {
            <warning descr="Declare this property before assigning it.">$this->listeners</warning> = [];
        });
        array_map(function () {
            <warning descr="Declare this property before assigning it.">$this->listeners</warning> = [];
        }, []);
        $x = function () {
            <warning descr="Declare this property before assigning it.">$this->listeners</warning> = [];
        };
        $x->bindTo($context);
        <warning descr="Declare this property before assigning it.">$this->listeners</warning> = [];
        $y = fn () => <warning descr="Declare this property before assigning it.">$this->listeners</warning> = [];
    }
}
