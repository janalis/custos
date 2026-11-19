<?php

function two($a, $b) {}

function readsAll($a)
{
    return func_get_args();
}

function readsOne($a)
{
    return \func_get_arg(1);
}

function counts($a)
{
    return FUNC_NUM_ARGS();
}

function includes($a)
{
    include 'other.php';
}

function evaluates($a)
{
    eval('return 1;');
}

function rest($a, ...$more) {}

function otherCall($a)
{
    return strlen($a);
}

interface Contract
{
    public function run($a);
}

abstract class Base
{
    abstract public function run($a);
}

class Plain
{
    public function __construct($a) {}

    public function run($a) {}
}

class NoConstructor {}

function calls(Contract $contract, Base $base, Plain $plain)
{
    two(1, 2, <error descr="Remove the extra argument: two() accepts at most 2.">3</error>);
    two(1, 2, <error descr="Remove the extra arguments: two() accepts at most 2.">3, 4</error>);
    otherCall(1, <error descr="Remove the extra argument: otherCall() accepts at most 1.">2</error>);
    $plain->run(1, <error descr="Remove the extra argument: Plain::run() accepts at most 1.">2</error>);
    new Plain(1, <error descr="Remove the extra argument: Plain::__construct() accepts at most 1.">2</error>);

    two(1, 2);
    readsAll(1, 2);
    readsOne(1, 2);
    counts(1, 2);
    includes(1, 2);
    evaluates(1, 2);
    rest(1, 2, 3);
    $contract->run(1, 2);
    $base->run(1, 2);
    new NoConstructor(1);
}
