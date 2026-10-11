<?php

function numbers()
{
    yield 1;
    yield 2;
}

// Two separate calls: each loop receives its own generator.
function freshCalls()
{
    foreach (numbers() as $_) {
    }
    foreach (numbers() as $n) {
        echo $n;
    }
}

// The closure rebinds the variable through its by-reference capture.
function reboundByClosure()
{
    $g = numbers();
    $renew = function () use (&$g) {
        $g = numbers();
    };
    $all = iterator_to_array($g);
    $renew();
    foreach ($g as $n) {
        echo $n;
    }
}

function renewShared()
{
    global $shared;
    $shared = numbers();
}

// Another function rebinds the global.
function reboundGlobal()
{
    global $shared;
    $shared = numbers();
    iterator_to_array($shared);
    renewShared();
    foreach ($shared as $n) {
        echo $n;
    }
}

// extract() imports a fresh generator under the same name.
function reboundByExtract()
{
    $g = numbers();
    iterator_to_array($g);
    extract(['g' => numbers()]);
    foreach ($g as $n) {
        echo $n;
    }
}

// The first loop overwrites its own iterable.
function overwrittenByLoop()
{
    $g = numbers();
    foreach ($g as $g) {
    }
    foreach ((array) $g as $n) {
        echo $n;
    }
}
