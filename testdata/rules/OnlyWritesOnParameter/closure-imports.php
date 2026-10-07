<?php

function makeHandlers(ArrayObject $registry, array $plain, string $template)
{
    $store = new SplObjectStorage();
    $register = function ($name) use ($registry, $plain, $store) {
        $registry[$name] = true;
        $store[] = $name;
        <weak_warning descr="Value is only written here and never read; the write is lost.">$plain[$name]</weak_warning> = true;
    };
    $render = function () use ($template, $registry) {
        require __DIR__ . '/view.php';
    };
    $renderRef = function () use (&$template) {
        include_once 'view.php';
    };
    $arrow = function () use ($template) {
        $f = fn () => include 'x.php';
        return $f;
    };
    $noop = function () use (<weak_warning descr="Variable is never used.">$template</weak_warning>) {
        $inner = function () { include 'x.php'; };
    };
    $noopRef = function () use (&<weak_warning descr="Variable is never used.">$template</weak_warning>) {
        $inner = new class { function m() { require 'x.php'; } };
    };
    return [$register, $render, $renderRef, $arrow, $noop, $noopRef];
}

/** @param \Countable|null $maybe */
function objectsByInference($maybe, $untyped, Closure $cb)
{
    $bag = $maybe ?? new ArrayObject();
    $list = [];
    $mixed = $untyped;
    $fn = $cb;
    return [
        function () use ($bag) { $bag['k'] = 1; },
        function () use ($maybe) { $maybe[] = 1; },
        function () use ($list) { <weak_warning descr="Value is only written here and never read; the write is lost.">$list[]</weak_warning> = 1; },
        function () use ($mixed) { <weak_warning descr="Value is only written here and never read; the write is lost.">$mixed[]</weak_warning> = 1; },
        function () use ($fn) { <weak_warning descr="Value is only written here and never read; the write is lost.">$fn</weak_warning> .= 'x'; },
        function () use (<weak_warning descr="Variable is never used.">$bag</weak_warning>) { return 1; },
    ];
}
