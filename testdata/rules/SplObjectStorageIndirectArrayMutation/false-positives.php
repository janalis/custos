<?php

function replacedThroughOffsetSet(): void
{
    $s = new SplObjectStorage();
    $o = new stdClass();
    $s[$o] = ["c" => 1];
    $s->offsetSet($o, new ArrayObject(["c" => 1]));
    $s[$o]["c"] = 2;
}

function replacedThroughAttach(): void
{
    $s = new SplObjectStorage();
    $o = new stdClass();
    $s[$o] = ["c" => 1];
    $s->attach($o, new ArrayObject());
    unset($s[$o]["c"]);
}

function replacedOnOneBranch(bool $flag): void
{
    $s = new SplObjectStorage();
    $o = new stdClass();
    $s[$o] = ["c" => 1];
    if ($flag) {
        $s[$o] = new ArrayObject(["c" => 1]);
    }
    $s[$o]["c"]++;
}

function replacedThroughAnotherKey(object $p): void
{
    $s = new SplObjectStorage();
    $o = new stdClass();
    $s[$o] = ["c" => 1];
    $s[$p] = new ArrayObject();
    $s[$o]["c"] = 2;
}

function replacedWhileIterating(): void
{
    $s = new SplObjectStorage();
    $o = new stdClass();
    $s[$o] = ["c" => 1];
    foreach ($s as $item) {
        $s[$item] = new ArrayObject();
    }
    $s[$o]["c"] = 2;
}

function replacedByCapturingClosure(): void
{
    $s = new SplObjectStorage();
    $o = new stdClass();
    $s[$o] = ["c" => 1];
    $swap = function () use ($s, $o) {
        $s[$o] = new ArrayObject();
    };
    $swap();
    $s[$o]["c"] = 2;
}

function detached(): void
{
    $s = new SplObjectStorage();
    $o = new stdClass();
    $s[$o] = ["c" => 1];
    unset($s[$o]);
    $s[$o]["c"] = 2;
}
