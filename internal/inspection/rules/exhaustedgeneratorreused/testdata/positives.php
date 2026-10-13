<?php

function items()
{
    yield 1;
    yield 2;
}

function emptyLoopThenLoop()
{
    $g = items();
    foreach ($g as $_) {
    }
    <error descr="Create a fresh generator before traversing again.">foreach ($g as $v) {
        echo $v;
    }</error>
}

// A by-value capture cannot rebind the outer variable.
function capturedByValue()
{
    $g = items();
    $peek = function () use ($g) {
        return $g;
    };
    iterator_to_array($g);
    $peek();
    <error descr="Create a fresh generator before traversing again.">foreach ($g as $v) {
        echo $v;
    }</error>
}

function namedArguments()
{
    $g = items();
    iterator_to_array(iterator: $g, preserve_keys: false);
    <error descr="Create a fresh generator before traversing again.">foreach ($g as $v) {
    }</error>
}
